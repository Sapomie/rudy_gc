import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import { dailyBestRoutes } from "@/features/dailybest/routes";

export const routes: RouteRecordRaw[] = [
  { path: "/", name: "home", redirect: "/dailybestoverview" },
  ...dailyBestRoutes,
  {
    path: "/:pathMatch(.*)*",
    name: "not-found",
    component: () => import("@/app/pages/NotFoundPage.vue"),
    meta: { title: "页面不存在" }
  }
];

export default createRouter({
  history: createWebHistory(),
  routes
});
