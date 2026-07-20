import type { LocationQuery, LocationQueryRaw } from "vue-router";
import type { PageInfo, PageQuery } from "@/features/dailybest/types";
import { DEFAULT_LIST_PAGE_SIZE } from "@/features/dailybest/api";

export interface PaginationItem {
  key: string;
  label: string;
  page?: number;
  active: boolean;
  disabled: boolean;
  ellipsis: boolean;
}

export function pageQueryFromRoute(query: LocationQuery): PageQuery {
  return {
    p: positiveQueryInteger(query.p, 1),
    ps: positiveQueryInteger(query.ps, DEFAULT_LIST_PAGE_SIZE)
  };
}

export function pageInfoOrFallback(pageInfo: PageInfo | undefined, page: PageQuery, itemCount: number): PageInfo {
  return (
    pageInfo || {
      total: itemCount,
      page: page.p,
      page_size: page.ps,
      page_total: 1
    }
  );
}

export function compactPageQuery(query: LocationQueryRaw, page: number, pageSize: number): LocationQueryRaw {
  const result: LocationQueryRaw = { ...query };
  if (page <= 1) delete result.p;
  else result.p = String(page);
  if (pageSize === DEFAULT_LIST_PAGE_SIZE) delete result.ps;
  else result.ps = String(pageSize);
  return result;
}

export function buildPaginationItems(pageInfo: PageInfo, window = 3): PaginationItem[] {
  const current = Math.min(Math.max(pageInfo.page, 1), Math.max(pageInfo.page_total, 1));
  const pageTotal = Math.max(pageInfo.page_total, 1);
  const items: PaginationItem[] = [
    pageItem("Start", 1, false, current === 1),
    pageItem("Prev", Math.max(current - 1, 1), false, current === 1)
  ];
  let left = current - window;
  let right = current + window;
  if (left > 1) {
    items.push(pageItem("1", 1));
    if (left > 2) items.push(ellipsisItem(items.length));
    left = Math.max(left, 2);
  }
  if (right < pageTotal) right = Math.min(right, pageTotal - 1);
  for (let page = left; page <= right; page += 1) {
    if (page < 1 || page > pageTotal) continue;
    items.push(pageItem(String(page), page, page === current));
  }
  if (right < pageTotal) {
    if (right < pageTotal - 1) items.push(ellipsisItem(items.length));
    items.push(pageItem(String(pageTotal), pageTotal, pageTotal === current));
  }
  items.push(
    pageItem("Next", Math.min(current + 1, pageTotal), false, current === pageTotal),
    pageItem("End", pageTotal, false, current === pageTotal)
  );
  return items;
}

function positiveQueryInteger(value: LocationQuery[string], fallback: number): number {
  const raw = Array.isArray(value) ? value[0] : value;
  const parsed = Number(raw);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
}

function pageItem(label: string, page: number, active = false, disabled = false): PaginationItem {
  return { key: `${label}-${page}`, label, page, active, disabled, ellipsis: false };
}

function ellipsisItem(index: number): PaginationItem {
  return { key: `ellipsis-${index}`, label: "...", active: false, disabled: true, ellipsis: true };
}
