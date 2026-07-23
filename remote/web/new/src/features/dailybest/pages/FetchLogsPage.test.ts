import assert from "node:assert/strict";
import { afterEach, test, vi } from "vitest";
import { createSSRApp } from "vue";
import { createMemoryHistory, createRouter } from "vue-router";
import { renderToString } from "@vue/server-renderer";
import FetchLogsPage from "./FetchLogsPage.vue";

afterEach(() => vi.unstubAllGlobals());

test("请求日志页面首次进入只展示筛选条件且不请求接口", async () => {
  const fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/dailybestfetchloglist", component: FetchLogsPage },
      { path: "/dailybestfetchrunlist", component: { template: "<div />" } },
      { path: "/dailybestfetchrundetail/:id", component: { template: "<div />" } }
    ]
  });
  await router.push("/dailybestfetchloglist?fetch_run_id=7");
  await router.isReady();

  const app = createSSRApp(FetchLogsPage);
  app.use(router);
  const html = await renderToString(app);

  assert.match(html, /请求日志/);
  assert.match(html, /尚未查询/);
  assert.match(html, /value="7"/);
  assert.equal(fetchMock.mock.calls.length, 0);
});
