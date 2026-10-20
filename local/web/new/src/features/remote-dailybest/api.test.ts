import { afterEach, describe, expect, it, vi } from "vitest";
import { listLocalRankDayComparisons, startFetch, startTargetProbe, syncLocalRankDays } from "./api";
import { formatRankDaySyncStatus, rankDaySyncStatusTone } from "./format";

describe("remote dailybest api", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("通过远端代理触发抓取", async () => {
    const fetchMock = vi.fn().mockImplementation(() => Promise.resolve(new Response("{}", { status: 200 })));
    vi.stubGlobal("fetch", fetchMock);

    await startFetch();

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/remote-api/api/v1/dailybest/fetch",
      expect.objectContaining({ method: "POST", credentials: "same-origin" })
    );
  });

  it("通过远端代理发起目标站探测", async () => {
    const fetchMock = vi.fn().mockImplementation(() => Promise.resolve(new Response("{}", { status: 201 })));
    vi.stubGlobal("fetch", fetchMock);

    await startTargetProbe();

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/remote-api/api/v1/dailybest/target-probes",
      expect.objectContaining({ method: "POST", credentials: "same-origin" })
    );
  });

  it("通过本地接口读取排名日对比", async () => {
    const response = {
      Days: [],
      Total: 0,
      Query: {
        BusinessDateStart: "2026-07-01",
        BusinessDateEnd: "2026-07-20",
        Page: 1,
        PageSize: 24
      },
      PageInfo: { Total: 0, Page: 1, PageSize: 24, PageTotal: 1 }
    };
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(response), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      listLocalRankDayComparisons({
        business_date_start: "2026-07-01",
        business_date_end: "2026-07-20",
        p: 1,
        ps: 24
      })
    ).resolves.toEqual(response);

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/dailybest/rank-days?business_date_start=2026-07-01&business_date_end=2026-07-20&p=1&ps=24",
      expect.objectContaining({ credentials: "same-origin" })
    );
  });

  it("通过本地接口提交选中日期的待同步分类", async () => {
    const response = { job_id: "job-1", task_type: "SpiderDailyBestSync", message: "任务已启动" };
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify(response), { status: 202 }));
    vi.stubGlobal("fetch", fetchMock);
    const targets = [
      { day_number: 1304, category: 1 },
      { day_number: 1305, category: 2 }
    ];

    await expect(syncLocalRankDays(targets)).resolves.toEqual(response);

    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      "/api/dailybest/rank-days/sync",
      expect.objectContaining({
        method: "POST",
        credentials: "same-origin",
        body: JSON.stringify({ targets })
      })
    );
  });
});

describe("rank day sync status format", () => {
  it.each([
    ["missing", "本地缺失", "warn"],
    ["incomplete", "数量不完整", "warn"],
    ["version_mismatch", "版本不一致", "warn"],
    ["pending", "待处理", "info"],
    ["failed", "处理失败", "danger"],
    ["completed", "已完成", "good"]
  ])("格式化 %s 状态", (status, label, tone) => {
    expect(formatRankDaySyncStatus(status)).toBe(label);
    expect(rankDaySyncStatusTone(status)).toBe(tone);
  });

  it("保留未知状态并使用中性色", () => {
    expect(formatRankDaySyncStatus("custom")).toBe("custom");
    expect(rankDaySyncStatusTone("custom")).toBe("neutral");
  });
});
