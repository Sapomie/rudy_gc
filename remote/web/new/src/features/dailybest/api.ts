import { requestJSON } from "@/shared/api/request";
import type {
  FetchAcceptedResponse,
  FetchLogFilters,
  FetchLogListResponse,
  FetchRun,
  FetchRunListResponse,
  HealthState,
  PageQuery,
  RankDayStatFilters,
  RankDayStatListResponse,
  RankFilters,
  RankListResponse,
  RuntimeEventFilters,
  RuntimeEventListResponse,
  RuntimeStatus,
  SourceConfig,
  TargetProbeFilters,
  TargetProbeListResponse
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

export function listRanks(filters: RankFilters = defaultPageQuery()): Promise<RankListResponse> {
  return requestJSON(withQuery("/api/v1/dailybest/ranks", filters));
}
export function listRankDayStats(filters: RankDayStatFilters = defaultPageQuery()): Promise<RankDayStatListResponse> {
  return requestJSON(withQuery("/api/v1/dailybest/rank-day-stats", filters));
}
export function listFetchRuns(page: PageQuery = defaultPageQuery()): Promise<FetchRunListResponse> {
  return requestJSON(withQuery("/api/v1/dailybest/fetch-runs", page));
}
export function listFetchLogs(filters: FetchLogFilters = defaultPageQuery()): Promise<FetchLogListResponse> {
  return requestJSON(withQuery("/api/v1/dailybest/fetch-logs", filters));
}
export function loadRuntimeStatus(): Promise<RuntimeStatus> {
  return requestJSON("/api/v1/dailybest/runtime-status");
}
export function listRuntimeEvents(
  filters: RuntimeEventFilters = defaultPageQuery()
): Promise<RuntimeEventListResponse> {
  return requestJSON(withQuery("/api/v1/dailybest/runtime-events", filters));
}
export function listTargetProbes(filters: TargetProbeFilters = defaultPageQuery()): Promise<TargetProbeListResponse> {
  return requestJSON(withQuery("/api/v1/dailybest/target-probes", filters));
}
export function getFetchRun(runID: number): Promise<FetchRun> {
  return requestJSON(`/api/v1/dailybest/fetch-runs/${encodeURIComponent(runID)}`);
}
export function startFetch(): Promise<FetchAcceptedResponse> {
  return requestJSON("/api/v1/dailybest/fetch", { method: "POST" });
}

export function getSourceConfig(): Promise<SourceConfig> {
  return requestJSON("/api/v1/dailybest/source-config");
}

export function updateSourceConfig(javAddress: string): Promise<SourceConfig> {
  return requestJSON("/api/v1/dailybest/source-config", { method: "PUT", body: { jav_address: javAddress } });
}

function withQuery(path: string, filters: object): string {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(filters)) {
    if (value !== undefined && value !== "" && value !== 0) query.set(key, String(value));
  }
  const encoded = query.toString();
  return encoded ? `${path}?${encoded}` : path;
}
