import { requestJSON } from "@/shared/api/request";
import type {
  FetchAcceptedResponse,
  FetchRequest,
  FetchRun,
  FetchRunListResponse,
  HealthState,
  PageQuery,
  RankDayStatFilters,
  RankDayStatListResponse,
  RankFilters,
  RankListResponse,
  RankPeriodFilters,
  RankPeriodItemFilters,
  RankPeriodItemListResponse,
  RankPeriodListResponse,
  RankPeriodStatFilters,
  RankPeriodStatListResponse,
  RankSyncManifestListResponse
} from "@/features/dailybest/types";

export type { HealthState } from "@/features/dailybest/types";
export const DEFAULT_LIST_PAGE_SIZE = 24;

function defaultPageQuery(): PageQuery {
  return { p: 1, ps: DEFAULT_LIST_PAGE_SIZE };
}

export async function loadHealthState(): Promise<HealthState> {
  const [health, ready] = await Promise.all([fetch("/healthz"), fetch("/readyz")]);
  return health.ok && ready.ok ? "ready" : "unavailable";
}

export function listRankSyncManifests(): Promise<RankSyncManifestListResponse> {
  return requestJSON("/api/v1/dailybest/rank-sync-manifests");
}
export function listRanks(filters: RankFilters = defaultPageQuery()): Promise<RankListResponse> {
  return requestJSON(withQuery("/api/v1/dailybest/ranks", filters));
}
export function listRankDayStats(filters: RankDayStatFilters = defaultPageQuery()): Promise<RankDayStatListResponse> {
  return requestJSON(withQuery("/api/v1/dailybest/rank-day-stats", filters));
}
export function listRankPeriods(filters: RankPeriodFilters = defaultPageQuery()): Promise<RankPeriodListResponse> {
  return requestJSON(withQuery("/api/v1/dailybest/rank-periods", filters));
}
export function listRankPeriodItems(
  filters: RankPeriodItemFilters = defaultPageQuery()
): Promise<RankPeriodItemListResponse> {
  return requestJSON(withQuery("/api/v1/dailybest/rank-period-items", filters));
}
export function listRankPeriodStats(
  filters: RankPeriodStatFilters = defaultPageQuery()
): Promise<RankPeriodStatListResponse> {
  return requestJSON(withQuery("/api/v1/dailybest/rank-period-stats", filters));
}
export function listFetchRuns(page: PageQuery = defaultPageQuery()): Promise<FetchRunListResponse> {
  return requestJSON(withQuery("/api/v1/dailybest/fetch-runs", page));
}
export function getFetchRun(runID: number): Promise<FetchRun> {
  return requestJSON(`/api/v1/dailybest/fetch-runs/${encodeURIComponent(runID)}`);
}
export function startFetch(request: FetchRequest): Promise<FetchAcceptedResponse> {
  return requestJSON("/api/v1/dailybest/fetch", { method: "POST", body: request });
}

function withQuery(path: string, filters: object): string {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(filters)) {
    if (value !== undefined && value !== "" && value !== 0) query.set(key, String(value));
  }
  const encoded = query.toString();
  return encoded ? `${path}?${encoded}` : path;
}
