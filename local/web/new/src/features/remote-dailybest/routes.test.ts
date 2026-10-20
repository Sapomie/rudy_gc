import { describe, expect, it } from "vitest";
import { remoteDailyBestRoutes } from "./routes";

describe("remote dailybest routes", () => {
  it("覆盖远端数据页面", () => {
    expect(remoteDailyBestRoutes.map((route) => route.path)).toEqual([
      "/remotedailybestsourceconfig",
      "/remotedailybestrankdaystats",
      "/remotedailybestfetchrunlist",
      "/remotedailybestfetchrundetail/:id",
      "/remotedailybestfetchloglist",
      "/remotedailybestfetch",
      "/remotedailybesttargetprobelist"
    ]);
  });

  it("每条路由都提供独立名称、标题和远端 API 契约", () => {
    for (const route of remoteDailyBestRoutes) {
      expect(String(route.name)).toMatch(/^remote-dailybest-/);
      expect(typeof route.meta?.title).toBe("string");
      expect(typeof route.meta?.apiPath).toBe("string");
      expect(typeof route.component).toBe("function");
    }
  });
});
