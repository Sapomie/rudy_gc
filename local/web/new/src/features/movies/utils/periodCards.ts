import type { PageInfo } from "@/shared/types/gc";
import type { MovieRecord } from "@/features/movies/types";

export type CardPeriodBasis = "release" | "birth";
export type CardPeriodSort = "time" | "count" | "owned";

export interface CardPeriod {
  Key: string;
  Label: string;
  RangeStart: string;
  RangeEnd: string;
  Count: number;
  CountOwned: number;
}

export interface PeriodCardsResponse {
  Basis: CardPeriodBasis;
  PeriodType: string;
  Periods: CardPeriod[];
  PeriodTotal: number;
  SelectedPeriod: CardPeriod | null;
  movies: MovieRecord[];
  total: number;
  PageInfo: PageInfo;
  PeriodPageInfo: PageInfo;
  sortQuery: MovieRecord;
  ownedQuery?: MovieRecord;
  Query: Record<string, string[]>;
}

export function periodCardsHref(
  basis: CardPeriodBasis,
  query: Record<string, string[]>,
  changes: Record<string, string | number | null>
): string {
  const params = new URLSearchParams();
  for (const [key, values] of Object.entries(query)) {
    for (const value of values) params.append(key, value);
  }
  for (const [key, value] of Object.entries(changes)) {
    if (value === null) params.delete(key);
    else params.set(key, String(value));
  }
  params.delete("basis");
  const encoded = params.toString();
  return `/listcardperiod${basis}${encoded ? `?${encoded}` : ""}`;
}

export function switchCardPeriodType(
  basis: CardPeriodBasis,
  query: PeriodCardsResponse["Query"],
  value: string
): string {
  return periodCardsHref(basis, query, { period_type: value, key: null, period_p: 1, p: 1 });
}

export function selectCardPeriod(basis: CardPeriodBasis, query: PeriodCardsResponse["Query"], key: string): string {
  return periodCardsHref(basis, query, { key, p: 1 });
}

export function sortCardPeriods(basis: CardPeriodBasis, query: PeriodCardsResponse["Query"], sort: CardPeriodSort): string {
  const direction = query.period_sort?.[0] === sort && query.period_order?.[0] === "desc" ? "asc" : "desc";
  return periodCardsHref(basis, query, { period_sort: sort, period_order: direction, period_p: 1, p: 1, key: null });
}
