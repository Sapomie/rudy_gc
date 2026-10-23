import { describe, expect, it } from "vitest";
import {
  collectRankDaySyncTargets,
  mergeRankDayStats,
  rankDayNeedsSync
} from "@/features/remote-dailybest/rank-day-stats";
import type {
  LocalRankCategoryState,
  RankDayCategoryComparison,
  RankDayComparison,
  RankDayStat,
  RankDaySyncStatus
} from "@/features/remote-dailybest/types";

describe("remote dailybest rank day stats", () => {
  it("按日序号合并本地排名和状态，不依赖响应顺序", () => {
    const rows = mergeRankDayStats(
      [stat(1305), stat(1304)],
      [comparison(1304, 475, "completed", false), comparison(1305, 0, "missing", true)]
    );

    expect(rows.map((row) => row.comparison?.local_rank_count)).toEqual([0, 475]);
    expect(rows.map((row) => row.comparison?.status)).toEqual(["missing", "completed"]);
  });

  it("缺少本地对比数据时保持远端统计并禁用同步", () => {
    const [row] = mergeRankDayStats([stat(1305)], []);

    expect(row?.day_number).toBe(1305);
    expect(row?.comparison).toBeUndefined();
    expect(row && rankDayNeedsSync(row)).toBe(false);
  });

  it("只生成已选日期中分类级待同步目标", () => {
    const pending = comparison(1305, 0, "missing", true, [category(1, true), category(2, false)]);
    const completed = comparison(1304, 475, "completed", false, [category(1, false)]);
    const rows = mergeRankDayStats([stat(1305), stat(1304)], [pending, completed]);

    expect(rows.filter(rankDayNeedsSync).map((row) => row.day_number)).toEqual([1305]);
    expect(collectRankDaySyncTargets(rows, [1305, 1304])).toEqual([{ day_number: 1305, category: 1 }]);
  });
});

function stat(dayNumber: number): RankDayStat {
  return {
    id: dayNumber,
    day_number: dayNumber,
    rank_date: `2026-07-${dayNumber === 1305 ? "28" : "27"}`,
    prev_day_number: dayNumber - 1,
    rank_row_count: 500,
    movie_count: 500,
    continued_movie_count: 450,
    first_seen_movie_count: 40,
    returning_movie_count: 5,
    dropped_movie_count: 20,
    source_checksum: "checksum",
    fetched_at: 1,
    synced_at: 1,
    calculated_on: 1,
    created_on: 1,
    updated_on: 1
  };
}

function comparison(
  dayNumber: number,
  localRankCount: number,
  status: RankDaySyncStatus,
  needsSync: boolean,
  categories: RankDayCategoryComparison[] = [category(1, needsSync)]
): RankDayComparison {
  return {
    business_date: `2026-07-${dayNumber === 1305 ? "28" : "27"}`,
    day_number: dayNumber,
    page_count: 25,
    item_count: 500,
    remote_rank_count: 500,
    local_rank_count: localRankCount,
    finished_at: 1,
    status,
    needs_sync: needsSync,
    categories
  };
}

function category(categoryNumber: number, needsSync: boolean): RankDayCategoryComparison {
  return {
    remote: {
      fetch_run_id: 49,
      category: categoryNumber,
      page_count: 25,
      item_count: 500,
      rank_count: 500,
      checksum: "remote-checksum",
      finished_at: 1
    },
    local: localCategoryState(),
    status: needsSync ? "missing" : "completed",
    needs_sync: needsSync
  };
}

function localCategoryState(): LocalRankCategoryState {
  return {
    rank_count: 0,
    checksum: "",
    remote_fetch_run_id: 0,
    sync_status: "",
    process_status: "",
    synced_at: 0,
    processed_at: 0,
    last_error: ""
  };
}
