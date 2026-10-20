import { defineFeatureRoutes } from "@/shared/router/routes";

export const shRoutes = defineFeatureRoutes([
  {
    path: "/shfetchall",
    name: "sh-fetch-all",
    component: () => import("@/features/crawler/pages/TriggerPage.vue"),
    title: "SH 一体抓取",
    apiPath: "/triggers/fetch-sehuatang-sh"
  },
  {
    path: "/shfetch",
    name: "sh-fetch-list",
    component: () => import("@/features/crawler/pages/TriggerPage.vue"),
    title: "SH 列表建池",
    apiPath: "/triggers/fetch-sehuatang-sh-thread-list"
  },
  {
    path: "/shfill",
    name: "sh-fill-detail",
    component: () => import("@/features/crawler/pages/TriggerPage.vue"),
    title: "SH 详情补全",
    apiPath: "/triggers/fetch-sehuatang-sh-thread-detail"
  },
  {
    path: "/shpool",
    name: "sh-pool",
    component: () => import("./pages/SHPoolPage.vue"),
    title: "SH 帖子池",
    apiPath: "/fetch-site-sehuatang-sh-list"
  },
  {
    path: "/shthread/:threadTid",
    name: "sh-thread-detail",
    component: () => import("./pages/SHThreadDetailPage.vue"),
    title: "SH 帖子详情",
    apiPath: "/fetch-site-sehuatang-sh-detail/:threadTid"
  },
  {
    path: "/shmagnetlist",
    name: "sh-magnets",
    component: () => import("./pages/SHMagnetPage.vue"),
    title: "SH 磁力",
    apiPath: "/sh-magnets"
  },
  {
    path: "/shbucketlist",
    name: "sh-buckets",
    component: () => import("./pages/SHBucketListPage.vue"),
    title: "SH 聚合桶",
    apiPath: "/sh-thread-agg/bucket-list"
  }
]);
