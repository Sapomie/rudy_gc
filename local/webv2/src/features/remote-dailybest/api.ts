import { getJSON, sendJSON } from "@/shared/api/request";
import type {
  FetchAcceptedResponse,
  FetchLogFilters,
  FetchLogListResponse,
  FetchRun,
  FetchRunListResponse,
  PageQuery,
  RankDayComparisonFilters,
  RankDayComparisonListResponse,
  RankDayStatFilters,
  RankDayStatListResponse,
  RankDaySyncResponse,
  RankDaySyncTarget,
  SourceConfig,
  TargetProbe,
  TargetProbeFilters,
  TargetProbeListResponse
} from "@/features/remote-dailybest/types";

export const REMOTE_API_PREFIX = "/remote-api";
export const DEFAULT_LIST_PAGE_SIZE = 24;

function defaultPageQuery(): PageQuery {
  return { p: 1, ps: DEFAULT_LIST_PAGE_SIZE };
}

function remotePath(path: string): string {
  return `${REMOTE_API_PREFIX}${path}`;
}

export function listRankDayStats(filters: RankDayStatFilters = defaultPageQuery()): Promise<RankDayStatListResponse> {
  return getJSON(withQuery(remotePath("/api/v1/dailybest/rank-day-stats"), filters));
}

export function listLocalRankDayComparisons(
  filters: RankDayComparisonFilters = defaultPageQuery()
): Promise<RankDayComparisonListResponse> {
  return getJSON(withQuery("/api/dailybest/rank-days", filters));
}

export function syncLocalRankDays(targets: RankDaySyncTarget[]): Promise<RankDaySyncResponse> {
  return sendJSON("/api/dailybest/rank-days/sync", "POST", { targets });
}

export function listFetchRuns(page: PageQuery = defaultPageQuery()): Promise<FetchRunListResponse> {
  return getJSON(withQuery(remotePath("/api/v1/dailybest/fetch-runs"), page));
}

export function listFetchLogs(filters: FetchLogFilters = defaultPageQuery()): Promise<FetchLogListResponse> {
  return getJSON(withQuery(remotePath("/api/v1/dailybest/fetch-logs"), filters));
}

export function listTargetProbes(filters: TargetProbeFilters = defaultPageQuery()): Promise<TargetProbeListResponse> {
  return getJSON(withQuery(remotePath("/api/v1/dailybest/target-probes"), filters));
}

export function startTargetProbe(): Promise<TargetProbe> {
  return sendJSON(remotePath("/api/v1/dailybest/target-probes"), "POST");
}

export function getFetchRun(runID: number): Promise<FetchRun> {
  return getJSON(remotePath(`/api/v1/dailybest/fetch-runs/${encodeURIComponent(runID)}`));
}

export function startFetch(): Promise<FetchAcceptedResponse> {
  return sendJSON(remotePath("/api/v1/dailybest/fetch"), "POST");
}

export function getSourceConfig(): Promise<SourceConfig> {
  return getJSON(remotePath("/api/v1/dailybest/source-config"));
}

export function updateSourceConfig(javAddress: string): Promise<SourceConfig> {
  return sendJSON(remotePath("/api/v1/dailybest/source-config"), "PUT", { jav_address: javAddress });
}

function withQuery(path: string, filters: object): string {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(filters)) {
    if (value !== undefined && value !== "" && value !== 0) query.set(key, String(value));
  }
  const encoded = query.toString();
  return encoded ? `${path}?${encoded}` : path;
}
