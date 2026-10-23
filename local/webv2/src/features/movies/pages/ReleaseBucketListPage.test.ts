import assert from "node:assert/strict";
import { createSSRApp } from "vue";
import { renderToString } from "@vue/server-renderer";
import { test, vi } from "vitest";
import ReleaseBucketListPage from "./ReleaseBucketListPage.vue";

const state = vi.hoisted(() => ({
  route: {
    path: "/bucketsbirth",
    fullPath: "/bucketsbirth",
    params: {},
    meta: { apiPath: "/w-media-agg/bucket-list" }
  },
  getJSON: vi.fn()
}));

vi.mock("vue-router", async (importOriginal) => ({
  ...(await importOriginal<typeof import("vue-router")>()),
  useRoute: () => state.route
}));

vi.mock("@/features/movies/api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/features/movies/api")>()),
  getJSON: state.getJSON
}));

async function render(payload: Record<string, unknown>, wmedia = true): Promise<string> {
  state.route.path = wmedia ? "/bucketsbirth" : "/bucketsrelease";
  state.route.fullPath = state.route.path;
  state.route.meta.apiPath = wmedia ? "/w-media-agg/bucket-list" : "/movie-agg-all/release-bucket-list";
  state.getJSON.mockResolvedValue(payload);
  const app = createSSRApp(ReleaseBucketListPage);
  app.mixin({
    async serverPrefetch() {
      await Promise.resolve();
    }
  });
  return renderToString(app);
}

function cells(rowHTML: string): string[] {
  return Array.from(rowHTML.matchAll(/<td\b[^>]*>([\s\S]*?)<\/td>/g), (match) =>
    match[1].replace(/<[^>]*>/g, "").trim()
  );
}

test("时间桶移除旧聚合页面跳转，保留统计和回填入口", async () => {
  for (const birth of [true, false]) {
    const oldPath = birth ? "/aggbirth" : "/aggrelease";
    const html = await render(
      {
        AggPageHref: oldPath,
        Rows: [
          {
            BucketDisplay: "2026-M08",
            ViewHref: `${oldPath}?year=2026&month=8`,
            CountAll: 12,
            CountOwned: 8,
            AddedCount: 2
          }
        ]
      },
      birth
    );
    assert.match(html, /2026-M08/);
    assert.match(html, /href="\/aggbackfill"/);
    assert.doesNotMatch(html, /href="\/aggbirth|href="\/aggrelease|返回聚合页/);
  }
});

test("下载时间桶按顺序展示五项统计并保留负数、零及正数大小", async () => {
  const gb = 1024 ** 3;
  const headerLabels = {
    Bucket: "时间桶",
    Added: "新增数目",
    AddedSize: "新增大小",
    Removed: "删除数目",
    RemovedSize: "删除大小",
    NetSize: "净增加大小",
    Subtitle: "字幕数",
    LatestBirth: "最新下载时间"
  };
  const html = await render({
    ViewType: "w-media-bucket-list",
    HeaderLinks: Object.fromEntries(
      Object.entries(headerLabels).map(([key, Label]) => [
        key,
        {
          Label,
          Href: `/bucketsbirth?sort=${key === "NetSize" ? "net_size" : key}&dir=asc`,
          Active: key === "NetSize",
          Desc: true
        }
      ])
    ),
    Rows: [
      {
        BucketDisplay: "2026-M08",
        AddedCount: 3,
        AddedSizeBytes: 3 * gb,
        RemovedCount: 5,
        RemovedSizeBytes: 5 * gb,
        NetSizeBytes: -2 * gb,
        HasSubCount: 1
      },
      {
        BucketDisplay: "2026-M07",
        AddedCount: 0,
        AddedSizeBytes: 0,
        RemovedCount: 1,
        RemovedSizeBytes: 512,
        NetSizeBytes: -512,
        HasSubCount: 0
      },
      {
        BucketDisplay: "2026-M06",
        AddedCount: 1,
        AddedSizeBytes: gb,
        RemovedCount: 1,
        RemovedSizeBytes: gb,
        NetSizeBytes: 0,
        HasSubCount: 0
      },
      {
        BucketDisplay: "2026-M05",
        AddedCount: 1,
        AddedSizeBytes: gb,
        RemovedCount: 0,
        RemovedSizeBytes: 0,
        NetSizeBytes: gb,
        HasSubCount: 0
      }
    ]
  });
  const heading = html.match(/<thead>([\s\S]*?)<\/thead>/)?.[1] || "";
  const labels = Array.from(heading.matchAll(/<th\b[^>]*>([\s\S]*?)<\/th>/g), (match) =>
    match[1]
      .replace(/<[^>]*>/g, "")
      .replace("↓", "")
      .trim()
  );
  assert.deepEqual(labels, Object.values(headerLabels));
  const body = html.match(/<tbody>([\s\S]*?)<\/tbody>/)?.[1] || "";
  const rows = Array.from(body.matchAll(/<tr\b[^>]*>([\s\S]*?)<\/tr>/g), (match) => cells(match[1]));
  assert.deepEqual(rows[0].slice(0, 7), ["2026-M08", "3", "3.0 GB", "5", "5.0 GB", "-2.0 GB", "1"]);
  assert.equal(rows[1][5], "-512 B");
  assert.equal(rows[2][5], "0 B");
  assert.equal(rows[3][5], "1.0 GB");
  assert.equal(rows[0].length, 8);
  assert.match(heading, /href="\/bucketsbirth\?sort=net_size&amp;dir=asc"/);
  assert.equal(state.getJSON.mock.lastCall?.[0], "/api/w-media-agg/bucket-list");
});

test("下载时间桶空状态跨越全部八列", async () => {
  const html = await render({ ViewType: "w-media-bucket-list", Rows: [] });
  assert.match(html, /colspan="8"/);
  assert.match(html, /当前筛选下没有/);
});

test("上映时间桶继续展示原有六列及统计值", async () => {
  const html = await render(
    {
      ViewType: "movie-release-bucket-list",
      Rows: [{ BucketDisplay: "2026-M08", CountAll: 12, CountOwned: 8, RemovedCount: 2, SizeBytes: 1024 ** 3 }]
    },
    false
  );
  const body = html.match(/<tbody>([\s\S]*?)<\/tbody>/)?.[1] || "";
  const row = cells(body);
  assert.equal(row.length, 6);
  assert.deepEqual(row.slice(0, 5), ["2026-M08", "12", "8", "2", "1.0 GB"]);
});
