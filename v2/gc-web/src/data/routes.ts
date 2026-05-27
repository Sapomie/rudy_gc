export interface NavRoute {
  path: string;
  key: string;
  title: string;
  group: "Movie" | "Library" | "Ranks" | "People" | "Crawler" | "FetchSite" | "Media" | "Aggregation" | "SC";
  kind: "cards" | "rank" | "page" | "operation";
}

export const cardRoutes: NavRoute[] = [
  { path: "/cards", key: "cards", title: "Cards", group: "Movie", kind: "cards" },
  { path: "/cardstoday", key: "cardstoday", title: "Today", group: "Movie", kind: "cards" },
  { path: "/cardshasrank", key: "cardshasrank", title: "Ranked", group: "Movie", kind: "cards" },
  { path: "/cardsmediamowned", key: "cardsmediamowned", title: "Media Owned", group: "Movie", kind: "cards" },
  { path: "/cardsneeddownload", key: "cardsneeddownload", title: "Need Download", group: "Movie", kind: "cards" },
  { path: "/cardsrandom", key: "cardsrandom", title: "Random", group: "Movie", kind: "cards" },
];

export const fixedRoutes: NavRoute[] = [
  { path: "/moviecarddayrank", key: "rank-day", title: "Day Rank", group: "Ranks", kind: "rank" },
  { path: "/moviecardperiodrank", key: "rank-period", title: "Period Rank", group: "Ranks", kind: "rank" },
];

export const genericRoutes: NavRoute[] = [
  { path: "/casts", key: "casts", title: "Casts", group: "People", kind: "page" },
  { path: "/records", key: "records", title: "Records", group: "Crawler", kind: "page" },
  { path: "/crawl-records", key: "crawl-records", title: "Crawl Records", group: "Crawler", kind: "page" },
  { path: "/sc-events", key: "sc-events", title: "SC Events", group: "SC", kind: "page" },
  { path: "/medias", key: "medias", title: "Medias", group: "Media", kind: "page" },
  { path: "/w-agg-events", key: "w-agg-events", title: "WAgg Events", group: "Aggregation", kind: "page" },
  { path: "/e-items", key: "e-items", title: "E Items", group: "Library", kind: "page" },
  { path: "/torrent-albums", key: "torrent-albums", title: "Torrent Albums", group: "Library", kind: "page" },
  { path: "/movie-albums", key: "movie-albums", title: "Movie Albums", group: "Library", kind: "page" },
  { path: "/d-seeds", key: "d-seeds", title: "D Seeds", group: "Crawler", kind: "page" },
  { path: "/fetch-site-javbus-list", key: "fetch-site-javbus-list", title: "JavBus Fetch", group: "FetchSite", kind: "page" },
  { path: "/fetch-site-sukebei-list", key: "fetch-site-sukebei-list", title: "Sukebei Fetch", group: "FetchSite", kind: "page" },
  { path: "/fetch-site-sehuatang-list", key: "fetch-site-sehuatang-list", title: "Sehuatang", group: "FetchSite", kind: "page" },
  { path: "/wdir", key: "wdir", title: "W Directory", group: "Aggregation", kind: "page" },
  { path: "/w-media-agg/birth", key: "w-media-agg-birth", title: "WMedia Birth", group: "Aggregation", kind: "page" },
  { path: "/w-media-agg/bucket-list", key: "w-media-agg-bucket-list", title: "WMedia Buckets", group: "Aggregation", kind: "page" },
  { path: "/movie-agg-all/release", key: "movie-agg-all-release", title: "Release Agg", group: "Aggregation", kind: "page" },
  { path: "/movie-agg-all/release-bucket-list", key: "movie-release-bucket-list", title: "Release Buckets", group: "Aggregation", kind: "page" },
  { path: "/crawler/tasks", key: "crawler-tasks", title: "Crawler Tasks", group: "Crawler", kind: "operation" },
  { path: "/triggers", key: "triggers", title: "Crawler Jobs", group: "Crawler", kind: "operation" },
  { path: "/crawler/detail-loop", key: "crawler-detail-loop", title: "Detail Loop", group: "Crawler", kind: "operation" },
  { path: "/triggers/dailybest", key: "triggers-dailybest", title: "DailyBest", group: "Crawler", kind: "operation" },
  { path: "/triggers/seeds", key: "triggers-seeds", title: "Seeds", group: "Crawler", kind: "operation" },
  { path: "/triggers/post-process", key: "triggers-post-process", title: "Post Process", group: "Crawler", kind: "operation" },
  { path: "/triggers/backfill", key: "triggers-backfill", title: "Backfill", group: "Crawler", kind: "operation" },
  { path: "/triggers/fetch-site", key: "triggers-fetch-site", title: "FetchSite Trigger", group: "FetchSite", kind: "operation" },
  { path: "/triggers/fetch-sehuatang", key: "triggers-fetch-sehuatang", title: "FetchSehuatang", group: "FetchSite", kind: "operation" },
  { path: "/triggers/media", key: "triggers-media", title: "Media Ingest", group: "Media", kind: "operation" },
  { path: "/triggers/media-rescan", key: "triggers-media-rescan", title: "Media Rescan", group: "Media", kind: "operation" },
  { path: "/triggers/media-rollback", key: "triggers-media-rollback", title: "Media Rollback", group: "Media", kind: "operation" },
  { path: "/triggers/sc-media", key: "triggers-sc-media", title: "SC Media Move", group: "SC", kind: "operation" },
  { path: "/sc-pick-smart-media", key: "sc-pick-smart-media", title: "SC Smart Pick", group: "SC", kind: "operation" },
  { path: "/triggers/sc", key: "triggers-sc", title: "SC Preview", group: "SC", kind: "operation" },
  { path: "/triggers/agg", key: "triggers-agg", title: "Agg Triggers", group: "Aggregation", kind: "operation" },
];

export const genericListRoutes = genericRoutes.filter((item) => item.kind === "page");

export const operationRoutes = genericRoutes.filter((item) => item.kind === "operation");

export const allNavRoutes = [...cardRoutes, ...fixedRoutes, ...genericListRoutes];

export const quickNav = [
  ...cardRoutes,
  ...fixedRoutes,
];
