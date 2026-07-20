import { afterEach, describe, expect, it, vi } from "vitest";
import { loadHealthState } from "./api";

describe("loadHealthState", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns ready when health and readiness checks succeed", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true }));
    await expect(loadHealthState()).resolves.toBe("ready");
  });

  it("returns unavailable when a readiness check fails", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValueOnce({ ok: true }).mockResolvedValueOnce({ ok: false }));
    await expect(loadHealthState()).resolves.toBe("unavailable");
  });
});
