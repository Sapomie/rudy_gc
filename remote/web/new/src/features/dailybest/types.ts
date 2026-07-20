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
export interface FetchRequest {
  business_date: string;
  category: number;
}
export interface FetchAcceptedResponse {
  run_id: number;
  status: string;
  message: string;
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
  rise_movie_count: number;
  fall_movie_count: number;
  unchanged_movie_count: number;
  source_checksum: string;
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

export interface RankPeriod {
  id: number;
  period_type: number;
  period_key: string;
  category: number;
  start_day_number: number;
  end_day_number: number;
  pick_days: number;
  top_n: number;
  status: number;
  created_on: number;
  updated_on: number;
}
export interface RankPeriodFilters {
  p?: number;
  ps?: number;
  period_type?: number;
  period_key?: string;
  category?: number;
}
export interface RankPeriodListResponse {
  periods: RankPeriod[];
  page_info?: PageInfo;
}

export interface RankPeriodItem {
  id: number;
  period_id: number;
  movie_jav_id: string;
  movie?: MovieBasic;
  rank_pos: number;
  score: number;
  days_in_rank: number;
  used_pick_days: number;
  first_rank_day_number: number;
  last_rank_day_number: number;
  best_rank: number;
  worst_picked_rank: number;
  prev_rank: number;
  rank_change: number;
  peak_rank: number;
  created_on: number;
  updated_on: number;
}
export interface RankPeriodItemFilters {
  p?: number;
  ps?: number;
  period_id?: number;
  movie_jav_id?: string;
}
export interface RankPeriodItemListResponse {
  items: RankPeriodItem[];
  page_info?: PageInfo;
}

export interface RankPeriodStat {
  id: number;
  period_id: number;
  prev_period_id: number;
  source_rank_row_count: number;
  source_rank_day_count: number;
  source_movie_count: number;
  missing_rank_day_count: number;
  period_movie_count: number;
  continued_movie_count: number;
  first_seen_movie_count: number;
  returning_movie_count: number;
  dropped_movie_count: number;
  rise_movie_count: number;
  fall_movie_count: number;
  unchanged_movie_count: number;
  peak_movie_count: number;
  rank_one_movie_jav_id: string;
  rank_one_movie?: MovieBasic;
  average_score: number;
  average_days_in_rank: number;
  source_fetched_start_on: number;
  source_fetched_end_on: number;
  source_checksum: string;
  calculated_on: number;
  created_on: number;
  updated_on: number;
}
export interface RankPeriodStatFilters {
  p?: number;
  ps?: number;
  period_id?: number;
}
export interface RankPeriodStatListResponse {
  stats: RankPeriodStat[];
  page_info?: PageInfo;
}

export interface RankSyncManifest {
  fetch_run_id: number;
  business_date: string;
  day_number: number;
  category: number;
  page_count: number;
  item_count: number;
  checksum: string;
  finished_at: number;
}
export interface RankSyncManifestListResponse {
  manifests: RankSyncManifest[];
}
