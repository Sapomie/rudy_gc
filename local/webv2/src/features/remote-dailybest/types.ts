import type { PageInfo as LocalPageInfo } from "@/shared/types/gc";

export interface PageQuery {
  p: number;
  ps: number;
}
export interface PageInfo {
  total: number;
  page: number;
  page_size: number;
  page_total: number;
}

export interface FetchRun {
  id: number;
  business_date: string;
  category: number;
  trigger_type: string;
  status: string;
  started_at: number;
  finished_at: number;
  page_count: number;
  item_count: number;
  checksum: string;
  error_message?: string;
  created_at: number;
  updated_at: number;
}
export interface FetchRunListResponse {
  runs: FetchRun[];
  page_info?: PageInfo;
}
export interface FetchAcceptedResponse {
  run_id: number;
  status: string;
  message: string;
}

export interface FetchLog {
  id: number;
  fetch_run_id: number;
  page_number: number;
  attempt_number: number;
  request_url: string;
  status: string;
  http_status: number;
  duration_ms: number;
  error_message: string;
  created_on: number;
  updated_on: number;
}
export interface FetchLogFilters {
  p?: number;
  ps?: number;
  fetch_run_id?: number;
  page_number?: number;
  status?: string;
}
export interface FetchLogListResponse {
  logs: FetchLog[];
  page_info?: PageInfo;
}

export interface TargetProbe {
  id: number;
  request_id: string;
  status: string;
  target_url: string;
  final_url: string;
  http_status: number;
  duration_ms: number;
  browser_status: string;
  error_type: string;
  error_message: string;
  started_at: number;
  finished_at: number;
  created_on: number;
  updated_on: number;
}

export interface TargetProbeFilters {
  p?: number;
  ps?: number;
  status?: string;
  started_from?: number;
  started_to?: number;
}

export interface TargetProbeListResponse {
  probes: TargetProbe[];
  page_info?: PageInfo;
}

export interface RankDayStat {
  id: number;
  day_number: number;
  rank_date: string;
  prev_day_number: number;
  rank_row_count: number;
  movie_count: number;
  continued_movie_count: number;
  first_seen_movie_count: number;
  returning_movie_count: number;
  dropped_movie_count: number;
  source_checksum: string;
  fetched_at: number;
  synced_at: number;
  calculated_on: number;
  created_on: number;
  updated_on: number;
}
export interface RankDayStatFilters {
  p?: number;
  ps?: number;
  rank_date?: string;
  day_number?: number;
}
export interface RankDayStatListResponse {
  stats: RankDayStat[];
  page_info?: PageInfo;
}

export type RankDaySyncStatus = "missing" | "incomplete" | "version_mismatch" | "pending" | "failed" | "completed";

export interface LocalRankCategoryState {
  rank_count: number;
  checksum: string;
  remote_fetch_run_id: number;
  sync_status: string;
  process_status: string;
  synced_at: number;
  processed_at: number;
  last_error: string;
}

export interface RankDayRemoteCategory {
  fetch_run_id: number;
  category: number;
  page_count: number;
  item_count: number;
  rank_count: number;
  checksum: string;
  finished_at: number;
}

export interface RankDayCategoryComparison {
  remote: RankDayRemoteCategory;
  local: LocalRankCategoryState;
  status: RankDaySyncStatus;
  needs_sync: boolean;
}

export interface RankDayComparison {
  business_date: string;
  day_number: number;
  page_count: number;
  item_count: number;
  remote_rank_count: number;
  local_rank_count: number;
  finished_at: number;
  status: RankDaySyncStatus;
  needs_sync: boolean;
  categories: RankDayCategoryComparison[];
}

export interface RankDayComparisonFilters {
  p?: number;
  ps?: number;
  business_date_start?: string;
  business_date_end?: string;
  day_number?: number;
}

export interface RankDayComparisonQuery {
  BusinessDateStart: string;
  BusinessDateEnd: string;
  Page: number;
  PageSize: number;
}

export interface RankDayComparisonListResponse {
  Days: RankDayComparison[];
  Total: number;
  Query: RankDayComparisonQuery;
  PageInfo: LocalPageInfo;
}

export interface RankDaySyncTarget {
  day_number: number;
  category: number;
}

export interface RankDaySyncResponse {
  job_id: string;
  task_type: string;
  message: string;
}
export interface SourceConfig {
  jav_address: string;
  target_probe_url: string;
  created_on: number;
  updated_on: number;
}
