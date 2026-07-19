import { defineFeatureRoutes } from "@/shared/router/routes";

const cardRoutes = [
  { path: "/fclistcard", name: "fc2-cards", apiPath: "/fc2/cards", title: "FC2 Cards" },
  { path: "/fclistcardowned", name: "fc2-cards-owned", apiPath: "/fc2/cards/owned", title: "FC2 Cards MediaOwned" },
  { path: "/fclistcardrandom", name: "fc2-cards-random", apiPath: "/fc2/cards/random", title: "FC2 Cards Random" }
].map((route) => ({ ...route, component: () => import("./pages/FC2CardsPage.vue") }));

export const fc2Routes = defineFeatureRoutes([
  ...cardRoutes,
  {
    path: "/fclist",
    name: "fc2-movies",
    component: () => import("./pages/FC2Page.vue"),
    title: "FC2影片",
    apiPath: "/fc2/movies"
  },
  {
    path: "/fcdetail/:fc2Key",
    name: "fc2-detail",
    component: () => import("./pages/FC2DetailPage.vue"),
    title: "FC2影片详情",
    apiPath: "/fc2/movies/:fc2Key"
  },
  {
    path: "/fcingest",
    name: "fc2-media-ingest",
    component: () => import("./pages/FC2MediaIngestPage.vue"),
    title: "FC2媒体入库",
    apiPath: "/triggers/fc2-media"
  }
]);
