import { defineFeatureRoutes } from "@/shared/router/routes";

export const mediaRoutes = defineFeatureRoutes([
  {
    path: "/mediadir",
    name: "media-directory",
    component: () => import("./pages/WDirPage.vue"),
    title: "Media Directory",
    apiPath: "/wdir"
  },
  {
    path: "/medialist",
    name: "media-list",
    component: () => import("./pages/MediasPage.vue"),
    title: "Media List",
    apiPath: "/medias"
  },
  {
    path: "/aggeventlist",
    name: "media-aggregate-events",
    component: () => import("./pages/WAggEventsPage.vue"),
    title: "W Agg Events",
    apiPath: "/w-agg-events"
  },
  {
    path: "/mediaingest",
    name: "media-ingest",
    component: () => import("./pages/MediaIngestPage.vue"),
    title: "媒体入库",
    apiPath: "/triggers/media"
  },
  {
    path: "/mediarescan",
    name: "media-rescan",
    component: () => import("./pages/MediaRescanPage.vue"),
    title: "媒体一致性检查",
    apiPath: "/triggers/media-rescan"
  },
  {
    path: "/mediarollback",
    name: "media-rollback",
    component: () => import("./pages/MediaRollbackPage.vue"),
    title: "媒体回滚",
    apiPath: "/triggers/media-rollback"
  }
]);
