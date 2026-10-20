import assert from "node:assert/strict";
import { test } from "vitest";
import { useTriggerForm } from "./useTriggerForm";

test("trigger form keeps the existing business defaults", () => {
  const form = useTriggerForm();
  assert.equal(form.number, 1000000);
  assert.equal(form.startPage, 0);
  assert.equal(form.fetchMode, "search");
  assert.equal(form.existingPolicy, "skip");
  assert.equal(form.detailLimit, undefined);
});
