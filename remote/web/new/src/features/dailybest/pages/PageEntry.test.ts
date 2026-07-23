import assert from "node:assert/strict";
import { afterEach, test, vi } from "vitest";
import { createSSRApp, type Component } from "vue";
import { createMemoryHistory, createRouter } from "vue-router";
import { renderToString } from "@vue/server-renderer";
import FetchRunDetailPage from "./FetchRunDetailPage.vue";
import FetchRunsPage from "./FetchRunsPage.vue";
import OverviewPage from "./OverviewPage.vue";
import RankDayStatsPage from "./RankDayStatsPage.vue";
import RanksPage from "./RanksPage.vue";

afterEach(() => vi.unstubAllGlobals());

const pages: Array<{ path: string; route: string; title: string; component: Component }> = [
  { path: "/dailybestoverview", route: "/dailybestoverview", title: "运行概览", component: OverviewPage },
  { path: "/dailybestfetchrunlist", route: "/dailybestfetchrunlist", title: "抓取任务", component: FetchRunsPage },
  {
    path: "/dailybestfetchrundetail/:id",
    route: "/dailybestfetchrundetail/10",
    title: "抓取任务 #10",
    component: FetchRunDetailPage
  },
  { path: "/ranklist", route: "/ranklist?day_number=1", title: "排名明细", component: RanksPage },
  { path: "/rankdaystats", route: "/rankdaystats", title: "每日统计", component: RankDayStatsPage }
];

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
