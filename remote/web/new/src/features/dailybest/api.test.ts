import { afterEach, describe, expect, it, vi } from "vitest";
import {
  getFetchRun,
  listFetchRuns,
  listRankDayStats,
  listRankPeriodItems,
  listRankPeriods,
  listRankPeriodStats,
  listRanks,
  listRankSyncManifests,
  loadHealthState,
  startFetch
} from "./api";

describe("dailybest api", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("returns ready when health and readiness checks succeed", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true }));
    await expect(loadHealthState()).resolves.toBe("ready");
  });

  it("requests rank sync manifests without an authorization header", async () => {
    const fetchMock = responseFetch('{"manifests":[]}');
    await listRankSyncManifests();
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/dailybest/rank-sync-manifests", {
      method: "GET",
      headers: { Accept: "application/json" },
      body: undefined
    });
  });

  it("builds all rank table requests with independent filters", async () => {
    const fetchMock = responseFetch("{}");
    await listRanks({ business_date: "2026-07-20", category: 1, rank_pos: 2, movie_jav_id: "abc", p: 2, ps: 12 });
    await listRankDayStats({ rank_date: "2026-07-20", p: 1, ps: 24 });
    await listRankPeriods({ period_type: 2, period_key: "2026-07", category: 1, p: 1, ps: 24 });
    await listRankPeriodItems({ period_id: 3, movie_jav_id: "abc", p: 1, ps: 24 });
    await listRankPeriodStats({ period_id: 3, p: 1, ps: 24 });
    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/v1/dailybest/ranks?business_date=2026-07-20&category=1&rank_pos=2&movie_jav_id=abc&p=2&ps=12",
      expect.any(Object)
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/dailybest/rank-day-stats?rank_date=2026-07-20&p=1&ps=24",
      expect.any(Object)
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      "/api/v1/dailybest/rank-periods?period_type=2&period_key=2026-07&category=1&p=1&ps=24",
      expect.any(Object)
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      4,
      "/api/v1/dailybest/rank-period-items?period_id=3&movie_jav_id=abc&p=1&ps=24",
      expect.any(Object)
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      5,
      "/api/v1/dailybest/rank-period-stats?period_id=3&p=1&ps=24",
      expect.any(Object)
    );
  });

  it("uses the public dailybest paths for fetch operations", async () => {
    const fetchMock = responseFetch("{}");
    await listFetchRuns({ p: 2, ps: 24 });
    await getFetchRun(12);
    await startFetch({ business_date: "", category: 1 });
    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/dailybest/fetch-runs?p=2&ps=24", expect.any(Object));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/v1/dailybest/fetch-runs/12", expect.any(Object));
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/dailybest/fetch", {
      method: "POST",
      headers: { Accept: "application/json", "Content-Type": "application/json" },
      body: JSON.stringify({ business_date: "", category: 1 })
    });
  });
});

function responseFetch(body: string) {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, text: vi.fn().mockResolvedValue(body) });
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}
