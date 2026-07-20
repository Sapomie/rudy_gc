import { describe, expect, it } from "vitest";
import { sidebarGroups, topNavGroups } from "@/app/navigation";
import { dailyBestRoutes } from "./routes";

describe("dailybest routes", () => {
  it("covers every remote data view", () => {
    expect(dailyBestRoutes.map((route) => route.path)).toEqual([
      "/dailybestoverview",
      "/ranklist",
      "/rankdaystats",
      "/rankperiodlist",
      "/rankperioditemlist",
      "/rankperiodstats",
      "/dailybestfetchrunlist",
      "/dailybestfetchrundetail/:id",
      "/dailybestfetch"
    ]);
  });

  it("uses local-style semantic route names", () => {
    expect(dailyBestRoutes.map((route) => route.name)).toEqual([
      "dailybest-overview",
      "dailybest-rank-list",
      "dailybest-rank-day-stats",
      "dailybest-rank-period-list",
      "dailybest-rank-period-item-list",
      "dailybest-rank-period-stats",
      "dailybest-fetch-run-list",
      "dailybest-fetch-run-detail",
      "dailybest-fetch"
    ]);
  });

  it("keeps every navigation entry on a registered page route", () => {
    const registeredPaths = new Set(dailyBestRoutes.map((route) => route.path));
    for (const group of [...sidebarGroups, ...topNavGroups]) {
      for (const link of group.links) expect(registeredPaths.has(link.href)).toBe(true);
    }
  });
});
