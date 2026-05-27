import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname, "..");
const router = readFileSync(resolve(root, "src/router/index.ts"), "utf8");
const routes = readFileSync(resolve(root, "src/data/routes.ts"), "utf8");

const failed = [];

const cardRoutes = [
  { path: "/cards", key: "cards" },
  { path: "/cardstoday", key: "cardstoday" },
  { path: "/cardshasrank", key: "cardshasrank" },
  { path: "/cardsmediamowned", key: "cardsmediamowned" },
  { path: "/cardsneeddownload", key: "cardsneeddownload" },
  { path: "/cardsrandom", key: "cardsrandom" },
];
const rankRoutes = [
  { path: "/moviecarddayrank", key: "rank-day" },
  { path: "/moviecardperiodrank", key: "rank-period" },
];
const genericListRoutes = [
  { path: "/casts", key: "casts" },
  { path: "/records", key: "records" },
  { path: "/crawl-records", key: "crawl-records" },
  { path: "/sc-events", key: "sc-events" },
  { path: "/medias", key: "medias" },
  { path: "/w-agg-events", key: "w-agg-events" },
  { path: "/e-items", key: "e-items" },
  { path: "/torrent-albums", key: "torrent-albums" },
  { path: "/movie-albums", key: "movie-albums" },
  { path: "/d-seeds", key: "d-seeds" },
  { path: "/fetch-site-javbus-list", key: "fetch-site-javbus-list" },
  { path: "/fetch-site-sukebei-list", key: "fetch-site-sukebei-list" },
  { path: "/fetch-site-sehuatang-list", key: "fetch-site-sehuatang-list" },
  { path: "/wdir", key: "wdir" },
  { path: "/w-media-agg/birth", key: "w-media-agg-birth" },
  { path: "/w-media-agg/bucket-list", key: "w-media-agg-bucket-list" },
  { path: "/movie-agg-all/release", key: "movie-agg-all-release" },
  { path: "/movie-agg-all/release-bucket-list", key: "movie-release-bucket-list" },
];

for (const item of cardRoutes) {
  if (!routes.includes(`path: "${item.path}"`) || !routes.includes(`key: "${item.key}"`)) {
    failed.push(`${item.path} route is missing from route data`);
  }
}
for (const item of rankRoutes) {
  if (!routes.includes(`path: "${item.path}"`) || !routes.includes(`key: "${item.key}"`)) {
    failed.push(`${item.path} route is missing from route data`);
  }
}
for (const item of genericListRoutes) {
  if (!routes.includes(`path: "${item.path}"`) || !routes.includes(`key: "${item.key}"`)) {
    failed.push(`${item.path} generic list route is missing from route data`);
  }
}
if (!router.includes('redirect: "/cards"')) {
  failed.push("root redirect to /cards is missing");
}
if (!router.includes('path: "movie/:movie"') || !router.includes('name: "movie-detail"') || !router.includes('@/pages/movie/index.vue')) {
  failed.push("movie detail route is missing");
}
if (!router.includes("genericListRoutes.map") || !router.includes("pageKey: item.key") || !router.includes('@/pages/generic/index.vue')) {
  failed.push("generic list route mapping is missing");
}
if (!router.includes('path: "cast/:id"') || !router.includes('pageKey: "cast-detail"')) {
  failed.push("cast detail generic route is missing");
}
for (const token of ["fixedRoutes.map", '@/pages/rank/day.vue', '@/pages/rank/period.vue']) {
  if (!router.includes(token)) {
    failed.push(`rank route mapping missing ${token}`);
  }
}
for (const token of ["cardRoutes.map", "cardView: item.key", '@/pages/cards/index.vue']) {
  if (!router.includes(token)) {
    failed.push(`cards route mapping missing ${token}`);
  }
}

if (failed.length > 0) {
  console.error(`Route checks failed:\n${failed.join("\n")}`);
  process.exit(1);
}

console.log("Verified cards, rank, movie detail, and generic list route scope.");
