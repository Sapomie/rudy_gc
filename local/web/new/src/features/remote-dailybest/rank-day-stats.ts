import type { RankDayComparison, RankDayStat, RankDaySyncTarget } from "@/features/remote-dailybest/types";

export interface RankDayStatsRow extends RankDayStat {
  comparison?: RankDayComparison;
}

export function mergeRankDayStats(stats: RankDayStat[], comparisons: RankDayComparison[]): RankDayStatsRow[] {
  const comparisonsByDay = new Map(comparisons.map((row) => [row.day_number, row]));
  return stats.map((row) => ({ ...row, comparison: comparisonsByDay.get(row.day_number) }));
}

export function rankDayNeedsSync(row: RankDayStatsRow): boolean {
  return row.comparison?.needs_sync === true;
}

export function collectRankDaySyncTargets(rows: RankDayStatsRow[], selectedDayNumbers: number[]): RankDaySyncTarget[] {
  const selected = new Set(selectedDayNumbers);
  const targets: RankDaySyncTarget[] = [];
  for (const row of rows) {
    if (!selected.has(row.day_number)) continue;
    for (const category of row.comparison?.categories || []) {
      if (!category.needs_sync) continue;
      targets.push({ day_number: row.day_number, category: category.remote.category });
    }
  }
  return targets;
}
