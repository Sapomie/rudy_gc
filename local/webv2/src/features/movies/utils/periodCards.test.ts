import assert from "node:assert/strict";
import { test } from "vitest";
import {
  periodCardsHref,
  periodListCardHref,
  selectCardPeriod,
  sortCardPeriods,
  switchCardPeriodType
} from "./periodCards";

const query = {
  basis: ["release"],
  period_type: ["month"],
  key: ["month:2020-12"],
  period_p: ["3"],
  period_ps: ["24"],
  period_sort: ["time"],
  period_order: ["desc"],
  p: ["4"],
  ps: ["18"],
  od: ["vw"],
  order: ["asc"]
};
const params = (href: string) => new URL(href, "http://localhost").searchParams;

test("选择周期只重置卡片页码并保留目录页码和两侧条数", () => {
  const result = params(selectCardPeriod("release", query, "month:2021-01"));
  assert.equal(result.get("key"), "month:2021-01");
  assert.equal(result.get("p"), "1");
  assert.equal(result.get("period_p"), "3");
  assert.equal(result.get("period_ps"), "24");
  assert.equal(result.get("ps"), "18");
  assert.equal(result.get("order"), "asc");
  assert.equal(result.get("od"), "vw");
});

test("两个页面切换周期类型清除旧周期并重置两侧页码", () => {
  for (const basis of ["release", "birth"] as const) {
    const href = new URL(switchCardPeriodType(basis, query, "quarter"), "http://localhost");
    const result = href.searchParams;
    assert.equal(href.pathname, `/listcardperiod${basis}`);
    assert.equal(result.has("basis"), false);
    assert.equal(result.get("period_type"), "quarter");
    assert.equal(result.has("key"), false);
    assert.equal(result.get("p"), "1");
    assert.equal(result.get("period_p"), "1");
    assert.equal(result.get("ps"), "18");
    assert.equal(result.get("od"), "vw");
    assert.equal(result.get("order"), "asc");
  }
});

test("页面路径固定周期依据，所有操作移除 URL 中的 basis", () => {
  for (const basis of ["release", "birth"] as const) {
    const tampered = { ...query, basis: [basis === "release" ? "birth" : "release"] };
    for (const href of [
      periodCardsHref(basis, {}, {}),
      periodCardsHref(basis, tampered, { basis: "rank", p: 2 }),
      selectCardPeriod(basis, tampered, "month:2021-01"),
      sortCardPeriods(basis, tampered, "count"),
      switchCardPeriodType(basis, tampered, "year")
    ]) {
      const url = new URL(href, "http://localhost");
      assert.equal(url.pathname, `/listcardperiod${basis}`);
      assert.equal(url.searchParams.has("basis"), false);
    }
    assert.deepEqual(tampered.basis, [basis === "release" ? "birth" : "release"]);
  }
});

test("目录排序切换方向并清除旧选择，卡片翻页保留选择", () => {
  const time = params(sortCardPeriods("release", query, "time"));
  assert.equal(time.get("period_order"), "asc");
  assert.equal(time.has("key"), false);
  const count = params(sortCardPeriods("release", query, "count"));
  assert.equal(count.get("period_sort"), "count");
  assert.equal(count.get("period_order"), "desc");
  const cards = params(periodCardsHref("release", query, { p: 5, ps: 36 }));
  assert.equal(cards.get("key"), "month:2020-12");
  assert.equal(cards.get("period_p"), "3");
  assert.equal(cards.get("ps"), "36");
  assert.equal(cards.get("od"), "vw");
  assert.equal(count.get("od"), "vw");
  assert.deepEqual(query.key, ["month:2020-12"]);
});

test("已拥有排序首次倒序，重复点击切换方向，保留卡片排序和两侧条数", () => {
  const cases: [Record<string, string[]>, string][] = [
    [query, "desc"],
    [{ ...query, period_sort: ["owned"], period_order: ["desc"] }, "asc"],
    [{ ...query, period_sort: ["owned"], period_order: ["asc"] }, "desc"]
  ];
  for (const [current, expected] of cases) {
    const href = new URL(sortCardPeriods("release", current, "owned"), "http://localhost");
    const result = href.searchParams;
    assert.equal(href.pathname, "/listcardperiodrelease");
    assert.equal(result.get("period_sort"), "owned");
    assert.equal(result.get("period_order"), expected);
    assert.equal(result.get("p"), "1");
    assert.equal(result.get("period_p"), "1");
    assert.equal(result.has("key"), false);
    assert.equal(result.get("period_type"), "month");
    assert.equal(result.get("period_ps"), "24");
    assert.equal(result.get("ps"), "18");
    assert.equal(result.get("od"), "vw");
    assert.equal(result.get("order"), "asc");
  }
  assert.deepEqual(query.period_sort, ["time"]);
});

test("周期选择、目录排序、类型和两侧条数链接均保留拥有筛选", () => {
  for (const value of ["1", "3", "7"]) {
    const current = { ...query, mowned: [value] };
    for (const href of [
      selectCardPeriod("release", current, "month:2021-01"),
      sortCardPeriods("release", current, "owned"),
      switchCardPeriodType("release", current, "year"),
      periodCardsHref("release", current, { p: 2, ps: 24 }),
      periodCardsHref("release", current, { period_ps: 36, period_p: 1, key: null, p: 1 })
    ]) {
      assert.equal(params(href).get("mowned"), value);
    }
  }
});

test("周期数量链接转换为 listcard 日期和库存筛选", () => {
  const period = {
    Key: "month:2020-12",
    Label: "2020-12",
    RangeStart: "2020-12-01",
    RangeEnd: "2020-12-31",
    Count: 7,
    CountOwned: 3
  };
  const releaseAll = new URL(periodListCardHref("release", period), "http://localhost");
  assert.equal(releaseAll.pathname, "/listcard");
  assert.equal(releaseAll.searchParams.get("rs"), period.RangeStart);
  assert.equal(releaseAll.searchParams.get("re"), period.RangeEnd);
  assert.equal(releaseAll.searchParams.get("mowned"), "1");

  const releaseOwned = new URL(periodListCardHref("release", period, true), "http://localhost");
  assert.equal(releaseOwned.pathname, "/listcard");
  assert.equal(releaseOwned.searchParams.get("rs"), period.RangeStart);
  assert.equal(releaseOwned.searchParams.get("re"), period.RangeEnd);
  assert.equal(releaseOwned.searchParams.get("mowned"), "3");

  const birth = new URL(periodListCardHref("birth", period), "http://localhost");
  assert.equal(birth.pathname, "/listcard");
  assert.equal(birth.searchParams.get("mbs"), period.RangeStart);
  assert.equal(birth.searchParams.get("mbe"), period.RangeEnd);
  assert.equal(birth.searchParams.get("mowned"), "3");
  assert.equal(birth.searchParams.has("rs"), false);
  assert.equal(birth.searchParams.has("re"), false);
});
