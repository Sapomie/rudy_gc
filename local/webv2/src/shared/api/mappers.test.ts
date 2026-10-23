import assert from "node:assert/strict";
import { test } from "vitest";
import { mapActionResult, mapRecordPage, pickField, stringField, toRecord } from "./mappers";

test("record mapper accepts alternate response casing at the boundary", () => {
  assert.deepEqual(mapRecordPage({ Items: [{ ID: 1 }], Total: "3" }), { items: [{ ID: 1 }], total: 3 });
  assert.deepEqual(mapRecordPage({ rows: [{ id: 2 }], count: 1 }), { items: [{ id: 2 }], total: 1 });
});

test("action mapper normalizes success, message and error fields", () => {
  assert.deepEqual(mapActionResult({ Success: 1, Message: "ok", Error: "" }), {
    success: true,
    message: "ok",
    error: ""
  });
});

test("field helpers keep raw compatibility inside the mapper boundary", () => {
  const record = toRecord({ MovieName: "ABC-123" });
  assert.equal(pickField(record, "movie_name", "MovieName"), "ABC-123");
  assert.equal(stringField(record, "movie_name", "MovieName"), "ABC-123");
  assert.deepEqual(toRecord([]), {});
});
