import assert from "node:assert/strict";
import { test } from "vitest";
import { crawlerRoutes } from "@/features/crawler/routes";
import { fc2Routes } from "@/features/fc2/routes";
import { mediaRoutes } from "@/features/media/routes";
import { movieRoutes } from "@/features/movies/routes";
import { scRoutes } from "@/features/sc/routes";
import { shRoutes } from "@/features/sh/routes";

const routes = [...movieRoutes, ...crawlerRoutes, ...shRoutes, ...scRoutes, ...mediaRoutes, ...fc2Routes];

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
    "/shpool",
    "/shthread/:threadTid",
    "/scevents",
    "/scpick",
    "/mediarescan",
    "/fclistcard",
    "/fcdetail/:fc2Key"
  ])
    assert.equal(paths.has(path), true, path);
});
