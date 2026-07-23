import assert from "node:assert/strict";
import { test } from "vitest";
import { createSSRApp } from "vue";
import { createMemoryHistory, createRouter } from "vue-router";
import { renderToString } from "@vue/server-renderer";
import FetchPage from "./FetchPage.vue";

test("手动抓取页面只展示抓取动作", async () => {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/dailybestfetch", component: FetchPage },
      { path: "/dailybestfetchrunlist", component: { template: "<div />" } }
    ]
  });
  await router.push("/dailybestfetch");
  await router.isReady();

  const app = createSSRApp(FetchPage);
  app.use(router);
  const html = await renderToString(app);

  assert.match(html, /创建抓取任务/);
  assert.doesNotMatch(html, /<input|<select|category=1/);
});
