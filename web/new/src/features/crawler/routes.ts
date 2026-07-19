import { defineFeatureRoutes } from "@/shared/router/routes";

const listRoutes = [
  { path: "/javbuslist", name: "crawler-javbus-list", apiPath: "/fetch-site-javbus-list", title: "JavBus List" },
  { path: "/sukebeilist", name: "crawler-sukebei-list", apiPath: "/fetch-site-sukebei-list", title: "Sukebei List" }
];

const triggerRoutes = [
  { path: "/dailybestfetch", name: "crawler-daily-best", apiPath: "/triggers/dailybest", title: "DailyBest" },
  { path: "/seedfetch", name: "crawler-seeds", apiPath: "/triggers/seeds", title: "Seeds" },
  { path: "/postprocess", name: "crawler-post-process", apiPath: "/triggers/post-process", title: "PostProcess" },
  { path: "/databackfill", name: "crawler-backfill", apiPath: "/triggers/backfill", title: "Backfill" },
  { path: "/sitefetch", name: "crawler-fetch-site", apiPath: "/triggers/fetch-site", title: "FetchSite" },
  {
    path: "/javbusfetch",
    name: "crawler-javbus-fetch",
    apiPath: "/triggers/fetch-site-javbus-filtered",
    title: "JavBus Filtered"
  },
  {
    path: "/sukebeifetch",
    name: "crawler-sukebei-fetch",
    apiPath: "/triggers/fetch-site-sukebei-filtered",
    title: "Sukebei Filtered"
  }
].map((route) => ({ ...route, component: () => import("./pages/TriggerPage.vue") }));

export const crawlerRoutes = defineFeatureRoutes([
  { ...listRoutes[0], component: () => import("./pages/JavbusListPage.vue") },
  { ...listRoutes[1], component: () => import("./pages/SukebeiListPage.vue") },
  ...triggerRoutes,
  {
    path: "/aggbackfill",
    name: "crawler-aggregate-backfill",
    component: () => import("./pages/AggTriggersPage.vue"),
    title: "聚合回填",
    apiPath: "/triggers/agg"
  },
  {
    path: "/crawlrecordlist",
    name: "crawler-records",
    component: () => import("./pages/RecordsPage.vue"),
    title: "Crawl Records",
    apiPath: "/crawl-records"
  },
  {
    path: "/seedlist",
    name: "crawler-seed-data",
    component: () => import("./pages/DSeedsPage.vue"),
    title: "Seed Data",
    apiPath: "/d-seeds"
  },
  {
    path: "/eitemlist",
    name: "crawler-e-items",
    component: () => import("./pages/EItemsPage.vue"),
    title: "EItems",
    apiPath: "/e-items"
  },
  {
    path: "/siteconfiglist",
    name: "crawler-sites",
    component: () => import("./pages/FetchSitesPage.vue"),
    title: "FetchSites",
    apiPath: "/fetch-sites"
  },
  {
    path: "/crawltasklist",
    name: "crawler-tasks",
    component: () => import("./pages/CrawlerTasksPage.vue"),
    title: "Crawler Tasks",
    apiPath: "/crawler/tasks"
  },
  {
    path: "/detailfetch",
    name: "crawler-detail-loop",
    component: () => import("./pages/CrawlerDetailLoopPage.vue"),
    title: "Detail Loop",
    apiPath: "/crawler/detail-loop/page"
  }
]);
