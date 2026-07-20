import { describe, expect, it } from "vitest";
import { buildPaginationItems, compactPageQuery, pageQueryFromRoute } from "./pagination";

describe("pagination utils", () => {
  it("reads positive page values and falls back to list defaults", () => {
    expect(pageQueryFromRoute({ p: "3", ps: "48" })).toEqual({ p: 3, ps: 48 });
    expect(pageQueryFromRoute({ p: "bad", ps: "0" })).toEqual({ p: 1, ps: 24 });
  });

  it("removes default page values from the route query", () => {
    expect(compactPageQuery({ p: "5", ps: "12", status: "success" }, 1, 24)).toEqual({
      status: "success"
    });
    expect(compactPageQuery({ status: "success" }, 2, 48)).toEqual({
      status: "success",
      p: "2",
      ps: "48"
    });
  });

  it("builds the same start, window and end controls as local lists", () => {
    const items = buildPaginationItems({ total: 480, page: 10, page_size: 24, page_total: 20 });

    expect(items.map((item) => item.label)).toEqual([
      "Start",
      "Prev",
      "1",
      "...",
      "7",
      "8",
      "9",
      "10",
      "11",
      "12",
      "13",
      "...",
      "20",
      "Next",
      "End"
    ]);
    expect(items.find((item) => item.label === "10")?.active).toBe(true);
  });
});
