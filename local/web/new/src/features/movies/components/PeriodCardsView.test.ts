import assert from "node:assert/strict";
import { createSSRApp } from "vue";
import { renderToString } from "@vue/server-renderer";
import { test, vi } from "vitest";
import CardsPeriodReleasePage from "../pages/CardsPeriodReleasePage.vue";
import CardsPeriodBirthPage from "../pages/CardsPeriodBirthPage.vue";
import type { CardPeriodBasis, CardPeriodSort } from "../utils/periodCards";
import { movieCardSortOptions } from "../utils/movieCardSortOptions";

const state = vi.hoisted(() => ({
  route: {
    path: "/listcardperiodrelease",
    fullPath: "/listcardperiodrelease",
    params: {},
    query: {} as Record<string, string>,
    meta: { apiPath: "/cardsperiodrelease" }
  },
  getJSON: vi.fn(),
  push: vi.fn()
}));
vi.mock("vue-router", async (importOriginal) => ({
  ...(await importOriginal<typeof import("vue-router")>()),
  useRoute: () => state.route,
  useRouter: () => ({ push: state.push })
}));
vi.mock("@/features/movies/api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/features/movies/api")>()),
  getJSON: state.getJSON
}));
vi.mock("@/features/movies/components/MovieGrid.vue", () => ({
  default: { props: ["cards"], template: '<div class="test-movie-grid">{{ cards.length }} 张卡片</div>' }
}));

async function render(
  basis: CardPeriodBasis = "release",
  periods = true,
  ascending = false,
  countOwned = 3,
  periodSort: CardPeriodSort = "time",
  periodOrder = "desc",
  mediaOwned = 1
) {
  state.route.query = {};
  const period = {
    Key: "month:2020-12",
    Label: "2020-12",
    RangeStart: "2020-12-01",
    RangeEnd: "2020-12-31",
    Count: 7,
    CountOwned: countOwned
  };
  const ownedQuery = Object.fromEntries(
    [["MediaAll", 1], ["MediaOwned", 3], ["MediaNotOwned", 7]].map(([key, value]) => [key, {
      Href: `/listcardperiodrelease?period_type=month&key=month%3A2020-12&period_p=3&period_ps=24&period_sort=${periodSort}&period_order=${periodOrder}&p=1&ps=18&od=vw&order=asc&mowned=${value}`,
      Active: value === mediaOwned
    }])
  );
  state.getJSON.mockResolvedValue({
    Basis: basis,
    PeriodType: "month",
    Periods: periods ? [period] : [],
    PeriodTotal: periods ? 1 : 0,
    SelectedPeriod: periods ? period : null,
    movies: [],
    total: 0,
    PageInfo: { PageSize: 18, Links: [{ Label: "2", Href: `/listcardperiod${basis}?key=month%3A2020-12&p=2&period_p=3` }] },
    PeriodPageInfo: { PageSize: 24, Links: [{ Label: "4", Href: `/listcardperiod${basis}?period_p=4&p=1` }] },
    ownedQuery: basis === "release" ? ownedQuery : undefined,
    Query: {
      period_type: ["month"],
      key: [period.Key],
      period_p: ["3"],
      p: ["1"],
      period_sort: [periodSort],
      period_order: [periodOrder]
    },
    sortQuery: Object.fromEntries(
      movieCardSortOptions.map(({ key }, index) => [
        key,
        {
          Href: `/listcardperiod${basis}?period_type=month&key=month%3A2020-12&period_p=3&p=1&od=${["du", "rd", "vw", "rk", "drk", "hrk", "bt", "lsst", "sc", "co"][index]}&order=${ascending ? "desc" : "asc"}`,
          Active: key === "ByViewer",
          Desc: key === "ByViewer" && !ascending
        }
      ])
    )
  });
  return renderPage(basis);
}

async function renderPage(basis: CardPeriodBasis = "release") {
  state.route.path = `/listcardperiod${basis}`;
  state.route.fullPath = state.route.path;
  state.route.meta.apiPath = `/cardsperiod${basis}`;
  const app = createSSRApp(basis === "release" ? CardsPeriodReleasePage : CardsPeriodBirthPage);
  app.mixin({
    async serverPrefetch() {
      await Promise.resolve();
    }
  });
  return renderToString(app);
}

test("左侧仅展示四种周期及选中行，右侧复用卡片网格和独立分页", async () => {
  const html = await render();
  assert.equal(state.getJSON.mock.lastCall?.[0], "/api/cardsperiodrelease");
  assert.match(html, /上映周期卡片/);
  assert.match(html, /按上映日期浏览影片/);
  assert.doesNotMatch(html, /下载周期卡片|basis=/);
  assert.match(html, /cards-period-directory/);
  assert.match(html, /class="is-selected"/);
  assert.match(html, /2020-12-01 至 2020-12-31/);
  assert.match(html, /影片数/);
  for (const kind of ["year", "quarter", "month", "day"]) assert.match(html, new RegExp(`period_type=${kind}`));
  assert.doesNotMatch(html, /period_type=week/);
  assert.doesNotMatch(html, />\s*周\s*</);
  assert.match(html, /aria-label="每页周期数"/);
  assert.match(html, /aria-label="每页卡片数"/);
  assert.match(html, /key=month%3A2020-12&amp;p=2&amp;period_p=3/);
  assert.match(html, /href="\/listcardperiodrelease\?period_p=4"/);
  assert.match(html, /aria-label="每行卡片数"/);
  assert.match(html, /test-movie-grid/);
});

test("下载周期标明媒体数，空目录保留可用的切换控件", async () => {
  assert.match(await render("birth"), /媒体数/);
  assert.equal(state.getJSON.mock.lastCall?.[0], "/api/cardsperiodbirth");
  const empty = await render("release", false);
  assert.match(empty, /暂无周期数据/);
  assert.match(empty, /请选择周期/);
  assert.match(empty, /上映周期/);
  assert.doesNotMatch(empty, /下载周期|basis=/);
  assert.doesNotMatch(empty, /cards-period-directory-pager/);
});

test("上映目录显示已拥有统计及零值，下载目录保持两列，空态跨列正确", async () => {
  const release = (await render()).replace(/<!--[\s\S]*?-->/g, "");
  assert.match(release, /<th class="sc-number"><a[^>]*>已拥有<\/a><\/th>/);
  assert.match(release, /<td class="sc-number">3<\/td>/);
  const zero = await render("release", true, false, 0);
  assert.match(zero, /<td class="sc-number">0<\/td>/);
  const birth = await render("birth");
  assert.doesNotMatch(birth, /已拥有/);
  assert.doesNotMatch(birth, /is-release-table/);
  assert.match(await render("release", false), /colspan="3"/);
  assert.match(await render("birth", false), /colspan="2"/);
});

test("已拥有表头复用排序控件，显示选中方向并提供反向排序链接", async () => {
  for (const [sort, order, nextOrder, indicator] of [
    ["time", "desc", "desc", ""],
    ["owned", "desc", "asc", "↓"],
    ["owned", "asc", "desc", "↑"]
  ] as const) {
    const html = (await render("release", true, false, 3, sort, order)).replace(/<!--[\s\S]*?-->/g, "");
    const link = html.match(/<a\b([^>]*)>已拥有([\s\S]*?)<\/a>/);
    assert.ok(link, "已拥有表头缺少排序链接");
    const href = link[1].match(/href="([^"]*)"/)?.[1];
    assert.ok(href);
    const url = new URL(href.replace(/&amp;/g, "&"), "http://localhost");
    assert.equal(url.pathname, "/listcardperiodrelease");
    assert.equal(url.searchParams.get("period_sort"), "owned");
    assert.equal(url.searchParams.get("period_order"), nextOrder);
    assert.equal(url.searchParams.get("period_p"), "1");
    // 共享链接规范省略默认卡片页码 p=1。
    assert.equal(url.searchParams.get("p"), null);
    assert.equal(url.searchParams.has("key"), false);
    if (indicator) {
      assert.match(link[1], /class="sort-link active/);
      assert.match(link[1], /aria-current="page"/);
      assert.ok(link[2].includes(indicator));
    } else {
      assert.doesNotMatch(link[1], /aria-current/);
      assert.equal(link[2], "");
    }
  }
  assert.doesNotMatch(await render("birth"), /period_sort=owned/);
});

test("上映卡片复用今日卡片的三段拥有筛选，切换只重置卡片分页", async () => {
  for (const mediaOwned of [1, 3, 7]) {
    const html = (await render("release", true, false, 3, "owned", "asc", mediaOwned)).replace(/<!--[\s\S]*?-->/g, "");
    const strip = html.match(/<div class="owned-line cards-toolbar-filters"><div class="link-strip">([\s\S]*?)<\/div>/)?.[1];
    assert.ok(strip, "缺少今日卡片同款分段筛选");
    const links = [...strip.matchAll(/<a href="([^"]*)" class="([^"]*)">\s*(All|Owned|NotOwned)\s*<\/a>/g)];
    assert.equal(links.length, 3);
    for (const [index, link] of links.entries()) {
      const value = [1, 3, 7][index];
      const url = new URL(link[1].replace(/&amp;/g, "&"), "http://localhost");
      assert.equal(link[3], ["All", "Owned", "NotOwned"][index]);
      assert.equal(link[2].includes("active"), value === mediaOwned);
      assert.equal(url.pathname, "/listcardperiodrelease");
      assert.equal(url.searchParams.get("mowned"), String(value));
      assert.equal(url.searchParams.get("p"), "1");
      assert.equal(url.searchParams.get("key"), "month:2020-12");
      assert.equal(url.searchParams.get("period_p"), "3");
      assert.equal(url.searchParams.get("period_ps"), "24");
      assert.equal(url.searchParams.get("period_sort"), "owned");
      assert.equal(url.searchParams.get("period_order"), "asc");
      assert.equal(url.searchParams.get("od"), "vw");
      assert.equal(url.searchParams.get("order"), "asc");
      assert.equal(url.searchParams.get("ps"), "18");
    }
  }
  assert.doesNotMatch(await render("birth"), /owned-line cards-toolbar-filters|NotOwned/);
});

test("上映和下载周期均展示今日卡片的全部排序及当前方向", async () => {
  const labels = [
    "Detail更新",
    "上映日期",
    "观看人数",
    "上榜日期",
    "在榜天数",
    "最高排名",
    "下载时间",
    "最近入选 SC",
    "sc次数",
    "co次数"
  ];
  for (const basis of ["release", "birth"] as const) {
    const html = await render(basis);
    for (const label of labels) assert.ok(html.includes(label), `${basis} 缺少 ${label}`);
    assert.match(html, /sort-link active desc strip-link/);
    assert.match(html, /观看人数[\s\S]*?↓/);
    assert.match(html, /key=month%3A2020-12&amp;period_p=3&amp;od=vw&amp;order=asc/);
  }
  const ascending = await render("release", true, true);
  assert.match(ascending, /sort-link active strip-link/);
  assert.match(ascending, /观看人数[\s\S]*?↑/);
  assert.doesNotMatch(ascending, /sort-link active desc strip-link/);
});

test("接口失败时保留左右布局，URL 参数无法改变页面依据", async () => {
  for (const basis of ["release", "birth"] as const) {
    state.route.query = { basis: basis === "release" ? "birth" : "release", period_type: "month", period_p: "3" };
    state.getJSON.mockRejectedValue(new Error("周期查询失败"));
    const html = await renderPage(basis);
    assert.match(html, /cards-period-directory/);
    assert.match(html, /cards-period-results/);
    assert.match(html, /周期查询失败/);
    assert.match(html, /加载失败/);
    assert.match(html, new RegExp(`href="/listcardperiod${basis}\\?period_type=year`));
    assert.match(html, basis === "release" ? /<small>上映周期<\/small>/ : /<small>下载周期<\/small>/);
    assert.doesNotMatch(html, /basis=|test-movie-grid/);
  }
});

test("初次加载保留目录与周期类型，暂时禁用每页数量设置", async () => {
  state.route.query = {};
  state.getJSON.mockReturnValue(new Promise(() => {}));
  const html = await renderPage("birth");
  assert.match(html, /加载周期中/);
  assert.match(html, /loading-state/);
  assert.match(html, /aria-busy="true"/);
  assert.match(html, /aria-label="每页周期数"[^>]*disabled/);
  assert.match(html, /收起时间周期/);
  assert.match(html, /<small>下载周期<\/small>/);
  assert.doesNotMatch(html, /test-movie-grid/);

  state.route.query = { basis: "birth" };
  const release = await renderPage("release");
  assert.match(release, /<small>上映周期<\/small>/);
});
