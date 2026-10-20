import assert from "node:assert/strict";
import { test } from "vitest";
import { createMemoryHistory, createRouter } from "vue-router";
import { sidebarGroups, topNavGroups } from "@/app/navigation";
import { crawlerRoutes } from "@/features/crawler/routes";
import { fc2Routes } from "@/features/fc2/routes";
import { mediaRoutes } from "@/features/media/routes";
import { movieRoutes } from "@/features/movies/routes";
import { remoteDailyBestRoutes } from "@/features/remote-dailybest/routes";
import { scRoutes } from "@/features/sc/routes";
import { shRoutes } from "@/features/sh/routes";

const routes = [
  ...movieRoutes,
  ...crawlerRoutes,
  ...shRoutes,
  ...scRoutes,
  ...mediaRoutes,
  ...fc2Routes,
  ...remoteDailyBestRoutes
];

test("周期卡片使用两个独立页面、API 和导航入口，删除旧统一路由", () => {
  const birth = movieRoutes.find((route) => route.path === "/listcardperiodbirth");
  const release = movieRoutes.find((route) => route.path === "/listcardperiodrelease");
  assert.equal(birth?.meta?.apiPath, "/cardsperiodbirth");
  assert.equal(birth?.meta?.title, "下载周期卡片");
  assert.equal(release?.meta?.apiPath, "/cardsperiodrelease");
  assert.equal(release?.meta?.title, "上映周期卡片");
  assert.notEqual(birth?.component, release?.component);
  assert.equal(
    movieRoutes.some((route) => route.path === "/listcardperiod"),
    false
  );
  for (const groups of [sidebarGroups, topNavGroups]) {
    const links = groups.flatMap((group) => group.links);
    for (const route of [birth, release]) {
      assert.equal(
        links.some((link) => link.href === route?.path && link.label === route?.meta?.title),
        true
      );
    }
    assert.equal(
      links.some((link) => link.href === "/listcardperiod"),
      false
    );
  }
});

test("旧聚合页面及导航入口删除，时间桶、周期卡片与回填入口保留", () => {
  for (const path of ["/aggbirth", "/aggrelease"]) {
    assert.equal(
      routes.some((route) => route.path === path),
      false
    );
    for (const groups of [sidebarGroups, topNavGroups]) {
      assert.equal(
        groups.flatMap((group) => group.links).some((link) => link.href === path),
        false
      );
    }
  }
  for (const path of [
    "/bucketsbirth",
    "/bucketsrelease",
    "/listcardperiodbirth",
    "/listcardperiodrelease",
    "/aggbackfill"
  ]) {
    assert.equal(
      routes.some((route) => route.path === path),
      true,
      path
    );
  }
});

test("all feature routes have a unique name, title, api contract and lazy component", () => {
  const names = routes.map((route) => route.name);
  assert.equal(new Set(names).size, names.length);
  for (const route of routes) {
    assert.equal(typeof route.name, "string");
    assert.equal(typeof route.meta?.title, "string");
    assert.equal(typeof route.meta?.apiPath, "string");
    assert.equal(typeof route.component, "function");
  }
});

test("the complete existing URL contract remains available", () => {
  const paths = new Set(routes.map((route) => route.path));
  for (const path of [
    "/listcardtoday",
    "/moviedetail/:movieName",
    "/moviealbumlist",
    "/dailybestfetch",
    "/remotedailybestrankdaystats",
    "/remotedailybestfetchrunlist",
    "/shpool",
    "/shthread/:threadTid",
    "/scevents",
    "/scoptionlist",
    "/rankdaystats",
    "/rankperiodstats",
    "/scpick",
    "/mediarescan",
    "/fclistcard",
    "/fcdetail/:fc2Key"
  ])
    assert.equal(paths.has(path), true, path);
});

test("SC 抽取使用 scpick 路由与导航，旧 V2 路径匹配 404 页面", () => {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [...routes, { path: "/:pathMatch(.*)*", name: "not-found", component: {} }]
  });
  assert.equal(router.resolve("/scpick").name, "sc-pick");
  assert.equal(router.resolve("/scpick").meta.title, "SC Pick");
  assert.equal(router.resolve("/scpickv2").name, "not-found");
  const sidebarLinks = sidebarGroups.flatMap((group) => group.links);
  assert.equal(
    sidebarLinks.some((link) => link.href === "/scpick" && link.label === "ScPick"),
    true
  );
  for (const groups of [sidebarGroups, topNavGroups]) {
    assert.equal(
      groups.flatMap((group) => group.links).some((link) => link.href === "/scpickv2"),
      false
    );
  }
});
