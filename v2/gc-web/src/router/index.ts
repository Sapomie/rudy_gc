import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";

import { cardRoutes } from "@/data/routes";

const cardsChildren: RouteRecordRaw[] = cardRoutes.map((item) => ({
  path: item.path.replace(/^\//, ""),
  name: item.key,
  meta: { title: item.title, cardView: item.key },
  component: () => import("@/pages/cards/index.vue"),
}));

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    component: () => import("@/layouts/AppLayout.vue"),
    children: [
      { path: "", redirect: "/cards" },
      ...cardsChildren,
      {
        path: "movie/:movie",
        name: "movie-detail",
        component: () => import("@/pages/movie/index.vue"),
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
