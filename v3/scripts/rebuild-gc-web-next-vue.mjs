#!/usr/bin/env node
import {mkdir, writeFile} from "node:fs/promises";
import path from "node:path";

const root = path.resolve(path.dirname(new URL(import.meta.url).pathname), "..");

const versions = [
  {
    id: "02",
    sourceVersion: "v2",
    port: 5022,
    styleSource: "/Users/gaojinwei/src/my/2026/tdu_cinema_template/gc-web-v2",
    theme: {
      bg: "#f9f8fb",
      bgLayer: "#f9f8fb",
      panel: "#ffffff",
      panel2: "#f8fafc",
      ink: "#344767",
      muted: "#64748b",
      line: "#d6dee9",
      accent: "#7a6fbe",
      accent2: "#58db83",
      danger: "#ec536c",
      sidebar: "#2b3a4a",
      sidebarText: "#a8b6c6",
      topbar: "#ffffff",
      radius: "12px",
      buttonRadius: "4px",
      shadow: "0 4px 12px rgba(30, 41, 59, 0.08)",
      cardShadow: "0 1px 2px rgba(15, 23, 42, 0.04)",
      font: "Inter, \"PingFang SC\", \"Microsoft YaHei\", sans-serif",
      displayFont: "Inter, \"PingFang SC\", \"Microsoft YaHei\", sans-serif",
      borderWidth: "1px",
      posterRadius: "8px",
      topbarShadow: "0 2px 6px rgba(52, 71, 103, 0.08)",
    },
  },
  {
    id: "03",
    sourceVersion: "v3",
    port: 5023,
    styleSource: "/Users/gaojinwei/src/my/2026/tdu_cinema_template/gc-web-v3",
    theme: {
      bg: "#10161a",
      bgLayer: "linear-gradient(180deg, rgba(64, 188, 244, 0.08), transparent 28rem), #10161a",
      panel: "#141f23",
      panel2: "#1c292e",
      ink: "#dbe8dd",
      muted: "#89a19a",
      line: "#2b4147",
      accent: "#00c030",
      accent2: "#40bcf4",
      danger: "#ff5f56",
      sidebar: "#0b1114",
      sidebarText: "#89a19a",
      topbar: "#141f23",
      radius: "7px",
      buttonRadius: "5px",
      shadow: "0 20px 48px rgba(0, 0, 0, 0.38)",
      cardShadow: "0 20px 48px rgba(0, 0, 0, 0.38)",
      font: "Inter, \"PingFang SC\", \"Microsoft YaHei\", sans-serif",
      displayFont: "Inter, \"PingFang SC\", \"Microsoft YaHei\", sans-serif",
      borderWidth: "1px",
      posterRadius: "5px",
      topbarShadow: "0 18px 40px rgba(0, 0, 0, 0.26)",
    },
  },
  {
    id: "04",
    sourceVersion: "v4",
    port: 5024,
    styleSource: "/Users/gaojinwei/src/my/2026/tdu_cinema_template/gc-web-v4",
    theme: {
      bg: "#f4f4ef",
      bgLayer: "#f4f4ef",
      panel: "#ffffff",
      panel2: "#f4f4ef",
      ink: "#101010",
      muted: "#696966",
      line: "#171717",
      accent: "#d71920",
      accent2: "#8d1116",
      danger: "#d71920",
      sidebar: "#ffffff",
      sidebarText: "#101010",
      topbar: "#f4f4ef",
      radius: "0",
      buttonRadius: "0",
      shadow: "0 20px 46px rgba(0, 0, 0, 0.16)",
      cardShadow: "0 20px 46px rgba(0, 0, 0, 0.16)",
      font: "Inter, \"PingFang SC\", \"Microsoft YaHei\", sans-serif",
      displayFont: "Georgia, \"Times New Roman\", \"Songti SC\", serif",
      borderWidth: "1px",
      posterRadius: "0",
      topbarShadow: "none",
    },
  },
  {
    id: "05",
    sourceVersion: "v5",
    port: 5025,
    styleSource: "/Users/gaojinwei/src/my/2026/tdu_cinema_template/gc-web-v5",
    theme: {
      bg: "#090c12",
      bgLayer: "radial-gradient(circle at 20% 0%, #1b2942 0, #090c12 38%)",
      panel: "#111827",
      panel2: "#192132",
      ink: "#f4f7fb",
      muted: "#93a1b5",
      line: "rgba(255, 255, 255, 0.12)",
      accent: "#f0c05a",
      accent2: "#54d2d2",
      danger: "#ff6f91",
      sidebar: "#111827",
      sidebarText: "#dbe3ee",
      topbar: "#111827",
      radius: "8px",
      buttonRadius: "999px",
      shadow: "0 24px 60px rgba(0, 0, 0, 0.38)",
      cardShadow: "0 24px 60px rgba(0, 0, 0, 0.38)",
      font: "Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif",
      displayFont: "Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif",
      borderWidth: "1px",
      posterRadius: "8px",
      topbarShadow: "0 22px 54px rgba(0, 0, 0, 0.28)",
    },
  },
];

const navGroups = [
  {
    label: "MovieCard",
    items: [
      ["/cards", "All"],
      ["/cardstoday", "Today"],
      ["/cardsmediamowned", "MediaOwned"],
      ["/cardshasrank", "HasRank"],
      ["/moviecarddayrank", "DayRank"],
      ["/moviecardperiodrank", "PeriodRank"],
      ["/cardsneeddownload", "NeedDownload"],
      ["/cardsrandom", "Random"],
      ["/cardsagg", "Agg"],
    ],
  },
  {
    label: "MovieAgg",
    items: [
      ["/movie-agg-all/release", "Agg-All-Release"],
      ["/movie-agg-all/release-bucket-list", "Release-Bucket"],
      ["/w-media-agg/bucket-list", "WMedia-Bucket"],
      ["/w-media-agg/birth", "WMedia-Birth"],
      ["/triggers/agg", "AggTriggers"],
      ["/wdir", "WDirectory"],
    ],
  },
  {
    label: "List",
    items: [
      ["/medias", "Medias"],
      ["/crawl-records", "CrawlRecords"],
      ["/w-agg-events", "WAggEvents"],
      ["/torrent-albums", "TorrentAlbums"],
      ["/movie-albums", "MovieAlbums"],
      ["/d-seeds", "DSeeds"],
      ["/casts", "Casts"],
      ["/e-items", "EItems"],
    ],
  },
  {
    label: "SC",
    items: [
      ["/sc-events", "ScEvents"],
      ["/sc-events-cards", "ScEventCards"],
      ["/sc-pick-smart-media", "SmartMedia"],
      ["/triggers/sc", "ScPreview"],
      ["/triggers/sc-media", "ScMediaMove"],
    ],
  },
  {
    label: "抓取任务",
    items: [
      ["/triggers/dailybest", "DailyBest"],
      ["/triggers/seeds", "Seeds"],
      ["/triggers/post-process", "PostProcess"],
      ["/triggers/backfill", "Backfill"],
      ["/crawler/detail-loop", "DetailLoop"],
      ["/crawler/tasks", "CrawlerTasks"],
    ],
  },
  {
    label: "FetchSite操作",
    items: [
      ["/triggers/fetch-site", "FetchSite"],
      ["/triggers/fetch-sehuatang", "FetchSehuatang"],
      ["/fetch-sites", "FetchSitesConfig"],
      ["/fetch-site-javbus-list", "JavbusList"],
      ["/fetch-site-sukebei-list", "SukebeiList"],
      ["/fetch-site-sehuatang-list", "SehuatangList"],
    ],
  },
  {
    label: "色花堂SH",
    items: [
      ["/triggers/fetch-sehuatang-sh", "SH一体抓取"],
      ["/triggers/fetch-sehuatang-sh-thread-list", "SH列表建池"],
      ["/triggers/fetch-sehuatang-sh-thread-detail", "SH详情补全"],
      ["/fetch-site-sehuatang-sh-list", "SH帖子池"],
    ],
  },
  {
    label: "Media操作",
    items: [
      ["/triggers/media", "Media"],
      ["/triggers/media-rescan", "MediaRescan"],
      ["/triggers/media-rollback", "MediaRollback"],
    ],
  },
];

const topGroups = [
  {
    label: "常用",
    items: [
      ["/crawler/tasks", "任务列表"],
      ["/movie-albums", "MovieAlbums"],
      ["/torrent-albums", "TorrentAlbums"],
      ["/triggers/dailybest", "DailyBest"],
      ["/triggers/fetch-sehuatang", "FetchSehuatang"],
    ],
  },
  {
    label: "色花堂SH",
    items: [
      ["/triggers/fetch-sehuatang-sh", "SH一体抓取"],
      ["/triggers/fetch-sehuatang-sh-thread-list", "SH列表建池"],
      ["/triggers/fetch-sehuatang-sh-thread-detail", "SH详情补全"],
      ["/fetch-site-sehuatang-sh-list", "SH帖子池"],
    ],
  },
];

const cardsPages = [
  ["/cards", "cards", "Movie Cards"],
  ["/cardstoday", "cardstoday", "Today"],
  ["/cardshasrank", "cardshasrank", "Has Rank"],
  ["/cardsmediamowned", "cardsmediamowned", "Media Owned"],
  ["/cardsneeddownload", "cardsneeddownload", "Need Download"],
  ["/cardsrandom", "cardsrandom", "Random"],
];

const listPages = [
  ["/medias", "medias", "Medias"],
  ["/crawl-records", "crawl-records", "Crawl Records"],
  ["/w-agg-events", "w-agg-events", "W Agg Events"],
  ["/torrent-albums", "torrent-albums", "Torrent Albums"],
  ["/movie-albums", "movie-albums", "Movie Albums"],
  ["/d-seeds", "d-seeds", "D Seeds"],
  ["/casts", "casts", "Casts"],
  ["/e-items", "e-items", "E Items"],
];

const movieAggPages = [
  ["/cardsagg", "cardsagg", "Cards Agg"],
  ["/movie-agg-all/release", "movie-agg-release", "Movie Agg Release"],
  ["/movie-agg-all/release-bucket-list", "movie-release-bucket-list", "Release Bucket List"],
  ["/w-media-agg/bucket-list", "w-media-bucket-list", "WMedia Bucket List"],
  ["/w-media-agg/birth", "w-media-birth", "WMedia Birth"],
  ["/triggers/agg", "agg-triggers", "Agg Triggers"],
  ["/wdir", "wdir", "W Directory"],
];

const scPages = [
  ["/sc-events", "sc-events", "SC Events"],
  ["/sc-events-cards", "sc-events-cards", "SC Event Cards"],
  ["/sc-pick-smart-media", "sc-pick-smart-media", "SC Pick Smart Media"],
  ["/triggers/sc", "triggers-sc", "SC Preview"],
  ["/triggers/media", "triggers-media", "Media Ingest"],
  ["/triggers/media-rescan", "triggers-media-rescan", "Media Rescan"],
  ["/triggers/media-rollback", "triggers-media-rollback", "Media Rollback"],
  ["/triggers/sc-media", "triggers-sc-media", "SC Media Move"],
];

const triggerPages = [
  ["/triggers/dailybest", "dailybest", "DailyBest"],
  ["/triggers/seeds", "seeds", "Seeds"],
  ["/triggers/post-process", "post-process", "Post Process"],
  ["/triggers/backfill", "backfill", "Backfill"],
  ["/triggers/fetch-site", "fetch-site-trigger", "Fetch Site"],
  ["/triggers/fetch-sehuatang", "fetch-sehuatang", "Fetch Sehuatang"],
  ["/triggers/fetch-sehuatang-sh", "fetch-sehuatang-sh", "Sehuatang SH"],
  ["/triggers/fetch-sehuatang-sh-thread-list", "fetch-sehuatang-sh-thread-list", "SH Thread List"],
  ["/triggers/fetch-sehuatang-sh-thread-detail", "fetch-sehuatang-sh-thread-detail", "SH Thread Detail"],
  ["/triggers/fetch-site-javbus-filtered", "fetch-site-javbus-filtered", "Javbus Filtered"],
  ["/triggers/fetch-site-sukebei-filtered", "fetch-site-sukebei-filtered", "Sukebei Filtered"],
];

const fetchSitePages = [
  ["/fetch-site-javbus-list", "fetch-site-javbus-list", "Fetch Site Javbus"],
  ["/fetch-site-sukebei-list", "fetch-site-sukebei-list", "Fetch Site Sukebei"],
  ["/fetch-site-sehuatang-list", "fetch-site-sehuatang-list", "Fetch Site Sehuatang"],
  ["/fetch-site-sehuatang-sh-list", "fetch-site-sehuatang-sh-list", "Fetch Site Sehuatang SH"],
];

async function main() {
  const onlyIndex = process.argv.indexOf("--only");
  const only = onlyIndex >= 0 ? String(process.argv[onlyIndex + 1] || "").padStart(2, "0") : "";
  for (const version of versions.filter((item) => !only || item.id === only)) {
    await writeVersion(version);
  }
}

async function writeVersion(version) {
  const dir = path.join(root, "gc-web-next" + version.id);
  const files = {
    "index.html": indexHtml(version),
    "package.json": packageJson(version),
    "vite.config.ts": viteConfig(version),
    "tsconfig.json": tsconfigJson(),
    "tsconfig.app.json": tsconfigAppJson(),
    "tsconfig.node.json": tsconfigNodeJson(),
    "src/App.vue": appVue(),
    "src/main.ts": mainTs(),
    "src/env.d.ts": envDts(),
    "src/api/request.ts": requestTs(),
    "src/types/gc.ts": typesTs(),
    "src/utils/view.ts": utilsTs(),
    "src/router/index.ts": routerTs(version),
    "src/layouts/AppLayout.vue": appLayoutVue(version),
    "src/pages/CardsPage.vue": cardsPageVue(),
    "src/pages/RankPage.vue": rankPageVue(),
    "src/pages/MoviePage.vue": moviePageVue(),
    "src/pages/DataPage.vue": dataPageVue(),
    "src/pages/CastPage.vue": castPageVue(),
    "src/pages/FetchSitesPage.vue": fetchSitesPageVue(),
    "src/pages/NotFoundPage.vue": notFoundPageVue(),
    "src/styles/index.css": css(version),
  };
  for (const [relative, content] of Object.entries(files)) {
    const target = path.join(dir, relative);
    await mkdir(path.dirname(target), {recursive: true});
    await writeFile(target, content);
  }
  console.log("rebuilt " + path.relative(root, dir) + " from " + version.styleSource);
}

function indexHtml(version) {
  return `<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>GC Web Next ${version.id}</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.ts"></script>
  </body>
</html>
`;
}

function packageJson(version) {
  return JSON.stringify(
    {
      name: "rudy-gc-web-next" + version.id,
      version: "0.1.0",
      private: true,
      type: "module",
      scripts: {
        dev: "vite --host 127.0.0.1",
        check: "vue-tsc --noEmit",
        build: "vue-tsc --noEmit && vite build",
        preview: "vite preview --host 127.0.0.1",
      },
      dependencies: {
        vue: "^3.5.13",
        "vue-router": "^4.5.0",
      },
      devDependencies: {
        "@tsconfig/node22": "^22.0.1",
        "@types/node": "^22.10.2",
        "@vitejs/plugin-vue": "^5.2.1",
        "@vue/tsconfig": "^0.7.0",
        typescript: "~5.7.2",
        vite: "^6.0.5",
        "vue-tsc": "^2.2.0",
      },
    },
    null,
    2,
  ) + "\n";
}

function viteConfig(version) {
  return `import fs from "node:fs";
import type {IncomingMessage, ServerResponse} from "node:http";
import path from "node:path";
import {fileURLToPath, URL} from "node:url";
import {defineConfig} from "vite";
import vue from "@vitejs/plugin-vue";

const textRoot = path.resolve(fileURLToPath(new URL("..", import.meta.url)), "z_text");
const volumeRoots = [
  "/Volumes/Expansion",
  "/Volumes/Getea",
  "/Volumes/movie-un",
  "/Volumes/T7/data",
];
const contentTypes: Record<string, string> = {
  ".css": "text/css; charset=utf-8",
  ".html": "text/html; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".txt": "text/plain; charset=utf-8",
  ".png": "image/png",
  ".jpg": "image/jpeg",
  ".jpeg": "image/jpeg",
  ".gif": "image/gif",
  ".svg": "image/svg+xml",
  ".mp4": "video/mp4",
  ".m4v": "video/mp4",
  ".mov": "video/quicktime",
  ".webm": "video/webm",
  ".mkv": "video/x-matroska",
  ".avi": "video/x-msvideo",
  ".ts": "video/mp2t",
  ".m2ts": "video/mp2t",
  ".woff": "font/woff",
  ".woff2": "font/woff2",
};

function contentTypeFor(filePath: string) {
  return contentTypes[path.extname(filePath).toLowerCase()] || "application/octet-stream";
}

function parseRangeHeader(rangeHeader: string | undefined, size: number) {
  const match = /^bytes=(\\d*)-(\\d*)$/.exec(String(rangeHeader || "").trim());
  if (!match) return null;
  const rawStart = match[1];
  const rawEnd = match[2];
  if (rawStart === "" && rawEnd === "") return null;
  let start: number;
  let end: number;
  if (rawStart === "") {
    const suffixLength = Number.parseInt(rawEnd, 10);
    if (!Number.isFinite(suffixLength) || suffixLength <= 0) return null;
    start = Math.max(size - suffixLength, 0);
    end = size - 1;
  } else {
    start = Number.parseInt(rawStart, 10);
    end = rawEnd === "" ? size - 1 : Number.parseInt(rawEnd, 10);
  }
  if (!Number.isFinite(start) || !Number.isFinite(end) || start < 0 || end < start || start >= size) return null;
  return {start, end: Math.min(end, size - 1)};
}

function sendStreamFile(req: IncomingMessage, res: ServerResponse, filePath: string) {
  fs.stat(filePath, (err, stat) => {
    if (err || !stat.isFile()) {
      res.writeHead(err?.code === "ENOENT" ? 404 : 500);
      res.end(err?.code === "ENOENT" ? "not found" : "server error");
      return;
    }
    const size = stat.size;
    const baseHeaders = {
      "Accept-Ranges": "bytes",
      "Content-Disposition": "inline",
      "Content-Type": contentTypeFor(filePath),
    };
    const range = parseRangeHeader(req.headers.range, size);
    if (req.headers.range && !range) {
      res.writeHead(416, {...baseHeaders, "Content-Range": "bytes */" + size});
      res.end();
      return;
    }
    if (range) {
      res.writeHead(206, {
        ...baseHeaders,
        "Content-Length": String(range.end - range.start + 1),
        "Content-Range": "bytes " + range.start + "-" + range.end + "/" + size,
      });
      if (req.method === "HEAD") {
        res.end();
        return;
      }
      fs.createReadStream(filePath, {start: range.start, end: range.end}).pipe(res);
      return;
    }
    res.writeHead(200, {...baseHeaders, "Content-Length": String(size)});
    if (req.method === "HEAD") {
      res.end();
      return;
    }
    fs.createReadStream(filePath).pipe(res);
  });
}

function safeJoin(root: string, requestPath: string) {
  const target = path.normalize(path.join(root, decodeURIComponent(requestPath)));
  if (!target.startsWith(root + path.sep) && target !== root) return "";
  return target;
}

function resolveVolumePath(requestPath: string) {
  const normalized = path.normalize(decodeURIComponent(requestPath));
  return volumeRoots.some((root) => normalized === root || normalized.startsWith(root + path.sep)) ? normalized : "";
}

function localFileMiddleware() {
  return {
    name: "gc-local-file-middleware",
    configureServer(server: any) {
      server.middlewares.use((req: IncomingMessage, res: ServerResponse, next: () => void) => {
        const url = new URL(req.url || "/", "http://127.0.0.1");
        if (url.pathname.startsWith("/Volumes/")) {
          const filePath = resolveVolumePath(url.pathname);
          if (filePath) {
            sendStreamFile(req, res, filePath);
            return;
          }
        }
        if (url.pathname.startsWith("/text/")) {
          const filePath = safeJoin(textRoot, url.pathname.replace(/^\\/text\\/?/, ""));
          if (filePath) {
            sendStreamFile(req, res, filePath);
            return;
          }
        }
        next();
      });
    },
  };
}

export default defineConfig({
  plugins: [vue(), localFileMiddleware()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    host: "127.0.0.1",
    port: Number(process.env.PORT || ${version.port}),
    proxy: {
      "/api": {
        target: process.env.V3_API_ORIGIN || "http://127.0.0.1:5013",
        changeOrigin: true,
      },
    },
  },
});
`;
}

function tsconfigJson() {
  return `{
  "files": [],
  "references": [
    {"path": "./tsconfig.app.json"},
    {"path": "./tsconfig.node.json"}
  ]
}
`;
}

function tsconfigAppJson() {
  return `{
  "extends": "@vue/tsconfig/tsconfig.dom.json",
  "compilerOptions": {
    "tsBuildInfoFile": "./node_modules/.tmp/tsconfig.app.tsbuildinfo",
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    },
    "strict": true,
    "noUnusedLocals": false,
    "noUnusedParameters": false
  },
  "include": ["src/**/*.ts", "src/**/*.tsx", "src/**/*.vue"]
}
`;
}

function tsconfigNodeJson() {
  return `{
  "extends": "@tsconfig/node22/tsconfig.json",
  "compilerOptions": {
    "tsBuildInfoFile": "./node_modules/.tmp/tsconfig.node.tsbuildinfo",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "types": ["node"],
    "strict": true,
    "noEmit": true
  },
  "include": ["vite.config.ts"]
}
`;
}

function appVue() {
  return `<template>
  <RouterView />
</template>
`;
}

function mainTs() {
  return `import {createApp} from "vue";
import App from "./App.vue";
import router from "./router";
import "./styles/index.css";

createApp(App).use(router).mount("#app");
`;
}

function envDts() {
  return `/// <reference types="vite/client" />
`;
}

function requestTs() {
  return `export type QueryValue = string | number | boolean | null | undefined;

export async function getJSON(path: string, query: Record<string, QueryValue> = {}) {
  const url = new URL("/api" + normalizePath(path), window.location.origin);
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined && value !== null && String(value).trim() !== "") {
      url.searchParams.set(key, String(value));
    }
  }
  const response = await fetch(url.toString(), {
    headers: {Accept: "application/json"},
  });
  if (!response.ok) {
    throw new Error(response.status + " " + response.statusText);
  }
  return response.json();
}

function normalizePath(path: string) {
  if (!path) return "/";
  return path.startsWith("/") ? path : "/" + path;
}
`;
}

function typesTs() {
  return `export type AnyRecord = Record<string, any>;

export interface NavItem {
  path: string;
  label: string;
  match?: string[];
  description?: string;
}

export interface GcPageMeta {
  title?: string;
  api?: string;
  mode?: string;
  group?: string;
}
`;
}

function utilsTs() {
  return `import type {RouteLocationNormalizedLoaded} from "vue-router";
import type {AnyRecord} from "@/types/gc";

export function asArray<T = AnyRecord>(value: unknown): T[] {
  if (Array.isArray(value)) return value as T[];
  if (value === undefined || value === null) return [];
  return [value as T];
}

export function field(record: unknown, ...keys: string[]) {
  const source = isObject(record) ? record as AnyRecord : {};
  for (const key of keys) {
    if (source[key] !== undefined && source[key] !== null && source[key] !== "") return source[key];
  }
  return "";
}

export function text(value: unknown): string {
  if (value === undefined || value === null) return "";
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean" || typeof value === "bigint") return String(value);
  return JSON.stringify(value);
}

export function queryObject(route: RouteLocationNormalizedLoaded): Record<string, string> {
  const query: Record<string, string> = {};
  for (const [key, value] of Object.entries(route.query)) {
    if (Array.isArray(value)) query[key] = value.map((item) => String(item ?? "")).join(",");
    else if (value !== undefined && value !== null) query[key] = String(value);
  }
  return query;
}

export function apiFromRoute(route: RouteLocationNormalizedLoaded) {
  const meta = route.meta as AnyRecord;
  return String(meta.api || route.path || "/");
}

export function pageLinks(data: AnyRecord) {
  const candidates = [
    data.PageLinks,
    data.pageLinks,
    data.Links,
    data.links,
    data.Pagination?.Links,
    data.pagination?.links,
    data.PageInfo?.Links,
    data.pageInfo?.links,
  ];
  for (const candidate of candidates) {
    const links = normalizeLinks(candidate);
    if (links.length) return links;
  }
  return [];
}

export function namedLinks(value: unknown) {
  if (!value) return [];
  if (Array.isArray(value)) {
    return value.map((item, index) => linkFromValue(String(index), item));
  }
  if (isObject(value)) {
    return Object.entries(value as AnyRecord).map(([key, item]) => linkFromValue(key, item));
  }
  return [];
}

export function primaryRows(data: AnyRecord) {
  const keys = ["Rows", "rows", "Items", "items", "Movies", "movies", "Cards", "cards", "Data", "data", "Records", "records", "Tasks", "tasks"];
  for (const key of keys) {
    const rows = asArray<AnyRecord>(data[key]);
    if (rows.length) return {key, rows};
  }
  for (const [key, value] of Object.entries(data)) {
    if (Array.isArray(value) && value.length && value.some((item) => isObject(item))) return {key, rows: value as AnyRecord[]};
  }
  return {key: "Rows", rows: [] as AnyRecord[]};
}

export function columnsFor(rows: AnyRecord[]) {
  const seen = new Set<string>();
  const columns: string[] = [];
  for (const row of rows) {
    for (const key of Object.keys(row || {})) {
      if (!seen.has(key)) {
        seen.add(key);
        columns.push(key);
      }
    }
  }
  return columns;
}

export function titleOf(data: AnyRecord, fallback: string) {
  return text(data.PageTitle || data.Title || data.title || fallback);
}

export function hrefFor(key: string, value: unknown) {
  const raw = text(value);
  if (!raw) return "";
  const lower = key.toLowerCase();
  if (lower === "name" || lower === "jav_id" || lower === "javid" || lower === "movie") return "/movie/" + encodeURIComponent(raw);
  if (lower.includes("personid") || lower === "castid" || lower === "actorid") return "/cast/" + encodeURIComponent(raw);
  if (lower.includes("url") || raw.startsWith("http://") || raw.startsWith("https://")) return raw;
  return "";
}

function normalizeLinks(value: unknown) {
  if (!Array.isArray(value)) return [];
  return value.map((item, index) => {
    const link = isObject(item) ? item as AnyRecord : {Label: text(item)};
    return {
      Label: text(link.Label || link.label || link.Text || link.text || index + 1),
      Href: text(link.Href || link.href || link.URL || link.url || "#"),
      Active: Boolean(link.Active || link.active || link.Current || link.current),
      Disabled: Boolean(link.Disabled || link.disabled),
    };
  });
}

function linkFromValue(key: string, item: unknown) {
  const link = isObject(item) ? item as AnyRecord : {Label: item};
  return {
    key,
    label: text(link.Label || link.label || link.Text || link.text || link.Name || link.name || key),
    href: text(link.Href || link.href || link.URL || link.url || "#"),
    active: Boolean(link.Active || link.active || link.Current || link.current),
  };
}

function isObject(value: unknown): value is AnyRecord {
  return typeof value === "object" && value !== null;
}
`;
}

function routerTs(version) {
  return `import {createRouter, createWebHistory, type RouteLocationNormalizedLoaded, type RouteRecordRaw} from "vue-router";
import type {GcPageMeta} from "@/types/gc";

type TitleMeta = string | ((to: RouteLocationNormalizedLoaded) => string);

const appTitle = "GC Web Next ${version.id}";

const cardPages = ${arrayLiteral(cardsPages)};
const listPages = ${arrayLiteral(listPages)};
const movieAggPages = ${arrayLiteral(movieAggPages)};
const scPages = ${arrayLiteral(scPages)};
const triggerPages = ${arrayLiteral(triggerPages)};
const fetchSitePages = ${arrayLiteral(fetchSitePages)};

function page(path: string, name: string, title: string, component: () => Promise<unknown>, meta: GcPageMeta): RouteRecordRaw {
  return {path: path.replace(/^\\//, ""), name, component, meta: {...meta, title}};
}

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    component: () => import("@/layouts/AppLayout.vue"),
    children: [
      {path: "", redirect: "/cards"},
      {path: "records", redirect: "/crawl-records"},
      ...cardPages.map(([path, name, title]) =>
        page(path, name, title, () => import("@/pages/CardsPage.vue"), {api: path, mode: "cards", group: "MovieCard"}),
      ),
      page("/moviecarddayrank", "moviecarddayrank", "MovieCard DayRank", () => import("@/pages/RankPage.vue"), {
        api: "/moviecarddayrank",
        mode: "rank",
        group: "Rank",
      }),
      page("/moviecardperiodrank", "moviecardperiodrank", "MovieCard PeriodRank", () => import("@/pages/RankPage.vue"), {
        api: "/moviecardperiodrank",
        mode: "rank",
        group: "Rank",
      }),
      page("/movie/:movieName", "movie-detail", "Movie Detail", () => import("@/pages/MoviePage.vue"), {
        mode: "movie",
        group: "Movie",
      }),
      page("/cast", "cast-search", "Cast", () => import("@/pages/CastPage.vue"), {mode: "cast", group: "Cast"}),
      page("/cast/:personId", "cast-detail", "Cast Detail", () => import("@/pages/CastPage.vue"), {
        mode: "cast",
        group: "Cast",
      }),
      ...listPages.map(([path, name, title]) =>
        page(path, name, title, () => import("@/pages/DataPage.vue"), {api: path, mode: "list", group: "List"}),
      ),
      ...movieAggPages.map(([path, name, title]) =>
        page(path, name, title, () => import("@/pages/DataPage.vue"), {api: path, mode: "movieAgg", group: "MovieAgg"}),
      ),
      ...scPages.map(([path, name, title]) =>
        page(path, name, title, () => import("@/pages/DataPage.vue"), {api: path, mode: "sc", group: "SC"}),
      ),
      page("/sc-events/:eventName", "sc-event-detail", "SC Event Detail", () => import("@/pages/DataPage.vue"), {
        mode: "sc",
        group: "SC",
      }),
      ...triggerPages.map(([path, name, title]) =>
        page(path, name, title, () => import("@/pages/DataPage.vue"), {api: path, mode: "trigger", group: "Crawler"}),
      ),
      page("/fetch-sites", "fetch-sites", "Fetch Sites", () => import("@/pages/FetchSitesPage.vue"), {
        api: "/fetch-sites",
        mode: "fetchSite",
        group: "FetchSite",
      }),
      ...fetchSitePages.map(([path, name, title]) =>
        page(path, name, title, () => import("@/pages/DataPage.vue"), {api: path, mode: "fetchSite", group: "FetchSite"}),
      ),
      page("/fetch-site-sehuatang-sh-detail/:threadTid", "fetch-site-sehuatang-sh-detail", "Sehuatang SH Detail", () => import("@/pages/DataPage.vue"), {
        mode: "fetchSite",
        group: "FetchSite",
      }),
      page("/crawler/tasks", "crawler-tasks", "Crawler Tasks", () => import("@/pages/DataPage.vue"), {
        api: "/crawler/tasks",
        mode: "crawler",
        group: "Crawler",
      }),
      page("/crawler/detail-loop", "crawler-detail-loop", "Crawler Detail Loop", () => import("@/pages/DataPage.vue"), {
        api: "/crawler/detail-loop",
        mode: "crawler",
        group: "Crawler",
      }),
      {path: ":pathMatch(.*)*", name: "not-found", component: () => import("@/pages/NotFoundPage.vue"), meta: {title: "Not Found"}},
    ],
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.afterEach((to) => {
  const title = to.meta.title as TitleMeta | undefined;
  const pageTitle = typeof title === "function" ? title(to) : title;
  document.title = pageTitle ? pageTitle + " | " + appTitle : appTitle;
});

export default router;
`;
}

function appLayoutVue(version) {
  return `<template>
  <div class="gc-shell source-${version.sourceVersion}">
    <aside class="gc-sidebar">
      <RouterLink to="/cards" class="gc-brand">
        <strong>GC Next ${version.id}</strong>
        <span>${version.sourceVersion} source style</span>
      </RouterLink>
      <nav class="gc-side-nav" aria-label="Sidebar">
        <section v-for="group in sideGroups" :key="group.label" class="gc-side-group">
          <h2>{{ group.label }}</h2>
          <RouterLink
            v-for="item in group.items"
            :key="item.path"
            :to="item.path"
            :class="{ active: isActive(item.path, item.match) }"
          >
            <span>{{ item.label }}</span>
          </RouterLink>
        </section>
      </nav>
    </aside>

    <div class="gc-workspace">
      <header class="gc-topbar">
        <div>
          <p>{{ currentGroup }}</p>
          <h1>{{ currentTitle }}</h1>
        </div>
        <nav class="gc-topnav" aria-label="Quick navigation">
          <span v-for="group in topGroups" :key="group.label" class="gc-topnav-group">
            <b>{{ group.label }}</b>
            <RouterLink v-for="item in group.items" :key="item.path" :to="item.path">{{ item.label }}</RouterLink>
          </span>
        </nav>
      </header>

      <details class="gc-mobile-nav">
        <summary>页面导航</summary>
        <div>
          <section v-for="group in sideGroups" :key="group.label">
            <h2>{{ group.label }}</h2>
            <RouterLink v-for="item in group.items" :key="item.path" :to="item.path">{{ item.label }}</RouterLink>
          </section>
        </div>
      </details>

      <main class="gc-main">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import {computed} from "vue";
import {useRoute} from "vue-router";
import type {NavItem} from "@/types/gc";

const route = useRoute();

const topGroups: Array<{label: string; items: NavItem[]}> = ${navLiteral(topGroups)};
const sideGroups: Array<{label: string; items: NavItem[]}> = ${navLiteral(navGroups)};

const currentTitle = computed(() => String(route.meta.title || "GC Web"));
const currentGroup = computed(() => String(route.meta.group || route.meta.mode || "Page"));

function isActive(path: string, match?: string[]) {
  if (match?.some((item) => route.path === item || route.path.startsWith(item + "/"))) return true;
  return route.path === path;
}
</script>
`;
}

function cardsPageVue() {
  return `<template>
  <section class="gc-page gc-cards-page">
    <div class="gc-sticky">
      <nav class="gc-sortbar" aria-label="Sort">
        <a v-for="item in sortButtons" :key="item.key" :href="item.href" :class="{ active: item.active }">
          {{ item.label }}
        </a>
      </nav>
      <div class="gc-owned-row">
        <div class="gc-owned-left">
          <span class="gc-chip muted">WMedia</span>
          <a v-for="item in ownedButtons" :key="item.key" :href="item.href" class="gc-chip" :class="{ active: item.active }">
            {{ item.label }}
          </a>
          <span class="gc-chip primary">Total: {{ total }}</span>
          <button class="gc-chip" type="button" @click="filterOpen = !filterOpen">筛选</button>
          <button v-if="pageDateLabel" class="gc-chip" type="button">{{ pageDateLabel }}</button>
        </div>
        <div class="gc-pagination">
          <a v-for="link in links" :key="link.Label + link.Href" :href="link.Href || '#'" :class="{ active: link.Active, disabled: link.Disabled }">
            {{ link.Label }}
          </a>
        </div>
      </div>
    </div>

    <form v-show="filterOpen" class="gc-filter-panel" @submit.prevent="submitFilter">
      <label v-for="item in filterFields" :key="item.name">
        <span>{{ item.label }}</span>
        <input v-model="filterModel[item.name]" :placeholder="item.placeholder" />
      </label>
      <label>
        <span>w_media 库存</span>
        <select v-model="filterModel.mowned">
          <option value="">全部</option>
          <option value="3">已拥有未移除</option>
          <option value="7">无 w_media 记录</option>
        </select>
      </label>
      <div class="gc-filter-actions">
        <button class="gc-button primary" type="submit">筛选</button>
        <RouterLink class="gc-button" :to="route.path">清空</RouterLink>
      </div>
    </form>

    <div v-if="error" class="gc-state danger">{{ error }}</div>
    <div v-else-if="loading" class="gc-state">加载中</div>

    <div v-else class="gc-movie-grid">
      <article v-for="movie in movies" :key="movie.name || movie.jav_id" class="gc-movie-card">
        <RouterLink class="gc-poster" :to="moviePath(movie)">
          <img v-if="movie.jacket_img" :src="movie.jacket_img" :alt="movie.name || movie.jav_id" loading="lazy" />
          <span v-else>No Poster</span>
          <b v-if="movie.highest_rank">#{{ movie.highest_rank }}</b>
        </RouterLink>
        <div class="gc-movie-body">
          <h2>
            <RouterLink :to="moviePath(movie)">{{ movie.name || movie.jav_id }}</RouterLink>
            <small v-if="movie.viewers_number_watched">({{ movie.viewers_number_watched }})</small>
          </h2>
          <p>{{ movie.title }}</p>
          <div class="gc-tag-row">
            <a v-for="genre in asArray<string>(movie.genre)" :key="genre" :href="genreHref(genre)">{{ genre }}</a>
          </div>
          <div class="gc-tag-row">
            <a v-for="cast in asArray(movie.cast)" :key="cast.Name || cast.NameShow" :href="castHref(cast)">
              {{ cast.NameShow || cast.DisplayName || cast.Name }}
            </a>
            <a v-if="movie.director" :href="directorHref(movie.director)">{{ movie.director }}</a>
          </div>
          <div class="gc-card-foot">
            <span>{{ movie.releasing_date }}</span>
            <a v-if="movie.bus_url" :href="movie.bus_url" target="_blank">Bus</a>
            <a v-if="movie.search_url" :href="movie.search_url" target="_blank">Suk</a>
            <a v-if="movie.jav_url" :href="movie.jav_url" target="_blank">Lib</a>
          </div>
        </div>
      </article>
      <div v-if="movies.length === 0" class="gc-empty">暂无数据</div>
    </div>
  </section>
</template>

<script setup lang="ts">
import {computed, reactive, ref, watch} from "vue";
import {useRoute, useRouter} from "vue-router";
import {getJSON} from "@/api/request";
import type {AnyRecord} from "@/types/gc";
import {asArray, namedLinks, pageLinks, queryObject} from "@/utils/view";

const sortLabels: Record<string, string> = {
  ByDetailUpdate: "Detail更新",
  ByReleasing: "上映日期",
  ByViewer: "观看人数",
  ByRankDate: "上榜日期",
  ByDaysInRank: "在榜天数",
  ByHighestRank: "最高排名",
  ByBirthTime: "下载时间",
  ByMediaBirthTime: "Media下载时间",
  ByLastScTime: "最后sc",
  ByScTimes: "sc次数",
  ByComeTimes: "co次数",
};

const filterFields = [
  {name: "cn", label: "演员", placeholder: "演员名，支持多值"},
  {name: "gn", label: "类型", placeholder: "类型名，支持多值"},
  {name: "dn", label: "导演", placeholder: "导演名"},
  {name: "pn", label: "前缀", placeholder: "前缀"},
  {name: "mn", label: "片商", placeholder: "片商"},
  {name: "ln", label: "标签", placeholder: "标签"},
  {name: "wd", label: "关键词", placeholder: "中文关键词"},
];

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const error = ref("");
const data = ref<AnyRecord>({});
const filterOpen = ref(false);
const filterModel = reactive<Record<string, string>>({});

const apiPath = computed(() => String(route.meta.api || route.path));
const movies = computed(() => asArray<AnyRecord>(data.value.movies || data.value.Movies));
const total = computed(() => data.value.total || data.value.Total || data.value.PageInfo?.Total || movies.value.length);
const links = computed(() => pageLinks(data.value));
const sortButtons = computed(() => {
  const raw = namedLinks(data.value.sortQuery || data.value.SortQuery);
  return Object.entries(sortLabels).map(([key, label]) => {
    const found = raw.find((link) => link.key === key);
    return found ? {...found, label} : {key, label, href: route.fullPath, active: false};
  });
});
const ownedButtons = computed(() => namedLinks(data.value.ownedQuery || data.value.OwnedQuery).map((item) => ({...item, label: item.label.replace("Media", "M")})));
const pageDateLabel = computed(() => {
  const edit = data.value.PageDateEdit;
  return edit ? String((edit.Label || "日期") + "：" + (edit.DisplayValue || "未设置")) : "";
});

watch(
  () => route.fullPath,
  async () => {
    loading.value = true;
    error.value = "";
    Object.assign(filterModel, queryObject(route));
    try {
      data.value = await getJSON(apiPath.value, queryObject(route));
    } catch (err: any) {
      error.value = err?.message || "加载失败";
    } finally {
      loading.value = false;
    }
  },
  {immediate: true},
);

function submitFilter() {
  const query = Object.fromEntries(Object.entries(filterModel).filter(([, value]) => String(value || "").trim() !== ""));
  router.push({path: route.path, query});
}

function moviePath(movie: AnyRecord) {
  return "/movie/" + encodeURIComponent(String(movie.name || movie.jav_id || ""));
}

function genreHref(genre: string) {
  return "/cards?gn=" + encodeURIComponent(String(genre || ""));
}

function directorHref(director: string) {
  return "/cards?dn=" + encodeURIComponent(String(director || ""));
}

function castHref(cast: AnyRecord) {
  const name = String(cast.Name || cast.NameShow || "");
  return cast.PersonId ? "/cast/" + encodeURIComponent(String(cast.PersonId)) : "/cards?cn=" + encodeURIComponent(name);
}
</script>
`;
}

function rankPageVue() {
  return `<template>
  <section class="gc-page">
    <div class="gc-sticky rank">
      <div class="gc-rank-summary">
        <div><strong>{{ isPeriod ? data.periodTypeLabel : "Rank Date" }}</strong><span>{{ isPeriod ? data.periodKey : data.rankDate }}</span></div>
        <div v-if="isPeriod"><strong>范围</strong><span>{{ data.rangeStart }} ~ {{ data.rangeEnd }}</span></div>
        <div><strong>Total</strong><span>{{ data.total || data.Total || rows.length }}</span></div>
      </div>
      <div class="gc-owned-row">
        <div class="gc-owned-left">
          <a v-if="!isPeriod" class="gc-chip" :href="dayHref(data.prevDate)">前一天</a>
          <a v-if="!isPeriod" class="gc-chip" :href="dayHref(data.nextDate)">后一天</a>
          <a v-if="!isPeriod" class="gc-chip" :href="dayHref(data.randDate)">随机</a>
          <a v-if="isPeriod" class="gc-chip" :href="data.prevHref || '#'">上一期</a>
          <a v-if="isPeriod" class="gc-chip" :href="data.nextHref || '#'">下一期</a>
          <a v-if="isPeriod" class="gc-chip" :href="data.latestHref || '/moviecardperiodrank'">最新</a>
        </div>
        <div class="gc-pagination">
          <a v-for="link in links" :key="link.Label + link.Href" :href="link.Href || '#'" :class="{ active: link.Active, disabled: link.Disabled }">
            {{ link.Label }}
          </a>
        </div>
      </div>
    </div>

    <div v-if="error" class="gc-state danger">{{ error }}</div>
    <div v-else-if="loading" class="gc-state">加载中</div>

    <div v-else class="gc-movie-grid">
      <article v-for="row in rows" :key="movieOf(row).name || movieOf(row).jav_id || row.Item?.ID" class="gc-movie-card">
        <RouterLink class="gc-poster" :to="moviePath(movieOf(row))">
          <img v-if="movieOf(row).jacket_img" :src="movieOf(row).jacket_img" :alt="movieOf(row).name" loading="lazy" />
          <span v-else>No Poster</span>
          <b v-if="movieOf(row).highest_rank">#{{ movieOf(row).highest_rank }}</b>
        </RouterLink>
        <div class="gc-movie-body">
          <h2><RouterLink :to="moviePath(movieOf(row))">{{ movieOf(row).name || movieOf(row).jav_id }}</RouterLink></h2>
          <p>{{ movieOf(row).title }}</p>
          <div v-if="isPeriod" class="gc-rank-lines">
            <span>周期 #{{ row.Item?.RankPos }}</span>
            <span>{{ row.RankChangeText }}</span>
            <small>上期 {{ row.PrevRankText || "--" }} · 峰值 #{{ row.Item?.PeakRank }} · 分数 {{ scoreOf(row) }}</small>
          </div>
          <div class="gc-tag-row">
            <a v-for="genre in asArray<string>(movieOf(row).genre)" :key="genre" :href="genreHref(genre)">{{ genre }}</a>
          </div>
          <div class="gc-card-foot">
            <span>{{ movieOf(row).releasing_date }}</span>
            <a v-if="movieOf(row).bus_url" :href="movieOf(row).bus_url" target="_blank">Bus</a>
            <a v-if="movieOf(row).search_url" :href="movieOf(row).search_url" target="_blank">Suk</a>
          </div>
        </div>
      </article>
      <div v-if="rows.length === 0" class="gc-empty">暂无数据</div>
    </div>
  </section>
</template>

<script setup lang="ts">
import {computed, ref, watch} from "vue";
import {useRoute} from "vue-router";
import {getJSON} from "@/api/request";
import type {AnyRecord} from "@/types/gc";
import {asArray, pageLinks, queryObject} from "@/utils/view";

const route = useRoute();
const loading = ref(false);
const error = ref("");
const data = ref<AnyRecord>({});
const isPeriod = computed(() => data.value.ViewType === "rank-period" || route.path.includes("period"));
const rows = computed(() => (isPeriod.value ? asArray<AnyRecord>(data.value.cards || data.value.Cards) : asArray<AnyRecord>(data.value.movies || data.value.Movies)));
const links = computed(() => pageLinks(data.value));

watch(
  () => route.fullPath,
  async () => {
    loading.value = true;
    error.value = "";
    try {
      data.value = await getJSON(String(route.meta.api || route.path), queryObject(route));
    } catch (err: any) {
      error.value = err?.message || "加载失败";
    } finally {
      loading.value = false;
    }
  },
  {immediate: true},
);

function movieOf(row: AnyRecord) {
  return isPeriod.value ? row.Movie || row.movie || {} : row;
}

function scoreOf(row: AnyRecord) {
  return Number(row.Item?.Score || 0).toFixed(2);
}

function moviePath(movie: AnyRecord) {
  return "/movie/" + encodeURIComponent(String(movie.name || movie.jav_id || ""));
}

function genreHref(genre: string) {
  return "/cards?gn=" + encodeURIComponent(String(genre || ""));
}

function dayHref(date: unknown) {
  return "/moviecarddayrank?date=" + encodeURIComponent(String(date || ""));
}
</script>
`;
}

function moviePageVue() {
  return `<template>
  <section class="gc-page">
    <div v-if="error" class="gc-state danger">{{ error }}</div>
    <div v-else-if="loading" class="gc-state">加载中</div>

    <article v-else class="gc-detail">
      <div class="gc-detail-hero">
        <div class="gc-detail-poster">
          <img v-if="poster" :src="poster" :alt="movieName" />
          <span v-else>No Poster</span>
        </div>
        <div class="gc-detail-main">
          <span class="gc-kicker">Movie Detail</span>
          <h1>{{ movieName }}</h1>
          <p>{{ movieTitle }}</p>
          <div class="gc-detail-actions">
            <a v-if="movie.bus_url" :href="movie.bus_url" target="_blank" class="gc-button">Bus</a>
            <a v-if="movie.search_url" :href="movie.search_url" target="_blank" class="gc-button">Suk</a>
            <a v-if="movie.jav_url" :href="movie.jav_url" target="_blank" class="gc-button">Lib</a>
          </div>
          <dl class="gc-kv-grid">
            <div v-for="item in coreInfo" :key="item.label">
              <dt>{{ item.label }}</dt>
              <dd>{{ item.value }}</dd>
            </div>
          </dl>
        </div>
      </div>

      <section class="gc-panel">
        <div class="gc-panel-title">
          <h2>页面数据</h2>
          <span>{{ visibleEntries.length }} fields</span>
        </div>
        <div class="gc-table-wrap">
          <table>
            <tbody>
              <tr v-for="[key, value] in visibleEntries" :key="key">
                <th>{{ key }}</th>
                <td><pre>{{ format(value) }}</pre></td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </article>
  </section>
</template>

<script setup lang="ts">
import {computed, ref, watch} from "vue";
import {useRoute} from "vue-router";
import {getJSON} from "@/api/request";
import type {AnyRecord} from "@/types/gc";
import {apiFromRoute, field} from "@/utils/view";

const route = useRoute();
const data = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");

const movie = computed(() => data.value.movieType || data.value.MovieType || data.value.movie || data.value.Movie || data.value);
const movieName = computed(() => String(field(movie.value, "name", "Name", "jav_id", "JavID") || route.params.movieName || ""));
const movieTitle = computed(() => String(field(movie.value, "title", "Title") || ""));
const poster = computed(() => String(field(movie.value, "jacket_img", "JacketImg", "cover", "Cover") || ""));
const coreInfo = computed(() => [
  {label: "上映日期", value: field(movie.value, "releasing_date", "ReleasingDate")},
  {label: "评分", value: field(movie.value, "score", "Score")},
  {label: "片长", value: field(movie.value, "length", "Length")},
  {label: "片商", value: field(movie.value, "maker", "Maker")},
  {label: "导演", value: field(movie.value, "director", "Director")},
  {label: "最高排名", value: field(movie.value, "highest_rank", "HighestRank")},
]);
const visibleEntries = computed(() => Object.entries(data.value).filter(([key]) => !["movieType", "MovieType", "movie", "Movie"].includes(key)));

watch(
  () => route.fullPath,
  async () => {
    loading.value = true;
    error.value = "";
    try {
      data.value = await getJSON(apiFromRoute(route));
    } catch (err: any) {
      error.value = err?.message || "加载失败";
    } finally {
      loading.value = false;
    }
  },
  {immediate: true},
);

function format(value: unknown) {
  if (typeof value === "string" || typeof value === "number" || typeof value === "boolean") return String(value);
  return JSON.stringify(value, null, 2);
}
</script>
`;
}

function dataPageVue() {
  return `<template>
  <section class="gc-page">
    <header class="gc-data-head">
      <div>
        <span class="gc-kicker">{{ groupLabel }}</span>
        <h1>{{ pageTitle }}</h1>
        <p v-if="pageNote">{{ pageNote }}</p>
      </div>
      <div class="gc-stat-row">
        <span><b>Total</b>{{ total }}</span>
        <span><b>View</b>{{ data.ViewType || route.meta.mode }}</span>
      </div>
    </header>

    <div v-if="controlGroups.length" class="gc-toolbar">
      <section v-for="group in controlGroups" :key="group.name">
        <strong>{{ group.name }}</strong>
        <div>
          <a v-for="item in group.items" :key="item.key" :href="item.href" :class="{ active: item.active }">{{ item.label }}</a>
        </div>
      </section>
    </div>

    <section v-if="queryFields.length" class="gc-filter-panel">
      <form class="gc-form-grid" @submit.prevent="submitQuery">
        <label v-for="item in queryFields" :key="item.key">
          <span>{{ item.key }}</span>
          <input v-model="queryModel[item.key]" />
        </label>
        <div class="gc-filter-actions">
          <button class="gc-button primary" type="submit">查询</button>
          <RouterLink class="gc-button" :to="route.path">清空</RouterLink>
        </div>
      </form>
    </section>

    <div class="gc-pagination" v-if="links.length">
      <a v-for="link in links" :key="link.Label + link.Href" :href="link.Href || '#'" :class="{ active: link.Active, disabled: link.Disabled }">
        {{ link.Label }}
      </a>
    </div>

    <div v-if="error" class="gc-state danger">{{ error }}</div>
    <div v-else-if="loading" class="gc-state">加载中</div>

    <section v-else class="gc-panel">
      <div class="gc-panel-title">
        <h2>{{ primary.key }}</h2>
        <span>{{ primary.rows.length }} rows</span>
      </div>
      <div v-if="primary.rows.length" class="gc-table-wrap">
        <table>
          <thead>
            <tr>
              <th v-for="column in columns" :key="column">{{ column }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, index) in primary.rows" :key="index">
              <td v-for="column in columns" :key="column">
                <a v-if="href(column, row[column])" :href="href(column, row[column])">{{ cell(row[column]) }}</a>
                <span v-else>{{ cell(row[column]) }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="gc-empty">暂无数据</div>
    </section>

    <section v-if="otherEntries.length" class="gc-panel">
      <div class="gc-panel-title">
        <h2>页面字段</h2>
        <span>{{ otherEntries.length }} fields</span>
      </div>
      <dl class="gc-kv-grid">
        <div v-for="[key, value] in otherEntries" :key="key">
          <dt>{{ key }}</dt>
          <dd>{{ cell(value) }}</dd>
        </div>
      </dl>
    </section>
  </section>
</template>

<script setup lang="ts">
import {computed, reactive, ref, watch} from "vue";
import {useRoute, useRouter} from "vue-router";
import {getJSON} from "@/api/request";
import type {AnyRecord} from "@/types/gc";
import {apiFromRoute, columnsFor, hrefFor, namedLinks, pageLinks, primaryRows, queryObject, text, titleOf} from "@/utils/view";

const route = useRoute();
const router = useRouter();
const data = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const queryModel = reactive<Record<string, string>>({});

const groupLabel = computed(() => String(route.meta.group || route.meta.mode || "Data"));
const pageTitle = computed(() => titleOf(data.value, String(route.meta.title || "Page")));
const pageNote = computed(() => String(data.value.PageNote || data.value.Message || data.value.SkipReason || ""));
const primary = computed(() => primaryRows(data.value));
const total = computed(() => data.value.Total || data.value.total || data.value.PageInfo?.Total || primary.value.rows.length || 0);
const links = computed(() => pageLinks(data.value));
const columns = computed(() => columnsFor(primary.value.rows));
const queryFields = computed(() => Object.entries(data.value.Query || {}).map(([key, value]) => ({key, value})));
const controlGroups = computed(() => {
  const groups: Array<{name: string; items: ReturnType<typeof namedLinks>}> = [];
  for (const key of ["HeaderLinks", "TypeLinks", "SortLinks", "LevelLinks", "ModeSwitches", "QuickTags", "TaskButtons"]) {
    const items = namedLinks(data.value[key]);
    if (items.length) groups.push({name: key, items});
  }
  for (const key of ["SortQuery", "sortQuery", "ScSortQuery", "FilmSortQuery", "CastSortQuery", "ItemSortQuery", "DSeedSort", "ownedQuery", "ScViewQuery"]) {
    const items = namedLinks(data.value[key]);
    if (items.length) groups.push({name: key, items});
  }
  return groups;
});
const otherEntries = computed(() => {
  const skip = new Set([primary.value.key, "Rows", "rows", "Items", "items", "movies", "Movies", "cards", "Cards", "PageInfo", "pageInfo", "Query"]);
  return Object.entries(data.value).filter(([key, value]) => !skip.has(key) && !Array.isArray(value) && (typeof value !== "object" || value === null));
});

watch(
  () => route.fullPath,
  async () => {
    loading.value = true;
    error.value = "";
    Object.assign(queryModel, queryObject(route));
    try {
      data.value = await getJSON(apiFromRoute(route), queryObject(route));
      for (const [key, value] of Object.entries(data.value.Query || {})) {
        if (queryModel[key] === undefined) queryModel[key] = String(value ?? "");
      }
    } catch (err: any) {
      error.value = err?.message || "加载失败";
    } finally {
      loading.value = false;
    }
  },
  {immediate: true},
);

function submitQuery() {
  const query = Object.fromEntries(Object.entries(queryModel).filter(([, value]) => String(value || "").trim() !== ""));
  router.push({path: route.path, query});
}

function cell(value: unknown) {
  return text(value);
}

function href(key: string, value: unknown) {
  return hrefFor(key, value);
}
</script>
`;
}

function castPageVue() {
  return `<template>
  <section class="gc-page">
    <form v-if="isSearchOnly" class="gc-panel gc-cast-search" @submit.prevent="goCast">
      <span class="gc-kicker">Cast</span>
      <h1>Cast Search</h1>
      <div class="gc-form-grid">
        <label>
          <span>ID</span>
          <input v-model="castId" autocomplete="off" />
        </label>
        <label>
          <span>Name</span>
          <input v-model="castName" autocomplete="off" />
        </label>
      </div>
      <div class="gc-filter-actions">
        <button class="gc-button primary" type="submit">查询</button>
        <RouterLink class="gc-button" to="/casts">演员列表</RouterLink>
      </div>
    </form>

    <DataPage v-else />
  </section>
</template>

<script setup lang="ts">
import {computed, ref} from "vue";
import {useRoute, useRouter} from "vue-router";
import DataPage from "./DataPage.vue";

const route = useRoute();
const router = useRouter();
const castId = ref("");
const castName = ref("");
const isSearchOnly = computed(() => route.name === "cast-search" && !route.query.id && !route.query.name);

function goCast() {
  if (castId.value.trim()) {
    router.push("/cast/" + encodeURIComponent(castId.value.trim()));
    return;
  }
  if (castName.value.trim()) {
    router.push({path: "/cast", query: {name: castName.value.trim()}});
  }
}
</script>
`;
}

function fetchSitesPageVue() {
  return `<template>
  <section class="gc-page">
    <header class="gc-data-head">
      <div>
        <span class="gc-kicker">FetchSite</span>
        <h1>Fetch Sites</h1>
        <p>站点抓取配置。</p>
      </div>
      <div class="gc-stat-row">
        <span><b>Sites</b>{{ rows.length }}</span>
      </div>
    </header>

    <div v-if="error" class="gc-state danger">{{ error }}</div>
    <div v-else-if="loading" class="gc-state">加载中</div>

    <section v-else class="gc-panel">
      <div class="gc-table-wrap">
        <table>
          <thead>
            <tr>
              <th v-for="column in columns" :key="column">{{ column }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, index) in rows" :key="row.id || row.site_code || index">
              <td v-for="column in columns" :key="column">{{ row[column] ?? "-" }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </section>
</template>

<script setup lang="ts">
import {computed, ref, watch} from "vue";
import {useRoute} from "vue-router";
import {getJSON} from "@/api/request";
import type {AnyRecord} from "@/types/gc";
import {asArray, columnsFor} from "@/utils/view";

const route = useRoute();
const data = ref<AnyRecord>({});
const loading = ref(false);
const error = ref("");
const rows = computed(() => asArray<AnyRecord>(data.value.items || data.value.Items || data.value.Rows));
const columns = computed(() => columnsFor(rows.value));

watch(
  () => route.fullPath,
  async () => {
    loading.value = true;
    error.value = "";
    try {
      data.value = await getJSON("/fetch-sites");
    } catch (err: any) {
      error.value = err?.message || "加载失败";
    } finally {
      loading.value = false;
    }
  },
  {immediate: true},
);
</script>
`;
}

function notFoundPageVue() {
  return `<template>
  <section class="gc-page">
    <div class="gc-empty">
      <h1>页面不存在</h1>
      <RouterLink class="gc-button primary" to="/cards">返回 Cards</RouterLink>
    </div>
  </section>
</template>
`;
}

function css(version) {
  const t = version.theme;
  const v4 = version.id === "04";
  const v3 = version.id === "03";
  const v5 = version.id === "05";
  const bodyExtra = v4 ? "border-top: 5px solid var(--accent);" : "";
  const titleCase = v4 ? "text-transform: uppercase;" : "";
  const cardBorder = v4 ? "border-top: 3px solid var(--line);" : "border: var(--border-width) solid var(--line);";
  const cardPadding = v3 ? "padding: 0 0 14px;" : "padding: 10px;";
  return `/* Visual source locked to ${version.styleSource}. */
:root {
  color-scheme: ${v3 || v5 ? "dark" : "light"};
  --bg: ${t.bg};
  --bg-layer: ${t.bgLayer};
  --panel: ${t.panel};
  --panel-2: ${t.panel2};
  --ink: ${t.ink};
  --muted: ${t.muted};
  --line: ${t.line};
  --accent: ${t.accent};
  --accent-2: ${t.accent2};
  --danger: ${t.danger};
  --sidebar: ${t.sidebar};
  --sidebar-text: ${t.sidebarText};
  --topbar: ${t.topbar};
  --radius: ${t.radius};
  --button-radius: ${t.buttonRadius};
  --shadow: ${t.shadow};
  --card-shadow: ${t.cardShadow};
  --font-main: ${t.font};
  --font-display: ${t.displayFont};
  --border-width: ${t.borderWidth};
  --poster-radius: ${t.posterRadius};
  --topbar-shadow: ${t.topbarShadow};
}

*,
*::before,
*::after {
  box-sizing: border-box;
}

html,
body,
#app {
  min-height: 100%;
  margin: 0;
}

body {
  min-width: 320px;
  background: var(--bg-layer);
  color: var(--ink);
  font-family: var(--font-main);
  font-size: 14px;
  -webkit-font-smoothing: antialiased;
  ${bodyExtra}
}

a {
  color: inherit;
  text-decoration: none;
}

button,
input,
select {
  font: inherit;
}

button {
  cursor: pointer;
}

img {
  display: block;
  max-width: 100%;
}

.gc-shell {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  min-height: 100vh;
}

.gc-sidebar {
  position: sticky;
  top: 0;
  height: 100vh;
  overflow: auto;
  border-right: var(--border-width) solid var(--line);
  background: var(--sidebar);
  color: var(--sidebar-text);
  padding: 22px 18px;
}

.gc-brand {
  display: grid;
  gap: 5px;
  border: var(--border-width) solid var(--line);
  border-radius: var(--radius);
  background: var(--panel);
  color: var(--ink);
  padding: 14px;
  box-shadow: var(--topbar-shadow);
}

.gc-brand strong {
  font-family: var(--font-display);
  font-size: ${v4 ? "24px" : "18px"};
  font-weight: 900;
  line-height: 1;
  ${titleCase}
}

.gc-brand span {
  color: var(--accent);
  font-size: 12px;
  font-weight: 900;
}

.gc-side-nav {
  display: grid;
  gap: 18px;
  margin-top: 22px;
}

.gc-side-group {
  display: grid;
  gap: 8px;
}

.gc-side-group h2 {
  margin: 0;
  color: var(--accent);
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0;
  ${titleCase}
}

.gc-side-group a {
  display: grid;
  gap: 2px;
  border: var(--border-width) solid transparent;
  border-radius: var(--button-radius);
  color: var(--sidebar-text);
  padding: 10px 11px;
  font-weight: 800;
}

.gc-side-group a:hover,
.gc-side-group a.active {
  border-color: var(--accent-2);
  background: color-mix(in srgb, var(--accent-2) 12%, transparent);
  color: var(--accent-2);
}

.gc-workspace {
  min-width: 0;
}

.gc-topbar {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  min-height: 70px;
  border-bottom: var(--border-width) solid var(--line);
  background: color-mix(in srgb, var(--topbar) 94%, transparent);
  box-shadow: var(--topbar-shadow);
  padding: 12px 24px;
  backdrop-filter: blur(10px);
}

.gc-topbar p,
.gc-topbar h1 {
  margin: 0;
}

.gc-topbar p,
.gc-kicker,
.gc-panel-title span,
.gc-kv-grid dt,
.gc-toolbar strong {
  color: var(--accent);
  font-size: 12px;
  font-weight: 900;
  ${titleCase}
}

.gc-topbar h1 {
  color: var(--ink);
  font-family: var(--font-display);
  font-size: ${v4 ? "30px" : "22px"};
  font-weight: 900;
  line-height: 1.08;
}

.gc-topnav {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.gc-topnav-group {
  display: inline-flex;
  align-items: center;
}

.gc-topnav b,
.gc-topnav a,
.gc-chip,
.gc-sortbar a,
.gc-pagination a,
.gc-button,
.gc-toolbar a {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 30px;
  border: var(--border-width) solid var(--line);
  border-radius: var(--button-radius);
  background: var(--panel);
  color: var(--ink);
  padding: 5px 10px;
  font-size: 12px;
  font-weight: 850;
  white-space: nowrap;
}

.gc-topnav b,
.gc-chip.muted,
.gc-pagination a.disabled {
  color: var(--muted);
}

.gc-topnav a:hover,
.gc-sortbar a.active,
.gc-pagination a.active,
.gc-button.primary,
.gc-chip.primary,
.gc-toolbar a.active {
  border-color: var(--accent);
  background: var(--accent);
  color: ${v3 ? "#07130a" : v4 ? "#ffffff" : v5 ? "#090c12" : "#ffffff"};
}

.gc-mobile-nav {
  display: none;
}

.gc-main {
  min-width: 0;
  padding: 24px;
}

.gc-page {
  display: grid;
  gap: 14px;
}

.gc-sticky {
  position: sticky;
  top: 70px;
  z-index: 18;
  display: grid;
  gap: 8px;
  border-bottom: var(--border-width) solid var(--line);
  background: color-mix(in srgb, var(--bg) 94%, transparent);
  padding: 8px 0;
  backdrop-filter: blur(10px);
}

.gc-sortbar,
.gc-owned-left,
.gc-pagination,
.gc-filter-actions,
.gc-detail-actions,
.gc-tag-row,
.gc-card-foot,
.gc-rank-summary,
.gc-stat-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.gc-sortbar {
  gap: 0;
}

.gc-sortbar a {
  margin-left: -1px;
}

.gc-owned-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.gc-filter-panel,
.gc-panel,
.gc-data-head,
.gc-toolbar,
.gc-detail-hero,
.gc-state,
.gc-empty {
  border: var(--border-width) solid var(--line);
  border-radius: var(--radius);
  background: var(--panel);
  box-shadow: var(--shadow);
}

.gc-filter-panel,
.gc-panel,
.gc-data-head,
.gc-toolbar,
.gc-state,
.gc-empty {
  padding: 16px;
}

.gc-state.danger {
  border-color: var(--danger);
  color: var(--danger);
}

.gc-form-grid,
.gc-filter-panel {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
}

.gc-form-grid label,
.gc-filter-panel label {
  display: grid;
  gap: 5px;
  color: var(--muted);
  font-size: 12px;
  font-weight: 800;
}

.gc-form-grid input,
.gc-form-grid select,
.gc-filter-panel input,
.gc-filter-panel select {
  min-height: 38px;
  border: var(--border-width) solid var(--line);
  border-radius: var(--button-radius);
  background: var(--panel-2);
  color: var(--ink);
  padding: 6px 10px;
}

.gc-movie-grid {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 12px;
}

.gc-movie-card {
  display: block;
  ${cardBorder}
  border-radius: var(--radius);
  background: var(--panel);
  color: var(--ink);
  box-shadow: var(--card-shadow);
  ${cardPadding}
}

.gc-movie-card:hover {
  border-color: var(--accent-2);
}

.gc-poster {
  position: relative;
  display: block;
  overflow: hidden;
  border-radius: var(--poster-radius);
  background: var(--panel-2);
}

.gc-poster img,
.gc-poster span {
  display: grid;
  width: 100%;
  aspect-ratio: 2 / 3;
  place-items: center;
  border: var(--border-width) solid var(--line);
  border-radius: var(--poster-radius);
  background: var(--panel-2);
  object-fit: cover;
}

.gc-poster b {
  position: absolute;
  right: 8px;
  bottom: 8px;
  border: var(--border-width) solid var(--line);
  border-radius: var(--button-radius);
  background: var(--panel);
  color: var(--accent);
  padding: 5px 8px;
  font-weight: 900;
}

.gc-movie-body {
  display: grid;
  gap: 8px;
  padding: ${v3 ? "12px 12px 0" : "9px 0 0"};
}

.gc-movie-body h2 {
  margin: 0;
  color: var(--ink);
  font-family: var(--font-display);
  font-size: ${v4 ? "20px" : v3 ? "20px" : "15px"};
  font-weight: 900;
  line-height: 1.2;
}

.gc-movie-body h2 small {
  margin-left: 6px;
  color: var(--muted);
  font-family: var(--font-main);
  font-size: 12px;
}

.gc-movie-body p {
  margin: 0;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.55;
}

.gc-tag-row a {
  border: var(--border-width) solid var(--line);
  border-radius: var(--button-radius);
  background: var(--panel-2);
  color: var(--accent-2);
  padding: 3px 7px;
  font-size: 12px;
  font-weight: 800;
}

.gc-card-foot {
  border-top: var(--border-width) solid var(--line);
  padding-top: 7px;
  color: var(--muted);
  font-size: 12px;
}

.gc-card-foot a {
  color: var(--accent-2);
}

.gc-rank-summary div,
.gc-stat-row span,
.gc-kv-grid div {
  display: grid;
  gap: 4px;
  min-width: 112px;
  border: var(--border-width) solid var(--line);
  border-radius: var(--radius);
  background: var(--panel-2);
  padding: 8px 10px;
}

.gc-rank-summary strong,
.gc-stat-row b {
  color: var(--accent);
  font-size: 12px;
}

.gc-rank-lines {
  display: grid;
  gap: 4px;
  color: var(--muted);
  font-size: 12px;
}

.gc-detail-hero {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 20px;
  padding: 18px;
}

.gc-detail-poster img,
.gc-detail-poster span {
  display: grid;
  width: 100%;
  min-height: 450px;
  place-items: center;
  border: var(--border-width) solid var(--line);
  border-radius: var(--poster-radius);
  background: var(--panel-2);
  object-fit: cover;
}

.gc-detail-main {
  display: grid;
  align-content: start;
  gap: 14px;
}

.gc-detail-main h1,
.gc-data-head h1,
.gc-cast-search h1 {
  margin: 0;
  color: var(--ink);
  font-family: var(--font-display);
  font-size: ${v4 ? "58px" : "34px"};
  font-weight: 900;
  line-height: 1.03;
}

.gc-detail-main p,
.gc-data-head p {
  margin: 0;
  color: var(--muted);
  line-height: 1.7;
}

.gc-kv-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 10px;
}

.gc-kv-grid dd {
  margin: 0;
  overflow-wrap: anywhere;
}

.gc-data-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
}

.gc-toolbar {
  display: grid;
  gap: 12px;
}

.gc-toolbar section {
  display: grid;
  gap: 8px;
}

.gc-toolbar section div {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.gc-panel-title {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.gc-panel-title h2 {
  margin: 0;
  color: var(--ink);
  font-family: var(--font-display);
  font-size: ${v4 ? "28px" : "17px"};
}

.gc-table-wrap {
  overflow-x: auto;
  border: var(--border-width) solid var(--line);
  border-radius: var(--radius);
}

.gc-table-wrap table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12.5px;
}

.gc-table-wrap th,
.gc-table-wrap td {
  border-bottom: var(--border-width) solid var(--line);
  border-left: var(--border-width) solid var(--line);
  padding: 8px 10px;
  text-align: left;
  vertical-align: top;
  white-space: nowrap;
}

.gc-table-wrap th:first-child,
.gc-table-wrap td:first-child {
  border-left: 0;
}

.gc-table-wrap th {
  background: var(--panel-2);
  color: var(--accent);
  font-weight: 900;
}

.gc-table-wrap pre {
  margin: 0;
  overflow: auto;
  white-space: pre-wrap;
}

.gc-empty {
  display: grid;
  gap: 12px;
  color: var(--muted);
}

@media (max-width: 1500px) {
  .gc-movie-grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}

@media (max-width: 1240px) {
  .gc-shell {
    grid-template-columns: 1fr;
  }

  .gc-sidebar {
    display: none;
  }

  .gc-mobile-nav {
    display: block;
    margin: 14px 14px 0;
  }

  .gc-mobile-nav summary {
    border: var(--border-width) solid var(--line);
    border-radius: var(--radius);
    background: var(--panel);
    color: var(--ink);
    padding: 12px;
    font-weight: 900;
  }

  .gc-mobile-nav div {
    display: grid;
    gap: 12px;
    border: var(--border-width) solid var(--line);
    border-top: 0;
    background: var(--panel);
    padding: 12px;
  }

  .gc-mobile-nav section {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .gc-mobile-nav h2 {
    width: 100%;
    margin: 0;
    color: var(--accent);
    font-size: 12px;
  }

  .gc-topnav {
    display: none;
  }

  .gc-movie-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (max-width: 920px) {
  .gc-main {
    padding: 16px 14px 40px;
  }

  .gc-owned-row,
  .gc-data-head,
  .gc-detail-hero {
    display: grid;
    grid-template-columns: 1fr;
  }

  .gc-movie-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 680px) {
  .gc-movie-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .gc-detail-main h1,
  .gc-data-head h1,
  .gc-cast-search h1 {
    font-size: 30px;
  }
}

@media (max-width: 460px) {
  .gc-movie-grid {
    grid-template-columns: 1fr;
  }
}
`;
}

function navLiteral(groups) {
  const mapped = groups.map((group) => ({
    label: group.label,
    items: group.items.map(([itemPath, label]) => ({path: itemPath, label, match: itemPath === "/cards" ? ["/cards"] : undefined})),
  }));
  return JSON.stringify(mapped, null, 2);
}

function arrayLiteral(value) {
  return JSON.stringify(value, null, 2) + " as Array<[string, string, string]>";
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
