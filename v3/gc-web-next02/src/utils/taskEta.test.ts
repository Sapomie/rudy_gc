import assert from "node:assert/strict";
import test from "node:test";
import { estimateRemainingSeconds, formatTaskEta } from "./taskEta.ts";

test("estimateRemainingSeconds calculates ETA from elapsed time and progress", () => {
  const remaining = estimateRemainingSeconds(1 * 3600 + 14 * 60 + 8, 448, 619);
  assert.equal(remaining, 1698);
  assert.equal(formatTaskEta(remaining), "00:28:18");
});

test("estimateRemainingSeconds waits for the first completed item", () => {
  assert.equal(estimateRemainingSeconds(120, 0, 619), null);
  assert.equal(formatTaskEta(null), "计算中");
});

test("estimateRemainingSeconds returns zero for completed work", () => {
  assert.equal(estimateRemainingSeconds(120, 619, 619), 0);
  assert.equal(formatTaskEta(0), "00:00:00");
});

test("estimateRemainingSeconds rejects invalid progress data", () => {
  assert.equal(estimateRemainingSeconds(Number.NaN, 1, 2), null);
  assert.equal(estimateRemainingSeconds(120, 1, 0), null);
});
