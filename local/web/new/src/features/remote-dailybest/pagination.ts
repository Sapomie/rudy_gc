import type { LocationQuery, LocationQueryRaw } from "vue-router";
import { DEFAULT_LIST_PAGE_SIZE } from "@/features/remote-dailybest/api";
import type { PageInfo, PageQuery } from "@/features/remote-dailybest/types";

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

function positiveQueryInteger(value: LocationQuery[string], fallback: number): number {
  const raw = Array.isArray(value) ? value[0] : value;
  const parsed = Number(raw);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
}
