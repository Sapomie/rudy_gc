import assert from "node:assert/strict";
import { afterEach, test, vi } from "vitest";
import { createSSRApp } from "vue";
import { createMemoryHistory, createRouter } from "vue-router";
import { renderToString } from "@vue/server-renderer";
import TargetProbesPage from "./TargetProbesPage.vue";

afterEach(() => vi.unstubAllGlobals());

test("目标站探测页面首次进入只展示筛选结构且不请求接口", async () => {
  const fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/dailybesttargetprobelist", component: TargetProbesPage },
      { path: "/dailybestruntimestatus", component: { template: "<div />" } }
    ]
  });
  await router.push("/dailybesttargetprobelist?status=failed");
  await router.isReady();
  const app = createSSRApp(TargetProbesPage);
  app.use(router);
  const html = await renderToString(app);
  assert.match(html, /目标站探测/);
  assert.match(html, /立即探测/);
  assert.match(html, /尚未查询/);
  assert.match(html, /value="failed" selected/);
  assert.equal(fetchMock.mock.calls.length, 0);
});
