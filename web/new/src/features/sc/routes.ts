import { defineFeatureRoutes } from "@/shared/router/routes";

export const scRoutes = defineFeatureRoutes([
  {
    path: "/scevents",
    name: "sc-events",
    component: () => import("./pages/ScEventsPage.vue"),
    title: "SC Events",
    apiPath: "/sc-events"
  },
  {
    path: "/sccards",
    name: "sc-event-cards",
    component: () => import("./pages/ScEventsPage.vue"),
    title: "SC Event Cards",
    apiPath: "/sc-events-cards"
  },
  {
    path: "/scevent/:eventName",
    name: "sc-event-detail",
    component: () => import("./pages/ScEventsPage.vue"),
    title: "SC Event Detail",
    apiPath: "/sc-events/:eventName"
  },
  {
    path: "/scpick",
    name: "sc-pick-smart",
    component: () => import("./pages/ScPickSmartPage.vue"),
    title: "SC Pick Smart Media",
    apiPath: "/sc-pick-smart-media"
  },
  {
    path: "/scadd",
    name: "sc-triggers",
    component: () => import("./pages/ScTriggersPage.vue"),
    title: "SC Triggers",
    apiPath: "/triggers/sc"
  },
  {
    path: "/scmove",
    name: "sc-media-move",
    component: () => import("./pages/ScMediaMovePage.vue"),
    title: "SC Media 移动",
    apiPath: "/triggers/sc-media"
  }
]);
