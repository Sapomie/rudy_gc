export type HealthState = "loading" | "ready" | "unavailable";

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

export interface BrowserRuntimeStatus {
  state: string;
  last_heartbeat_at: number;
  last_error: string;
  version: string;
  user_agent: string;
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

export interface RuntimeEvent {
  id: number;
  component: string;
  event_type: string;
  level: string;
  message: string;
  fetch_run_id: number;
  target_probe_id: number;
  occurred_at: number;
  created_on: number;
}

export interface RuntimeEventFilters {
  p?: number;
  ps?: number;
  component?: string;
  level?: string;
}

export interface RuntimeEventListResponse {
  events: RuntimeEvent[];
  page_info?: PageInfo;
}

export interface RuntimeStatus {
  api_status: string;
  database_status: string;
  browser_status: BrowserRuntimeStatus;
  browser_error: string;
  target_status: string;
  latest_probe?: TargetProbe;
  next_probe_at: number;
  last_probe_success: number;
  last_probe_failure: number;
}

export interface MovieBasic {
  movie_name: string;
  movie_code: string;
  title: string;
  cover_url: string;
}

export interface Rank {
  id: number;
  rank_key: string;
  movie_jav_id: string;
  day_number: number;
  rank_pos: number;
  category: number;
  movie?: MovieBasic;
  created_on: number;
  updated_on: number;
}
export interface RankFilters {
  p?: number;
  ps?: number;
  business_date?: string;
  day_number?: number;
  category?: number;
  rank_pos?: number;
  movie_jav_id?: string;
}
export interface RankListResponse {
  ranks: Rank[];
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
export interface SourceConfig {
  jav_address: string;
  target_probe_url: string;
  created_on: number;
  updated_on: number;
}
