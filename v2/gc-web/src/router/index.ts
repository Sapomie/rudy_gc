import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";

import { cardRoutes, fixedRoutes, genericListRoutes } from "@/data/routes";

const cardsChildren: RouteRecordRaw[] = cardRoutes.map((item) => ({
  path: item.path.replace(/^\//, ""),
  name: item.key,
  meta: { title: item.title, cardView: item.key },
  component: () => import("@/pages/cards/index.vue"),
}));

const rankChildren: RouteRecordRaw[] = fixedRoutes.map((item) => ({
  path: item.path.replace(/^\//, ""),
  name: item.key,
  meta: { title: item.title },
  component: () => (item.key === "rank-day" ? import("@/pages/rank/day.vue") : import("@/pages/rank/period.vue")),
}));

const genericChildren: RouteRecordRaw[] = genericListRoutes.map((item) => ({
  path: item.path.replace(/^\//, ""),
  name: item.key,
  meta: { title: item.title, pageKey: item.key },
  component: () => import("@/pages/generic/index.vue"),
})).filter((item) => item.name !== "records");

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    component: () => import("@/layouts/AppLayout.vue"),
    children: [
      { path: "", redirect: "/cards" },
      ...cardsChildren,
      ...rankChildren,
      { path: "records", redirect: "/crawl-records" },
      ...genericChildren,
      {
        path: "movie/:movie",
        name: "movie-detail",
        component: () => import("@/pages/movie/index.vue"),
      },
      {
        path: "cast/:id",
        name: "cast-detail",
        meta: { title: "Cast Detail", pageKey: "cast-detail" },
        component: () => import("@/pages/generic/index.vue"),
      },
      {
        path: ":pathMatch(.*)*",
        name: "not-found",
        component: () => import("@/pages/notfound/index.vue"),
      },
    ],
  },
];

export default createRouter({
  history: createWebHistory(),
  routes,
});
