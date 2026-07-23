import { afterEach, describe, expect, it, vi } from "vitest";
import {
  getFetchRun,
  listFetchLogs,
  listFetchRuns,
  listRankDayStats,
  listRanks,
  listRuntimeEvents,
  listTargetProbes,
  loadHealthState,
  loadRuntimeStatus,
  startFetch
} from "./api";

describe("dailybest api", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("returns ready when health and readiness checks succeed", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true }));
    await expect(loadHealthState()).resolves.toBe("ready");
  });

  it("uses the runtime monitoring paths", async () => {
    const fetchMock = responseFetch("{}");
    await loadRuntimeStatus();
    await listRuntimeEvents({ component: "target", level: "error", p: 2, ps: 24 });
    await listTargetProbes({ status: "failed", started_from: 100, started_to: 200, p: 3, ps: 24 });
    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/dailybest/runtime-status", expect.any(Object));
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/dailybest/runtime-events?component=target&level=error&p=2&ps=24",
      expect.any(Object)
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      3,
      "/api/v1/dailybest/target-probes?status=failed&started_from=100&started_to=200&p=3&ps=24",
      expect.any(Object)
    );
  });

  it("builds rank table requests with independent filters", async () => {
    const fetchMock = responseFetch("{}");
    await listRanks({ business_date: "2026-07-20", category: 1, rank_pos: 2, movie_jav_id: "abc", p: 2, ps: 12 });
    await listRankDayStats({ rank_date: "2026-07-20", p: 1, ps: 24 });
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
  });

  it("uses the public dailybest paths for fetch operations", async () => {
    const fetchMock = responseFetch("{}");
    await listFetchRuns({ p: 2, ps: 24 });
    await listFetchLogs({ fetch_run_id: 12, page_number: 3, status: "failed", p: 2, ps: 24 });
    await getFetchRun(12);
    await startFetch();
    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/v1/dailybest/fetch-runs?p=2&ps=24", expect.any(Object));
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/dailybest/fetch-logs?fetch_run_id=12&page_number=3&status=failed&p=2&ps=24",
      expect.any(Object)
    );
    expect(fetchMock).toHaveBeenNthCalledWith(3, "/api/v1/dailybest/fetch-runs/12", expect.any(Object));
    expect(fetchMock).toHaveBeenNthCalledWith(4, "/api/v1/dailybest/fetch", {
      method: "POST",
      headers: { Accept: "application/json" },
      body: undefined
    });
  });
});

function responseFetch(body: string) {
  const fetchMock = vi.fn().mockResolvedValue({ ok: true, text: vi.fn().mockResolvedValue(body) });
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}
