import assert from "node:assert/strict";
import { afterEach, test, vi } from "vitest";
import { ApiError, getJSON, sendJSON } from "./request";

afterEach(() => vi.unstubAllGlobals());

test("getJSON parses a successful JSON response", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response('{"items":[1]}', { status: 200 })));
  assert.deepEqual(await getJSON("/api/items"), { items: [1] });
});

test("getJSON preserves a non-JSON error message and status", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response("service unavailable", { status: 503 })));
  await assert.rejects(
    () => getJSON("/api/items"),
    (error: unknown) => {
      assert.ok(error instanceof ApiError);
      assert.equal(error.status, 503);
      assert.equal(error.message, "service unavailable");
      return true;
    }
  );
});

test("sendJSON keeps method and payload contracts", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response('{"success":true}', { status: 200 }));
  vi.stubGlobal("fetch", fetchMock);
  await sendJSON("/api/items", "POST", { id: 7 });
  const [, init] = fetchMock.mock.calls[0] as [string, RequestInit];
  assert.equal(init.method, "POST");
  assert.equal(init.body, '{"id":7}');
  assert.equal((init.headers as Record<string, string>)["Content-Type"], "application/json");
});
