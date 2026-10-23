import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import { crawlerRoutes } from "@/features/crawler/routes";
import { fc2Routes } from "@/features/fc2/routes";
import { mediaRoutes } from "@/features/media/routes";
import { movieRoutes } from "@/features/movies/routes";
import { remoteDailyBestRoutes } from "@/features/remote-dailybest/routes";
import { scRoutes } from "@/features/sc/routes";
import { shRoutes } from "@/features/sh/routes";
import { styleLibraryRoutes } from "@/features/style-library/routes";
import { compactRouteQuery, type CompactQuery } from "@/shared/utils/query";

export const routes: RouteRecordRaw[] = [
  { path: "/", name: "home", redirect: "/listcardtoday" },
  ...movieRoutes,
  ...crawlerRoutes,
  ...shRoutes,
  ...scRoutes,
  ...mediaRoutes,
  ...fc2Routes,
  ...remoteDailyBestRoutes,
  ...styleLibraryRoutes,
  {
    path: "/:pathMatch(.*)*",
    name: "not-found",
    component: () => import("@/app/pages/NotFoundPage.vue"),
    meta: { title: "Not Found" }
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

export function querySignature(query: Record<string, unknown>): string {
  const params = new URLSearchParams();
  Object.keys(query)
    .sort()
    .forEach((key) => {
      const rawValues = Array.isArray(query[key]) ? (query[key] as unknown[]) : [query[key]];
      rawValues.forEach((value) => params.append(key, value === undefined || value === null ? "" : String(value)));
    });
  return params.toString();
}

router.beforeEach((to) => {
  const compacted = compactRouteQuery(to.path, to.query);
  if (querySignature(to.query) === querySignature(compacted as CompactQuery)) {
    return true;
  }
  return {
    path: to.path,
    query: compacted,
    hash: to.hash,
    replace: true
  };
});

router.afterEach((to) => {
  document.title = `${String(to.meta.title || "Rudy-GC")} - Rudy-GC`;
});

export default router;
