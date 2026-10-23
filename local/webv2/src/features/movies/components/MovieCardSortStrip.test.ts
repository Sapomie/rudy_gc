import assert from "node:assert/strict";
import { test } from "vitest";
import { createSSRApp } from "vue";
import { renderToString } from "@vue/server-renderer";
import MovieCardSortStrip from "./MovieCardSortStrip.vue";

test("卡片排序条显示当前方向并保留切换链接", async () => {
  const app = createSSRApp(MovieCardSortStrip, {
    options: [
      { key: "ByReleasing", label: "上映日期" },
      { key: "ByHighestRank", label: "最高排名" }
    ],
    query: {
      ByReleasing: {
        Href: "/listcard?od=rd&order=asc&p=1",
        Active: true,
        Desc: true
      },
      ByHighestRank: {
        Href: "/listcard?od=hrk&order=asc&p=1",
        Active: false,
        Desc: false
      }
    }
  });

  const html = await renderToString(app);

  assert.match(html, /class="sort-link active desc strip-link"/);
  assert.match(html, /上映日期/);
  assert.match(html, /↓/);
  assert.match(html, /order=asc/);
  assert.doesNotMatch(html, /最高排名[\s\S]*↑/);
});

test("卡片排序条显示正序并链接到倒序", async () => {
  const app = createSSRApp(MovieCardSortStrip, {
    options: [{ key: "ByLastSelectedScTime", label: "最近入选 SC" }],
    query: {
      ByLastSelectedScTime: {
        Href: "/listcard?od=lsst&order=desc&p=1",
        Active: true,
        Desc: false
      }
    }
  });

  const html = await renderToString(app);

  assert.match(html, /class="sort-link active strip-link"/);
  assert.match(html, /最近入选 SC/);
  assert.match(html, /↑/);
  assert.match(html, /order=desc/);
  assert.doesNotMatch(html, /class="sort-link active desc strip-link"/);
});
