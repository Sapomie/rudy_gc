import assert from "node:assert/strict";
import { test } from "vitest";
import { rankDateFromDayNumber, rankPeriodDateRange } from "@/features/movies/components/stats/rankStats";
import { compactRouteQuery } from "@/shared/utils/query";

test("rank day numbers include period ranges before the first stored rank day", () => {
  assert.equal(rankDateFromDayNumber(1), "2023-01-01");
  assert.equal(rankDateFromDayNumber(0), "-");
  assert.equal(rankDateFromDayNumber(0, false), "2022-12-31");
  assert.equal(rankDateFromDayNumber(-5, false), "2022-12-26");
  assert.equal(rankPeriodDateRange({ StartDayNumber: -5, EndDayNumber: 1 }), "2022-12-26 — 2023-01-01");
});

test("rank stat routes compact only default pagination and sorting", () => {
  assert.deepEqual(
    compactRouteQuery("/rankdaystats", {
      p: "1",
      ps: "24",
      sort: "day_number",
      order: "desc",
      start_date: "2026-07-01"
    }),
    { start_date: "2026-07-01" }
  );
  assert.deepEqual(
    compactRouteQuery("/rankperiodstats", {
      ps: "24",
      sort: "start_day_number",
      order: "desc",
      type: "month",
      category: "2",
      start_date: "2026-07-01"
    }),
    { start_date: "2026-07-01" }
  );
});
