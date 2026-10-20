import assert from "node:assert/strict";
import { test } from "vitest";
import { normalizeScDuration } from "./scDuration.ts";

test("normalizeScDuration handles numeric input", () => {
  assert.equal(normalizeScDuration(8), 8);
});

test("normalizeScDuration converts an empty input to zero", () => {
  assert.equal(normalizeScDuration(""), 0);
});

test("normalizeScDuration keeps the existing integer-minute behavior", () => {
  assert.equal(normalizeScDuration(8.9), 8);
});

test("normalizeScDuration rejects a non-finite number", () => {
  assert.throws(() => normalizeScDuration(Number.NaN), /时长\(分钟\)必须是数字/);
});
