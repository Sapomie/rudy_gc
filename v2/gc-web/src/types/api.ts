export interface ApiEnvelope<T> {
  data: T | null;
  error: ApiErrorBody | null;
}

export interface ApiErrorBody {
  code: string;
  message: string;
}

export interface ApiError extends Error {
  status?: number;
  detail?: unknown;
}

export interface CastTag {
  person_id: number;
  name: string;
  display_name: string;
  name_show: string;
  href: string;
}

export interface MovieCard {
  movie_name: string;
  movie_jav_id: string;
  movie_href: string;
  title: string;
  encode_name: string;
  jacket_img: string;
  come_times: number;
  sc_times: number;
  highest_rank: number;
  last_sc_time: number;
  length: number;
  score: number;
  viewers_number_want: number;
  viewers_number_owned: number;
  viewers_number_watched: number;
  owned_w_media: number;
  video_url_w_media: string;
  film_birth_date_w_media: string;
  genre: string[] | null;
  cast: CastTag[] | null;
  director: string;
  director_href: string;
  prefix: string;
  prefix_href: string;
  bus_url: string;
  search_url: string;
  jav_url: string;
  releasing_date: string;
  update_date: string;
  first_ranking_day: string;
  label: string;
  label_href: string;
  maker: string;
  maker_href: string;
  need_download: number;
}

export interface CardsFilterSnapshot {
  cast_names: string;
  person_ids: string;
  genre_names: string;
  director_name: string;
  prefix_name: string;
  maker_name: string;
  label_name: string;
  label_jav_id: string;
  album_name: string;
  releasing_date_start: string;
  releasing_date_end: string;
  media_birth_time_start: string;
  media_birth_time_end: string;
  cast_age_min?: number;
  cast_age_max?: number;
  start_ranking_date_start: string;
  start_ranking_date_end: string;
  days_in_rank_min?: number;
  need_download: number;
  word: string;
  media_owned: number;
  view_watched_min?: number;
  view_watched_max?: number;
  score_min?: number;
  score_max?: number;
  last_sc_time_min: string;
  last_sc_time_max: string;
  sc_times_min?: number;
  sc_times_max?: number;
  come_times_min?: number;
  come_times_max?: number;
  media_dir_1: string;
  media_dir_2: string;
  media_dir_3: string;
  media_dir_4: string;
}

export interface CardsListResponse {
  view: string;
  title: string;
  items: MovieCard[];
  total: number;
  page: number;
  page_size: number;
  order_by: string;
  order: string;
  views: Array<{ key: string; label: string; href: string }>;
  filter_snapshot: CardsFilterSnapshot;
  random_requested?: number;
}

export interface MovieDetailResponse {
  movie: MovieCard;
  media: MovieMedia | null;
  movie_albums: MovieAlbumOption[] | null;
  marked_delete: boolean;
  rank_infos: MovieRankInfo[] | null;
  sc_events: MovieScEvent[] | null;
  javbus_fetch: MovieFetchSiteStatus | null;
  javbus_magnets: MovieJavbusMagnet[] | null;
  sukebei_fetch: MovieFetchSiteStatus | null;
  sukebei_rows: MovieSukebeiTorrent[] | null;
  sehuatang_rows: MovieSehuatangMagnet[] | null;
}

export interface MovieAlbumOption {
  id: number;
  name: string;
  selected: boolean;
}

export interface MovieMedia {
  movie_jav_id: string;
  movie_name: string;
  file_name: string;
  directory: string;
  file_path: string;
  source_torrent_hash: string;
  birth_time: string;
  size: number;
  height: number;
  bit_rate: number;
  duration_minutes: number;
  frame: number;
  self_make: string;
  is_removed: number;
}

export interface MovieRankInfo {
  date: string;
  rank: number;
  href: string;
}

export interface MovieScEvent {
  sc_name: string;
  sc_time: number;
  cooldown: number;
  is_come: boolean;
  href: string;
}

export interface MovieFetchSiteStatus {
  movie_jav_id: string;
  movie_name: string;
  release_date: string;
  fetch_status: number;
  fetch_status_text: string;
  try_count: number;
  last_fetch_time: string;
  last_fetch_ago: string;
  last_error: string;
  torrent_hash_count: number;
  latest_publish_time: string;
  source_url: string;
}

export interface MovieJavbusMagnet {
  row_id: number;
  magnet_name: string;
  info_hash: string;
  size_text: string;
  share_date: string;
  has_hd: boolean;
  has_subtitle: boolean;
  has_matched_hash: boolean;
  is_in_download: boolean;
  is_in_pending: boolean;
}

export interface MovieSukebeiTorrent {
  row_id: number;
  torrent_title: string;
  view_url: string;
  info_hash: string;
  size_text: string;
  publish_time: string;
  seeders: number;
  leechers: number;
  completed: number;
  has_matched_hash: boolean;
  is_in_download: boolean;
  is_in_pending: boolean;
}

export interface MovieSehuatangMagnet {
  row_id: number;
  thread_title: string;
  thread_url: string;
  info_hash: string;
  post_time: string;
  has_matched_hash: boolean;
  is_in_download: boolean;
  is_in_pending: boolean;
}

export interface RankDayResponse {
  title: string;
  rank_date: string;
  items: MovieCard[];
  total: number;
  page: number;
  page_size: number;
  prev_date: string;
  next_date: string;
  random_date: string;
  latest_date: string;
  prev_disabled: boolean;
  next_disabled: boolean;
}

export interface RankPeriodResponse {
  title: string;
  period_key: string;
  period_type: string;
  period_type_label: string;
  category: number;
  category_label: string;
  range_start: string;
  range_end: string;
  items: RankPeriodCard[];
  total: number;
  page: number;
  page_size: number;
  prev_key: string;
  next_key: string;
  prev_href: string;
  next_href: string;
  random_href: string;
  prev_disabled: boolean;
  next_disabled: boolean;
  type_links: RankSwitchLink[];
  category_links: RankSwitchLink[];
  latest_href: string;
}

export interface RankPeriodCard {
  movie: MovieCard;
  rank_pos: number;
  score: number;
  days_in_rank: number;
  used_pick_days: number;
  best_rank: number;
  worst_picked_rank: number;
  prev_rank: number;
  prev_rank_text: string;
  rank_change: number;
  rank_change_text: string;
  rank_change_class: string;
  peak_rank: number;
}

export interface RankSwitchLink {
  label: string;
  href: string;
  active: boolean;
}

export interface PageColumn {
  key: string;
  label: string;
  sortable: boolean;
  link?: string;
}

export interface PageFilter {
  key: string;
  label: string;
  type: "text" | "select" | string;
  placeholder?: string;
  value?: string;
  options?: PageFilterOption[];
}

export interface PageFilterOption {
  label: string;
  value: string;
}

export interface PageAction {
  key: string;
  label: string;
  method: string;
  path: string;
  variant: string;
  disabled: boolean;
  reason?: string;
}

export interface PageLink {
  label: string;
  href: string;
  kind: string;
}

export interface PageStat {
  label: string;
  value: string;
  note?: string;
}

export type PageRow = Record<string, string>;

export interface PageListResponse {
  key: string;
  title: string;
  description: string;
  legacy_path: string;
  kind: "list" | "operation" | string;
  page: number;
  page_size: number;
  total: number;
  order_by: string;
  order: string;
  query: string;
  columns: PageColumn[];
  rows: PageRow[];
  filters: PageFilter[];
  actions: PageAction[];
  links: PageLink[];
  stats: PageStat[];
}

export interface PageSummary {
  key: string;
  title: string;
  description: string;
  legacy_path: string;
  kind: string;
  group: string;
}
