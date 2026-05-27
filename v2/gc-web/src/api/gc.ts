import { apiGet } from "@/api/request";
import type {
  CardsListResponse,
  MovieDetailResponse,
  PageListResponse,
  PageSummary,
  RankDayResponse,
  RankPeriodResponse,
} from "@/types/api";

export function listCards(params: Record<string, unknown>) {
  return apiGet<CardsListResponse>("/cards", params);
}

export function getMovieDetail(movie: string) {
  return apiGet<MovieDetailResponse>(`/movie/${encodeURIComponent(movie)}`);
}

export function getRankDay(params: Record<string, unknown>) {
  return apiGet<RankDayResponse>("/rank/day", params);
}

export function getRankPeriod(params: Record<string, unknown>) {
  return apiGet<RankPeriodResponse>("/rank/period", params);
}

export function listPageSummaries() {
  return apiGet<PageSummary[]>("/pages");
}

export function getGenericPage(key: string, params: Record<string, unknown>) {
  return apiGet<PageListResponse>(`/pages/${encodeURIComponent(key)}`, params);
}
