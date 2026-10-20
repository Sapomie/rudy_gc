import assert from "node:assert/strict";
import { afterEach, test, vi } from "vitest";
import { createSSRApp, type Component } from "vue";
import { createMemoryHistory, createRouter } from "vue-router";
import { renderToString } from "@vue/server-renderer";
import FetchLogsPage from "./FetchLogsPage.vue";
import FetchRunDetailPage from "./FetchRunDetailPage.vue";
import FetchRunsPage from "./FetchRunsPage.vue";
import RankDayStatsPage from "./RankDayStatsPage.vue";
import TargetProbesPage from "./TargetProbesPage.vue";

afterEach(() => vi.unstubAllGlobals());

const pages: Array<{ path: string; route: string; title: string; component: Component }> = [
  {
    path: "/remotedailybestfetchrundetail/:id",
    route: "/remotedailybestfetchrundetail/10",
    title: "抓取任务 #10",
    component: FetchRunDetailPage
  }
];

const listPages: Array<{
  path: string;
  title: string;
  component: Component;
  responseKey: string;
  requestPath: string;
  localRequestPath?: string;
  actionText?: string;
}> = [
  {
    path: "/remotedailybestfetchrunlist",
    title: "抓取任务",
    component: FetchRunsPage,
    responseKey: "runs",
    requestPath: "/remote-api/api/v1/dailybest/fetch-runs?p=1&ps=24"
  },
  {
    path: "/remotedailybestrankdaystats",
    title: "每日统计",
    component: RankDayStatsPage,
    responseKey: "stats",
    requestPath: "/remote-api/api/v1/dailybest/rank-day-stats?p=1&ps=24",
    localRequestPath: "/api/dailybest/rank-days?p=1&ps=24"
  },
  {
    path: "/remotedailybestfetchloglist",
    title: "请求日志",
    component: FetchLogsPage,
    responseKey: "logs",
    requestPath: "/remote-api/api/v1/dailybest/fetch-logs?p=1&ps=24"
  },
  {
    path: "/remotedailybesttargetprobelist",
    title: "目标站探测",
    component: TargetProbesPage,
    responseKey: "probes",
    requestPath: "/remote-api/api/v1/dailybest/target-probes?p=1&ps=24",
    actionText: "立即探测"
  }
];

for (const page of listPages) {
  test(`${page.title}页面首次进入立即请求列表`, async () => {
    const fetchMock = vi.fn().mockImplementation((input: string) => {
      if (input === page.localRequestPath) {
        return Promise.resolve(
          new Response(
            JSON.stringify({
              Days: [],
              Total: 0,
              Query: { BusinessDateStart: "", BusinessDateEnd: "", Page: 1, PageSize: 24 },
              PageInfo: { Total: 0, Page: 1, PageSize: 24, PageTotal: 1 }
            }),
            { status: 200 }
          )
        );
      }
      return Promise.resolve(
        new Response(
          JSON.stringify({
            [page.responseKey]: [],
            page_info: { total: 0, page: 1, page_size: 24, page_total: 1 }
          }),
          { status: 200 }
        )
      );
    });
    vi.stubGlobal("fetch", fetchMock);
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: page.path, component: page.component },
        { path: "/:pathMatch(.*)*", component: { template: "<div />" } }
      ]
    });
    await router.push(page.path);
    await router.isReady();

    const app = createSSRApp(page.component);
    app.use(router);
    const html = await renderToString(app);

    assert.match(html, new RegExp(page.title));
    if (page.actionText) assert.match(html, new RegExp(page.actionText));
    assert.doesNotMatch(html, /尚未查询/);
    const expectedPaths = [page.requestPath, ...(page.localRequestPath ? [page.localRequestPath] : [])];
    assert.equal(fetchMock.mock.calls.length, expectedPaths.length);
    for (const expectedPath of expectedPaths) {
      assert.ok(
        fetchMock.mock.calls.some(([path]) => path === expectedPath),
        `缺少请求：${expectedPath}`
      );
    }
  });
}

for (const page of pages) {
  test(`${page.title}页面首次进入只展示结构且不请求接口`, async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: page.path, component: page.component },
        { path: "/:pathMatch(.*)*", component: { template: "<div />" } }
      ]
    });
    await router.push(page.route);
    await router.isReady();

    const app = createSSRApp(page.component);
    app.use(router);
    const html = await renderToString(app);

    assert.match(html, new RegExp(page.title));
    assert.match(html, /尚未查询/);
    assert.equal(fetchMock.mock.calls.length, 0);
  });
}
