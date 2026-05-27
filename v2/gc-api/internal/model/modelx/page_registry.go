package modelx

import (
	"rudy-gc-api/internal/types"
)

func pageConfigs() []*PageConfig {
	return []*PageConfig{
		castsPage(),
		castDetailPage(),
		recordsPage(),
		crawlRecordsPage(),
		scEventsPage(),
		mediasPage(),
		wAggEventsPage(),
		eItemsPage(),
		torrentAlbumsPage(),
		movieAlbumsPage(),
		dSeedsPage(),
		fetchJavbusPage(),
		fetchSukebeiPage(),
		fetchSehuatangPage(),
		wdirPage(),
		wMediaBirthPage(),
		wMediaBirthBucketPage(),
		movieReleaseAggPage(),
		movieReleaseBucketPage(),
		triggersPage(),
		crawlerTasksPage(),
		crawlerDetailLoopPage(),
		dailyBestTriggerPage(),
		seedsTriggerPage(),
		postProcessTriggerPage(),
		backfillTriggerPage(),
		fetchSiteTriggerPage(),
		fetchSehuatangTriggerPage(),
		mediaTriggerPage(),
		mediaRescanPage(),
		mediaRollbackPage(),
		scMediaMovePage(),
		scPickSmartPage(),
		scTriggersPage(),
		aggTriggersPage(),
	}
}

func castsPage() *PageConfig {
	return &PageConfig{
		Key:         "casts",
		Title:       "Casts",
		Description: "演员聚合列表，覆盖 legacy /casts 的筛选、排序、分页输出。",
		Group:       "List",
		LegacyPath:  "/casts",
		Kind:        pageKindList,
		BaseSQL:     "FROM c_person p",
		Columns: []*types.PageColumn{
			hiddenCol("id"),
			linkCol("name", "Name", "/cast/{id}", true),
			col("chinese", "Chinese", true),
			col("age", "Age", true),
			col("height", "Height", true),
			col("movie_number", "Movies", true),
			col("owned_w_media_number", "WMedia", true),
			col("sc_times", "SC", true),
			col("come_times", "Come", true),
			col("last_sc_time", "Last SC", true),
			col("highest_rank", "Highest Rank", true),
		},
		SearchColumns: []string{"p.name", "p.chinese", "p.alias"},
		SortColumns: map[string]string{
			"id": "p.id", "name": "p.name", "chinese": "p.chinese", "age": "p.birth_day", "birth_day": "p.birth_day", "height": "p.height", "movie_number": "p.movie_number",
			"owned_movie_number": "p.owned_movie_number", "owned_w_media_number": "p.owned_w_media_number",
			"sc_times": "p.sc_times", "come_times": "p.come_times", "last_sc_time": "p.last_sc_time",
			"highest_rank": "p.highest_rank",
		},
		OrderColumns: map[string]string{
			"name":                 "p.name",
			"chinese":              "CASE WHEN p.chinese <> '' THEN 0 ELSE 1 END ASC, p.chinese",
			"age":                  "CASE WHEN p.birth_day > 0 THEN 0 ELSE 1 END ASC, p.birth_day",
			"height":               "CASE WHEN p.height > 0 THEN 0 ELSE 1 END ASC, p.height",
			"movie_number":         "p.movie_number",
			"owned_movie_number":   "p.owned_w_media_number",
			"owned_w_media_number": "p.owned_w_media_number",
			"sc_times":             "p.sc_times",
			"come_times":           "p.come_times",
			"last_sc_time":         "p.last_sc_time",
			"highest_rank":         "p.highest_rank",
		},
		DefaultOrderBy:  "owned_w_media_number",
		DefaultOrder:    "desc",
		DefaultPageSize: 24,
		Filters: []*types.PageFilter{
			numberFilter("wowned_min", "WOwned Min", ""),
			numberFilter("wowned_max", "WOwned Max", ""),
			numberFilter("sc_times_min", "SC Min", ""),
			numberFilter("sc_times_max", "SC Max", ""),
			numberFilter("come_times_min", "Come Min", ""),
			numberFilter("come_times_max", "Come Max", ""),
			dateFilter("last_sc_from", "Last SC From"),
			dateFilter("last_sc_to", "Last SC To"),
		},
		FilterRules: []*PageFilterRule{
			rule("wowned_min", "p.owned_w_media_number", ">=", "int"),
			rule("wowned_max", "p.owned_w_media_number", "<=", "int"),
			rule("owned_min", "p.owned_w_media_number", ">=", "int"),
			rule("owned_max", "p.owned_w_media_number", "<=", "int"),
			rule("sc_times_min", "p.sc_times", ">=", "int"),
			rule("sc_times_max", "p.sc_times", "<=", "int"),
			rule("come_times_min", "p.come_times", ">=", "int"),
			rule("come_times_max", "p.come_times", "<=", "int"),
			rule("last_sc_from", "p.last_sc_time", ">=", "date_start"),
			rule("last_sc_to", "p.last_sc_time", "<=", "date_end"),
		},
		Links: []*types.PageLink{link("SC Events", "/sc-events", "related"), link("Movie Cards", "/cards", "related")},
	}
}

func castDetailPage() *PageConfig {
	cfg := castsPage()
	cfg.Key = "cast-detail"
	cfg.Title = "Cast Detail"
	cfg.Description = "演员详情基础信息，前端会联动 cards 的 pid 筛选展示影片。"
	cfg.LegacyPath = "/cast/:id"
	cfg.Filters = []*types.PageFilter{textFilter("id", "演员 ID", "例如 1")}
	cfg.FilterRules = []*PageFilterRule{rule("id", "p.id", "=", "int")}
	cfg.DefaultPageSize = 1
	return cfg
}

func recordsPage() *PageConfig {
	return &PageConfig{
		Key:         "records",
		Title:       "Records",
		Description: "legacy /records 与 /crawl-records 的抓取记录表。",
		Group:       "Crawler",
		LegacyPath:  "/records",
		Kind:        pageKindList,
		BaseSQL:     "FROM e_record r",
		Columns: []*types.PageColumn{
			col("start_time", "Start", true),
			col("type", "Type", true),
			col("detail_number", "Detail", true),
			col("duration", "Duration", true),
			col("name", "Name", true),
		},
		SearchColumns: []string{"r.name", "r.type"},
		SortColumns: map[string]string{
			"id": "r.id", "name": "r.name", "type": "r.type", "start_time": "r.start_time",
			"end_time": "r.end_time", "duration": "GREATEST(r.end_time - r.start_time, 0)", "detail_number": "r.detail_number",
		},
		OrderColumns: map[string]string{
			"start_time":    "r.start_time",
			"type":          "r.type",
			"detail_number": "r.detail_number",
			"duration":      "GREATEST(r.end_time - r.start_time, 0)",
			"name":          "r.name",
		},
		DefaultOrderBy:  "start_time",
		DefaultOrder:    "desc",
		DefaultPageSize: 50,
		DefaultFilters:  map[string]string{"type": "SeedsActive"},
		Filters: []*types.PageFilter{
			selectFilter("type", "Type", opt("All", ""), opt("DailyBest", "DailyBest"), opt("SeedsActive", "SeedsActive"), opt("SeedName", "SeedName")),
		},
		FilterRules: []*PageFilterRule{
			rule("type", "r.type", "=", "text"),
		},
	}
}

func crawlRecordsPage() *PageConfig {
	cfg := recordsPage()
	cfg.Key = "crawl-records"
	cfg.Title = "Crawl Records"
	cfg.LegacyPath = "/crawl-records"
	cfg.Description = "抓取记录列表，保留 legacy /crawl-records 的记录输出。"
	return cfg
}

func scEventsPage() *PageConfig {
	return &PageConfig{
		Key:         "sc-events",
		Title:       "SC Events",
		Description: "SC 事件列表，覆盖 /sc-events、/sc-events-cards 与详情入口。",
		Group:       "SC",
		LegacyPath:  "/sc-events",
		Kind:        pageKindList,
		BaseSQL:     "FROM g_sc s",
		Columns: []*types.PageColumn{
			hiddenCol("name"),
			linkCol("sc_time", "Time", "/sc-events/{name}", true),
			col("movie_number", "SC", true),
			linkCol("come_movie_name", "Come Movie", "/movie/{come_movie_name}", true),
			col("cooldown", "Cooldown", true),
			col("duration", "Duration", true),
			col("kind", "Kind", true),
			col("movie_cast", "Cast", true),
			col("vessel", "Vessel", true),
			col("fg", "FG", true),
		},
		SearchColumns: []string{"s.name", "s.come_movie_name", "s.movie_cast", "s.remarks", "s.kind"},
		SortColumns: map[string]string{
			"id": "s.id", "name": "s.name", "movie_number": "s.movie_number", "come_movie_name": "s.come_movie_name", "movie_cast": "s.movie_cast",
			"sc_time": "s.sc_time", "cooldown": "s.cooldown", "duration": "s.duration", "kind": "s.kind", "remarks": "s.remarks",
			"fg": "s.fg", "vessel": "s.vessel",
		},
		OrderColumns: map[string]string{
			"sc_time":         "s.sc_time",
			"movie_number":    "s.movie_number",
			"come_movie_name": "s.come_movie_name",
			"cooldown":        "s.cooldown",
			"duration":        "s.duration",
			"kind":            "s.kind",
			"movie_cast":      "s.movie_cast",
			"vessel":          "s.vessel",
			"fg":              "s.fg",
		},
		DefaultOrderBy:  "sc_time",
		DefaultOrder:    "desc",
		DefaultPageSize: 24,
		Filters: []*types.PageFilter{
			textFilter("kind", "Kind", "事件类型"),
			textFilter("movie_cast", "Movie Cast", "演员名"),
		},
		FilterRules: []*PageFilterRule{
			rule("kind", "s.kind", "=", "text"),
			rule("movie_cast", "s.movie_cast", "LIKE", "like"),
		},
		Links: []*types.PageLink{link("Smart Pick", "/sc-pick-smart-media", "operation"), link("SC Preview", "/triggers/sc", "operation")},
	}
}

func mediasPage() *PageConfig {
	return &PageConfig{
		Key:         "medias",
		Title:       "Medias",
		Description: "w_media 媒体文件列表，覆盖 legacy /medias。",
		Group:       "List",
		LegacyPath:  "/medias",
		Kind:        pageKindList,
		BaseSQL:     "FROM w_media m LEFT JOIN g_sc_stat gss ON gss.movie_jav_id = m.movie_jav_id",
		StaticWhere: "m.source_type = 2 AND m.is_removed = 1",
		Columns: []*types.PageColumn{
			col("id", "ID", true),
			linkCol("movie_name", "Movie", "/movie/{movie_name}", true),
			col("size", "Size", true),
			col("height", "Height", true),
			col("duration", "Duration", true),
			col("bit_rate", "Bit Rate", true),
			col("frame_average", "Frame", true),
			col("self_make", "Self", true),
			col("has_mask", "Mask", true),
			col("sc_times", "SC", true),
			col("last_sc_time", "Last SC", true),
			col("birth_time", "Birth", true),
			col("releasing_date", "Release", true),
		},
		SearchColumns: []string{"m.movie_name", "m.movie_jav_id", "m.file_name", "m.full_dir", "m.source_torrent_hash"},
		SortColumns: map[string]string{
			"id": "m.id", "movie_name": "m.movie_name", "file_name": "m.file_name", "full_dir": "m.full_dir",
			"size": "m.size", "width": "m.width", "height": "m.height", "bit_rate": "m.bit_rate", "duration": "m.duration",
			"frame_average": "m.frame_average", "birth_time": "m.birth_time", "releasing_date": "m.releasing_date",
			"has_sub": "m.has_sub", "self_make": "m.self_make", "has_mask": "m.has_mask", "is_removed": "m.is_removed",
			"sc_times": "COALESCE(gss.sc_times, 0)", "last_sc_time": "COALESCE(gss.last_sc_time, 0)",
		},
		OrderColumns: map[string]string{
			"movie_name":     "m.movie_name",
			"size":           "m.size",
			"height":         "m.height",
			"duration":       "m.duration",
			"bit_rate":       "m.bit_rate",
			"frame_average":  "m.frame_average",
			"self_make":      "m.self_make",
			"has_mask":       "m.has_mask",
			"sc_times":       "COALESCE(gss.sc_times, 0)",
			"last_sc_time":   "COALESCE(gss.last_sc_time, 0)",
			"releasing_date": "m.releasing_date",
			"birth_time":     "m.birth_time",
		},
		DefaultOrderBy:  "birth_time",
		DefaultOrder:    "desc",
		DefaultPageSize: 24,
		Filters: []*types.PageFilter{
			numberFilter("size_min", "Size Min(GB)", ""),
			numberFilter("size_max", "Size Max(GB)", ""),
			numberFilter("height_min", "Height Min", ""),
			numberFilter("height_max", "Height Max", ""),
			numberFilter("duration_min", "Duration Min", "分钟"),
			numberFilter("duration_max", "Duration Max", "分钟"),
			numberFilter("bit_rate_min", "BitRate Min", ""),
			numberFilter("bit_rate_max", "BitRate Max", ""),
			numberFilter("frame_average_min", "Frame Min", ""),
			numberFilter("frame_average_max", "Frame Max", ""),
			numberFilter("sc_times_min", "SC Min", ""),
			numberFilter("sc_times_max", "SC Max", ""),
			selectFilter("self_make", "Self Make", opt("All", ""), opt("No", "1"), opt("Yes", "2")),
			selectFilter("has_mask", "Mask", opt("All", ""), opt("有码", "1"), opt("去码", "2"), opt("无码", "3")),
			dateFilter("last_sc_from", "Last SC From"),
			dateFilter("last_sc_to", "Last SC To"),
			dateFilter("birth_from", "Birth From"),
			dateFilter("birth_to", "Birth To"),
			dateFilter("releasing_from", "Release From"),
			dateFilter("releasing_to", "Release To"),
		},
		FilterRules: []*PageFilterRule{
			rule("size_min", "m.size", ">=", "gb"),
			rule("size_max", "m.size", "<=", "gb"),
			rule("height_min", "m.height", ">=", "int"),
			rule("height_max", "m.height", "<=", "int"),
			rule("duration_min", "m.duration", ">=", "minutes"),
			rule("duration_max", "m.duration", "<=", "minutes"),
			rule("bit_rate_min", "m.bit_rate", ">=", "int"),
			rule("bit_rate_max", "m.bit_rate", "<=", "int"),
			rule("frame_average_min", "m.frame_average", ">=", "float"),
			rule("frame_average_max", "m.frame_average", "<=", "float"),
			rule("sc_times_min", "COALESCE(gss.sc_times, 0)", ">=", "int"),
			rule("sc_times_max", "COALESCE(gss.sc_times, 0)", "<=", "int"),
			rule("self_make", "m.self_make", "=", "int"),
			rule("has_mask", "m.has_mask", "=", "int"),
			rule("last_sc_from", "COALESCE(gss.last_sc_time, 0)", ">=", "date_start"),
			rule("last_sc_to", "COALESCE(gss.last_sc_time, 0)", "<=", "date_end"),
			rule("birth_from", "m.birth_time", ">=", "date_start"),
			rule("birth_to", "m.birth_time", "<=", "date_end"),
			rule("releasing_from", "m.releasing_date", ">=", "date_start"),
			rule("releasing_to", "m.releasing_date", "<=", "date_end"),
		},
		Links: []*types.PageLink{link("Media Ingest", "/triggers/media", "operation"), link("Media Rescan", "/triggers/media-rescan", "operation")},
	}
}

func wAggEventsPage() *PageConfig {
	return &PageConfig{
		Key:         "w-agg-events",
		Title:       "WAgg Events",
		Description: "聚合事件执行记录，覆盖 /w-agg-events。",
		Group:       "MovieAgg",
		LegacyPath:  "/w-agg-events",
		Kind:        pageKindList,
		BaseSQL:     "FROM w_agg_event e",
		Columns: []*types.PageColumn{
			col("started_time", "Started", true),
			col("agg_key", "Agg Key", true),
			col("flow_key", "Flow", true),
			col("scope_count", "Scope", true),
			col("bucket_count", "Bucket", true),
			col("top_count", "Top", true),
			col("finished_time", "Finished", true),
			col("duration_ms", "Duration", true),
		},
		SearchColumns: []string{"e.agg_key", "e.flow_key", "e.status", "e.error_message"},
		SortColumns: map[string]string{
			"id": "e.id", "agg_key": "e.agg_key", "flow_key": "e.flow_key", "status": "e.status",
			"scope_count": "e.scope_count", "bucket_count": "e.bucket_count", "top_count": "e.top_count",
			"duration_ms": "e.duration_ms", "error_message": "e.error_message", "started_time": "e.started_time", "finished_time": "e.finished_time",
		},
		OrderColumns: map[string]string{
			"started_time":  "e.started_time",
			"agg_key":       "e.agg_key",
			"flow_key":      "e.flow_key",
			"scope_count":   "e.scope_count",
			"bucket_count":  "e.bucket_count",
			"top_count":     "e.top_count",
			"finished_time": "e.finished_time",
			"duration_ms":   "e.duration_ms",
		},
		DefaultOrderBy:  "started_time",
		DefaultOrder:    "desc",
		DefaultPageSize: 50,
		Filters: []*types.PageFilter{
			textFilter("agg_key", "Agg Key", ""),
			textFilter("flow_key", "Flow Key", ""),
			textFilter("status", "Status", ""),
		},
		FilterRules: []*PageFilterRule{
			rule("agg_key", "e.agg_key", "=", "text"),
			rule("flow_key", "e.flow_key", "=", "text"),
			rule("status", "e.status", "=", "text"),
		},
	}
}

func eItemsPage() *PageConfig {
	return &PageConfig{
		Key:         "e-items",
		Title:       "E Items",
		Description: "外部条目扫描列表，覆盖 /e-items。",
		Group:       "List",
		LegacyPath:  "/e-items",
		Kind:        pageKindList,
		BaseSQL:     "FROM e_item i",
		Columns: []*types.PageColumn{
			col("id", "ID", true),
			linkCol("name", "Movie", "/movie/{name}", true),
			col("jav_id", "Jav ID", true),
			col("prefix", "Prefix", true),
			col("search_type", "Search Type", true),
			col("has_detail", "Detail", true),
			col("has_download_cover", "Cover", true),
			col("has_chinese", "Chinese", true),
			col("detail_need_scan", "Need Scan", true),
			col("detail_birth_time", "Detail Birth", true),
			col("last_query_detail_time", "Last Query", true),
		},
		SearchColumns: []string{"i.name", "i.jav_id", "i.prefix", "i.search_by"},
		SortColumns: map[string]string{
			"id": "i.id", "name": "i.name", "jav_id": "i.jav_id", "prefix": "i.prefix", "search_type": "i.search_type",
			"has_detail": "i.has_detail", "has_download_cover": "i.has_download_cover", "has_chinese": "i.has_chinese",
			"detail_need_scan": "i.detail_need_scan", "detail_birth_time": "i.detail_birth_time",
			"last_query_detail_time": "i.last_query_detail_time", "updated_on": "i.updated_on",
		},
		DefaultOrderBy:  "updated_on",
		DefaultOrder:    "desc",
		DefaultPageSize: 50,
		Filters: []*types.PageFilter{
			textFilter("jav_id", "Jav ID", ""),
			textFilter("name", "Name", ""),
			selectFilter("has_detail", "Detail", opt("All", ""), opt("None", "none"), opt("OK", "ok")),
			selectFilter("has_download_cover", "Cover", opt("All", ""), opt("None", "none"), opt("OK", "ok"), opt("Wrong", "wrong")),
			selectFilter("has_chinese", "Chinese", opt("All", ""), opt("None", "none"), opt("OK", "ok"), opt("Error", "error"), opt("Sensitive", "sensitive")),
			selectFilter("detail_need_scan", "Need Scan", opt("All", ""), opt("Need", "need"), opt("No Need", "no_need"), opt("Wrong", "wrong")),
			dateFilter("detail_birth_time_from", "Detail Birth From"),
			dateFilter("detail_birth_time_to", "Detail Birth To"),
			dateFilter("last_query_detail_time_from", "Last Query From"),
			dateFilter("last_query_detail_time_to", "Last Query To"),
		},
		FilterRules: []*PageFilterRule{
			rule("jav_id", "i.jav_id", "LIKE", "like"),
			rule("name", "i.name", "LIKE", "like"),
			rule("has_detail", "i.has_detail", "=", "item_detail_status"),
			rule("has_download_cover", "i.has_download_cover", "=", "item_cover_status"),
			rule("has_chinese", "i.has_chinese", "=", "item_chinese_status"),
			rule("detail_need_scan", "i.detail_need_scan", "=", "item_need_scan_status"),
			rule("detail_birth_time_from", "i.detail_birth_time", ">=", "date_start"),
			rule("detail_birth_time_to", "i.detail_birth_time", "<=", "date_end"),
			rule("last_query_detail_time_from", "i.last_query_detail_time", ">=", "date_start"),
			rule("last_query_detail_time_to", "i.last_query_detail_time", "<=", "date_end"),
		},
	}
}

func torrentAlbumsPage() *PageConfig {
	return &PageConfig{
		Key:         "torrent-albums",
		Title:       "Torrent Albums",
		Description: "磁力资源相册列表，覆盖 /torrent-albums 与 /album-items。",
		Group:       "List",
		LegacyPath:  "/torrent-albums",
		Kind:        pageKindList,
		BaseSQL:     "FROM tm_album_item i LEFT JOIN t_album a ON a.id = i.album_id",
		Columns: []*types.PageColumn{
			col("id", "ID", true),
			col("album_name", "Album", true),
			linkCol("movie_name", "Movie", "/movie/{movie_name}", true),
			col("source_type", "Source", true),
			col("info_hash", "Hash", true),
			col("size", "Size", true),
			col("publish_time", "Publish", true),
			col("created_on", "Created", true),
		},
		SearchColumns: []string{"a.name", "i.movie_name", "i.movie_jav_id", "i.info_hash", "i.source_type"},
		SortColumns: map[string]string{
			"id": "i.id", "album_name": "a.name", "movie_name": "i.movie_name", "source_type": "i.source_type",
			"info_hash": "i.info_hash", "size": "i.size", "publish_time": "i.publish_time", "created_on": "i.created_on",
		},
		OrderColumns: map[string]string{
			"movie_name":   "i.movie_name",
			"source_type":  "i.source_type",
			"publish_time": "i.publish_time",
			"created_on":   "i.created_on",
		},
		DefaultOrderBy:  "created_on",
		DefaultOrder:    "desc",
		DefaultPageSize: 50,
		Filters: []*types.PageFilter{
			textFilter("album_name", "Album", "下载中 / 待下载 / Media"),
			textFilter("source_type", "Source", "javbus_magnet / sukebei_torrent"),
			textFilter("info_hash", "Hash", ""),
			dateFilter("publish_from", "Publish From"),
			dateFilter("publish_to", "Publish To"),
			dateFilter("created_from", "Created From"),
			dateFilter("created_to", "Created To"),
		},
		FilterRules: []*PageFilterRule{
			rule("album_name", "a.name", "=", "text"),
			rule("album", "a.name", "=", "text"),
			rule("source_type", "i.source_type", "=", "text"),
			rule("info_hash", "i.info_hash", "LIKE", "like"),
			rule("publish_from", "i.publish_time", ">=", "date_start"),
			rule("publish_to", "i.publish_time", "<=", "date_end"),
			rule("created_from", "i.created_on", ">=", "date_start"),
			rule("created_to", "i.created_on", "<=", "date_end"),
		},
	}
}

func movieAlbumsPage() *PageConfig {
	return &PageConfig{
		Key:         "movie-albums",
		Title:       "Movie Albums",
		Description: "电影相册条目，覆盖 /movie-albums。",
		Group:       "List",
		LegacyPath:  "/movie-albums",
		Kind:        pageKindList,
		BaseSQL:     "FROM c_movie_album_item i LEFT JOIN c_movie_album a ON a.id = i.album_id",
		Columns: []*types.PageColumn{
			col("id", "ID", true),
			col("album_name", "Album", true),
			linkCol("movie_name", "Movie", "/movie/{movie_name}", true),
			col("movie_jav_id", "Jav ID", true),
			col("releasing_date", "Release", true),
			col("sort_no", "Sort", true),
			col("created_on", "Created", true),
		},
		SearchColumns: []string{"a.name", "i.movie_name", "i.movie_jav_id"},
		SortColumns: map[string]string{
			"id": "i.id", "album_name": "a.name", "movie_name": "i.movie_name", "movie_jav_id": "i.movie_jav_id",
			"releasing_date": "i.releasing_date", "sort_no": "i.sort_no", "created_on": "i.created_on",
		},
		OrderColumns: map[string]string{
			"movie_name":   "i.movie_name",
			"movie_jav_id": "i.movie_jav_id",
			"sort_no":      "i.sort_no",
			"created_on":   "i.created_on",
		},
		DefaultOrderBy:  "created_on",
		DefaultOrder:    "desc",
		DefaultPageSize: 50,
		Filters: []*types.PageFilter{
			textFilter("album_name", "Album", "相册名"),
		},
		FilterRules: []*PageFilterRule{
			rule("album_name", "a.name", "=", "text"),
			rule("album", "a.name", "=", "text"),
		},
	}
}

func dSeedsPage() *PageConfig {
	return &PageConfig{
		Key:         "d-seeds",
		Title:       "D Seeds",
		Description: "抓取种子配置列表，覆盖 /d-seeds。",
		Group:       "Crawler",
		LegacyPath:  "/d-seeds",
		Kind:        pageKindList,
		BaseSQL:     "FROM d_seed s",
		Columns: []*types.PageColumn{
			col("id", "ID", true),
			col("name", "Name", true),
			col("active", "Active", true),
			col("search_type", "Search", true),
			col("name_type", "Name Type", true),
			col("page_now", "Page", true),
			col("last_query_time", "Last Query", true),
			col("last_status", "Status", true),
			col("movie_total", "Movies", true),
			col("movie_latest_releasing_movie_name", "Latest Movie", true),
			col("movie_latest_releasing_date", "Latest Release", true),
			col("movie_last_added_time", "Last Added", true),
			col("last_insert_count", "Inserted", true),
			col("last_error", "Error", false),
		},
		SearchColumns: []string{"s.name", "s.last_error", "s.movie_latest_releasing_movie_name"},
		SortColumns: map[string]string{
			"id": "s.id", "name": "s.name", "active": "s.active", "search_type": "s.search_type", "page_now": "s.page_now",
			"name_type":   "s.name_type",
			"last_status": "s.last_status", "movie_total": "s.movie_total",
			"movie_latest_releasing_movie_name": "s.movie_latest_releasing_movie_name", "last_error": "s.last_error", "updated_on": "s.updated_on",
			"movie_latest_releasing_date": "s.movie_latest_releasing_date", "movie_last_added_time": "s.movie_last_added_time",
			"last_insert_count": "s.last_insert_count", "last_query_time": "s.last_query_time", "created_on": "s.created_on",
		},
		DefaultOrderBy:  "updated_on",
		DefaultOrder:    "desc",
		DefaultPageSize: 50,
		Filters: []*types.PageFilter{
			textFilter("name", "Name", ""),
			selectFilter("active", "Active", opt("All", ""), opt("启用", "2"), opt("停用", "1")),
			selectFilter("search_type", "Search", opt("All", ""), opt("Offset", "1"), opt("StartEnd", "2")),
			selectFilter("name_type", "Name Type", opt("All", ""), opt("Prefix", "1"), opt("Label", "2")),
			textFilter("last_status", "Status", "0/1/2/3"),
		},
		FilterRules: []*PageFilterRule{
			rule("name", "s.name", "LIKE", "like"),
			rule("active", "s.active", "=", "int"),
			rule("search_type", "s.search_type", "=", "int"),
			rule("name_type", "s.name_type", "=", "int"),
			rule("last_status", "s.last_status", "=", "int"),
		},
	}
}

func fetchJavbusPage() *PageConfig {
	return fetchStatusPage("fetch-site-javbus-list", "JavBus Fetch List", "/fetch-site-javbus-list", "FROM t_javbus_magnet_fetch f")
}

func fetchSukebeiPage() *PageConfig {
	return fetchStatusPage("fetch-site-sukebei-list", "Sukebei Fetch List", "/fetch-site-sukebei-list", "FROM t_sukebei_torrent_fetch f")
}

func fetchStatusPage(key, title, legacyPath, baseSQL string) *PageConfig {
	return &PageConfig{
		Key:         key,
		Title:       title,
		Description: "站点抓取状态列表，保留筛选、排序、分页和单条触发入口。",
		Group:       "FetchSite",
		LegacyPath:  legacyPath,
		Kind:        pageKindList,
		BaseSQL:     baseSQL,
		Columns: []*types.PageColumn{
			col("id", "ID", true),
			linkCol("movie_name", "Movie", "/movie/{movie_name}", true),
			col("movie_jav_id", "Jav ID", true),
			col("release_date", "Release", true),
			col("fetch_status", "Status", true),
			col("try_count", "Try", true),
			col("last_fetch_time", "Last Fetch", true),
			col("last_result_count", "Result", true),
			col("torrent_hash_count", "Hash Count", true),
			col("latest_publish_time", "Latest Publish", true),
			col("media_birth_time", "Media Birth", true),
			col("last_error", "Error", false),
		},
		SearchColumns: []string{"f.movie_name", "f.movie_jav_id", "f.last_error", "f.source_url"},
		SortColumns: map[string]string{
			"id": "f.id", "movie_name": "f.movie_name", "movie_jav_id": "f.movie_jav_id",
			"fetch_status": "f.fetch_status", "try_count": "f.try_count", "last_fetch_time": "f.last_fetch_time",
			"torrent_hash_count": "f.torrent_hash_count", "latest_publish_time": "f.latest_publish_time", "last_error": "f.last_error", "status": "f.fetch_status",
			"release_date": "f.release_date", "last_result_count": "f.last_result_count", "media_birth_time": fetchMediaBirthSQL(),
		},
		OrderColumns: map[string]string{
			"movie_name":          "f.movie_name",
			"release_date":        "f.release_date",
			"fetch_status":        "f.fetch_status",
			"last_fetch_time":     "f.last_fetch_time",
			"last_result_count":   "f.last_result_count",
			"torrent_hash_count":  "f.torrent_hash_count",
			"latest_publish_time": "f.latest_publish_time",
			"media_birth_time":    fetchMediaBirthSQL(),
		},
		DefaultOrderBy: "last_fetch_time",
		DefaultOrder:   "desc",
		Filters: []*types.PageFilter{
			selectFilter("status", "Status", opt("全部", ""), opt("待抓取", "1"), opt("抓取中", "2"), opt("成功", "3"), opt("失败", "4")),
			dateFilter("last_fetch_from", "Last Fetch From"),
			dateFilter("last_fetch_to", "Last Fetch To"),
			dateFilter("release_date_from", "Release From"),
			dateFilter("release_date_to", "Release To"),
			dateFilter("media_birth_from", "Media Birth From"),
			dateFilter("media_birth_to", "Media Birth To"),
		},
		FilterRules: []*PageFilterRule{
			rule("status", "f.fetch_status", "=", "int"),
			rule("last_fetch_from", "f.last_fetch_time", ">=", "date_start"),
			rule("last_fetch_to", "f.last_fetch_time", "<=", "date_end"),
			rule("release_date_from", "f.release_date", ">=", "date_start"),
			rule("release_date_to", "f.release_date", "<=", "date_end"),
			rule("media_birth_from", fetchMediaBirthSQL(), ">=", "date_start"),
			rule("media_birth_to", fetchMediaBirthSQL(), "<=", "date_end"),
		},
		DefaultPageSize: 50,
	}
}

func fetchMediaBirthSQL() string {
	return "(SELECT MAX(wm.birth_time) FROM w_media wm WHERE wm.movie_jav_id = f.movie_jav_id AND wm.source_type = 2)"
}

func fetchSehuatangPage() *PageConfig {
	return &PageConfig{
		Key:         "fetch-site-sehuatang-list",
		Title:       "Sehuatang Magnet List",
		Description: "色花堂磁力列表，覆盖 /fetch-site-sehuatang-list。",
		Group:       "FetchSite",
		LegacyPath:  "/fetch-site-sehuatang-list",
		Kind:        pageKindList,
		BaseSQL:     "FROM t_sehuatang_magnet m",
		Columns: []*types.PageColumn{
			col("id", "ID", true),
			linkCol("movie_name", "Movie", "/movie/{movie_name}", true),
			col("tag", "Tag", true),
			col("thread_title", "Title", true),
			linkCol("thread_url", "URL", "{thread_url}", false),
			col("info_hash", "Hash", true),
			col("post_time", "Post Time", true),
			col("post_date", "Post Date", true),
			col("last_seen_time", "Last Seen", true),
			col("created_on", "Created", true),
			col("updated_on", "Updated", true),
		},
		SearchColumns: []string{"m.movie_name", "m.movie_jav_id", "m.thread_title", "m.info_hash", "m.tag"},
		SortColumns: map[string]string{
			"id": "m.id", "movie_name": "m.movie_name", "tag": "m.tag", "thread_title": "m.thread_title",
			"thread_url": "m.thread_url", "info_hash": "m.info_hash", "post_time": "m.post_time", "last_seen_time": "m.last_seen_time",
			"movie_jav_id": "m.movie_jav_id", "post_date": "m.post_date", "created_on": "m.created_on", "updated_on": "m.updated_on",
		},
		OrderColumns: map[string]string{
			"movie_jav_id":   "m.movie_jav_id",
			"movie_name":     "m.movie_name",
			"post_time":      "m.post_time",
			"post_date":      "m.post_date",
			"last_seen_time": "m.last_seen_time",
			"created_on":     "m.created_on",
			"updated_on":     "m.updated_on",
		},
		DefaultOrderBy:  "post_time",
		DefaultOrder:    "desc",
		DefaultPageSize: 50,
		Filters: []*types.PageFilter{
			textFilter("movie_jav_id", "Jav ID", ""),
			textFilter("info_hash", "Hash", ""),
			textFilter("tag", "Tag", ""),
			dateFilter("post_time_from", "Post From"),
			dateFilter("post_time_to", "Post To"),
		},
		FilterRules: []*PageFilterRule{
			rule("movie_jav_id", "m.movie_jav_id", "LIKE", "like"),
			rule("info_hash", "m.info_hash", "LIKE", "like"),
			rule("tag", "m.tag", "=", "text"),
			rule("post_time_from", "m.post_time", ">=", "date_start"),
			rule("post_time_to", "m.post_time", "<=", "date_end"),
		},
	}
}

func wdirPage() *PageConfig {
	return &PageConfig{
		Key:         "wdir",
		Title:       "W Directory",
		Description: "媒体目录树列表，覆盖 /wdir 目录入口与详情参数。",
		Group:       "MovieAgg",
		LegacyPath:  "/wdir",
		Kind:        pageKindList,
		BaseSQL:     "FROM w_folder f",
		StaticWhere: "f.source_type = 2",
		Columns: []*types.PageColumn{
			linkCol("id", "ID", "/wdir?id={id}", true),
			col("parent_id", "Parent", true),
			col("name", "Name", true),
			col("depth", "Depth", true),
			col("path", "Path", true),
			col("source_type", "Source", true),
		},
		SearchColumns: []string{"f.name", "f.path", "f.path_hash"},
		SortColumns: map[string]string{
			"id": "f.id", "parent_id": "f.parent_id", "name": "f.name", "depth": "f.depth", "path": "f.path", "source_type": "f.source_type",
		},
		DefaultOrderBy:         "path",
		DefaultOrder:           "asc",
		DefaultPageSize:        60,
		DefaultFilters:         map[string]string{"parent_id": "0"},
		DefaultFiltersSkipWhen: map[string][]string{"parent_id": []string{"id"}},
		Filters: []*types.PageFilter{
			textFilter("id", "目录 ID", "留空查看根目录列表"),
			textFilter("parent_id", "Parent", ""),
			textFilter("source_type", "Source", ""),
		},
		FilterRules: []*PageFilterRule{
			rule("id", "f.id", "=", "int"),
			rule("parent_id", "f.parent_id", "=", "int"),
			rule("source_type", "f.source_type", "=", "int"),
		},
	}
}

func wMediaBirthPage() *PageConfig {
	return &PageConfig{
		Key:         "w-media-agg-birth",
		Title:       "WMedia Birth Agg",
		Description: "w_media 下载时间聚合总览，包含年份桶与 root Top 统计。",
		Group:       "MovieAgg",
		LegacyPath:  "/w-media-agg/birth",
		Kind:        pageKindList,
		BaseSQL:     wMediaBirthOverviewSQL(),
		Columns: []*types.PageColumn{
			hiddenCol("id"),
			hiddenCol("href"),
			col("section", "Section", true),
			linkCol("label", "Label", "{href}", true),
			col("media_count", "Media", true),
			col("size_bytes", "Size", true),
			col("latest_time", "Latest", true),
		},
		SearchColumns: []string{"x.section", "x.label", "x.scope_key"},
		SortColumns: map[string]string{
			"default":     "x.sort_group",
			"id":          "x.id",
			"href":        "x.href",
			"section":     "x.section",
			"label":       "x.label",
			"media_count": "x.media_count",
			"size_bytes":  "x.size_bytes",
			"latest_time": "x.latest_time",
			"rank_no":     "x.rank_no",
		},
		OrderColumns: map[string]string{
			"default":     "x.sort_group ASC, x.rank_no ASC, x.label",
			"section":     "x.section",
			"label":       "x.label",
			"media_count": "x.media_count",
			"size_bytes":  "x.size_bytes",
			"latest_time": "x.latest_time",
		},
		DefaultOrderBy:  "default",
		DefaultOrder:    "asc",
		DefaultPageSize: 120,
		Links: []*types.PageLink{
			link("Bucket List", "/w-media-agg/bucket-list", "related"),
			link("Agg Triggers", "/triggers/agg", "operation"),
		},
	}
}

func wMediaBirthBucketPage() *PageConfig {
	return &PageConfig{
		Key:         "w-media-agg-bucket-list",
		Title:       "WMedia Birth Buckets",
		Description: "w_media 下载时间桶聚合列表，覆盖 /w-media-agg/bucket-list。",
		Group:       "MovieAgg",
		LegacyPath:  "/w-media-agg/bucket-list",
		Kind:        pageKindList,
		BaseSQL:     "FROM w_media_birth_bucket_stat b",
		Columns: []*types.PageColumn{
			col("id", "ID", true),
			col("scope_key", "Scope", true),
			col("level", "Level", true),
			col("year", "Year", true),
			col("quarter", "Quarter", true),
			col("month", "Month", true),
			col("day", "Day", true),
			col("media_count", "Media", true),
			col("removed_count", "Removed", true),
			col("has_sub_count", "Sub", true),
			col("size_bytes", "Size", true),
		},
		SearchColumns: []string{"b.scope_key", "b.level"},
		SortColumns: map[string]string{
			"id": "b.id", "scope_key": "b.scope_key", "level": "b.level", "year": "b.year", "quarter": "b.quarter",
			"month": "b.month", "day": "b.day", "media_count": "b.media_count", "removed_count": "b.removed_count",
			"has_sub_count": "b.has_sub_count", "size_bytes": "b.size_bytes", "latest_birth_time": "b.latest_birth_time",
			"scope": "b.scope_key", "media": "b.media_count", "removed": "b.removed_count", "size": "b.size_bytes", "subtitle": "b.has_sub_count", "latest_birth": "b.latest_birth_time", "updated": "b.updated_on",
		},
		OrderColumns: map[string]string{
			"scope_key":    "b.scope_key",
			"scope":        "b.scope_key",
			"media":        "b.media_count",
			"media_count":  "b.media_count",
			"removed":      "b.removed_count",
			"size":         "b.size_bytes",
			"size_bytes":   "b.size_bytes",
			"subtitle":     "b.has_sub_count",
			"latest_birth": "b.latest_birth_time",
			"updated":      "b.updated_on",
		},
		DefaultOrderBy:  "scope",
		DefaultOrder:    "desc",
		DefaultPageSize: 100,
		DefaultFilters:  map[string]string{"level": "month"},
		Filters: []*types.PageFilter{
			selectFilter("level", "层级", opt("Year", "year"), opt("Quarter", "quarter"), opt("Month", "month"), opt("Day", "day")),
			textFilter("scope_key", "Scope", "2026-04"),
			numberFilter("year", "Year", ""),
			numberFilter("quarter", "Quarter", ""),
			numberFilter("month", "Month", ""),
			numberFilter("day", "Day", ""),
		},
		FilterRules: []*PageFilterRule{
			rule("level", "b.level", "=", "text"),
			rule("scope_key", "b.scope_key", "LIKE", "like"),
			rule("year", "b.year", "=", "int"),
			rule("quarter", "b.quarter", "=", "int"),
			rule("month", "b.month", "=", "int"),
			rule("day", "b.day", "=", "int"),
		},
		Links: []*types.PageLink{link("Agg Triggers", "/triggers/agg", "operation")},
	}
}

func movieReleaseAggPage() *PageConfig {
	return &PageConfig{
		Key:         "movie-agg-all-release",
		Title:       "Movie Release Agg",
		Description: "电影上映时间聚合页，按 legacy 支持 root/year/quarter/month/day 同页下钻与 Top tab。",
		Group:       "MovieAgg",
		LegacyPath:  "/movie-agg-all/release",
		Kind:        pageKindList,
		BaseSQL:     movieReleaseOverviewSQL(),
		Columns: []*types.PageColumn{
			hiddenCol("id"),
			hiddenCol("href"),
			hiddenCol("agg_mode"),
			hiddenCol("scope_key"),
			hiddenCol("level"),
			hiddenCol("page_level"),
			hiddenCol("year"),
			hiddenCol("quarter"),
			hiddenCol("month"),
			hiddenCol("day"),
			hiddenCol("range_start"),
			hiddenCol("range_end"),
			hiddenCol("rank_no"),
			hiddenCol("agg_type"),
			col("section", "Section", true),
			linkCol("label", "Label", "{href}", true),
			col("count_all", "All", true),
			col("count_owned", "Owned", true),
			col("size_bytes", "Size", true),
			col("latest_time", "Latest", true),
		},
		SearchColumns: []string{"x.section", "x.label", "x.scope_key"},
		SortColumns: map[string]string{
			"default":     "x.sort_group",
			"id":          "x.id",
			"href":        "x.href",
			"agg_mode":    "x.agg_mode",
			"scope_key":   "x.scope_key",
			"level":       "x.level",
			"page_level":  "x.page_level",
			"year":        "x.year",
			"quarter":     "x.quarter",
			"month":       "x.month",
			"day":         "x.day",
			"range_start": "x.range_start",
			"range_end":   "x.range_end",
			"agg_type":    "x.agg_type",
			"section":     "x.section",
			"label":       "x.label",
			"count_all":   "x.count_all",
			"count_owned": "x.count_owned",
			"size_bytes":  "x.size_bytes",
			"latest_time": "x.latest_time",
			"rank_no":     "x.rank_no",
		},
		OrderColumns: map[string]string{
			"default":     "x.sort_group ASC, x.sort_no ASC, x.label",
			"section":     "x.section",
			"label":       "x.label",
			"count_all":   "x.count_all",
			"count_owned": "x.count_owned",
			"size_bytes":  "x.size_bytes",
			"latest_time": "x.latest_time",
		},
		DefaultOrderBy:  "default",
		DefaultOrder:    "asc",
		DefaultPageSize: 200,
		DefaultFilters:  map[string]string{"agg_mode": "all"},
		Filters: []*types.PageFilter{
			selectFilter("agg_mode", "Mode", opt("全量", "all"), opt("已拥有", "owned")),
			numberFilter("year", "Year", ""),
			numberFilter("quarter", "Quarter", ""),
			numberFilter("month", "Month", ""),
			numberFilter("day", "Day", ""),
		},
		FilterRules: []*PageFilterRule{
			rule("agg_mode", "x.agg_mode", "=", "text"),
			rule("year", "x.year", "=", "int"),
			rule("quarter", "x.quarter", "=", "int"),
			rule("month", "x.month", "=", "int"),
			rule("day", "x.day", "=", "int"),
		},
		Links: []*types.PageLink{link("Bucket List", "/movie-agg-all/release-bucket-list", "related")},
	}
}

func movieReleaseBucketPage() *PageConfig {
	return &PageConfig{
		Key:         "movie-release-bucket-list",
		Title:       "Movie Release Buckets",
		Description: "电影上映时间桶列表，覆盖 /movie-agg-all/release-bucket-list。",
		Group:       "MovieAgg",
		LegacyPath:  "/movie-agg-all/release-bucket-list",
		Kind:        pageKindList,
		BaseSQL:     "FROM movie_release_bucket_stat b",
		Columns: []*types.PageColumn{
			hiddenCol("id"),
			col("scope_key", "时间桶", true),
			hiddenCol("level"),
			hiddenCol("year"),
			hiddenCol("quarter"),
			hiddenCol("month"),
			hiddenCol("day"),
			col("count_all", "All", true),
			col("count_owned", "Owned", true),
			col("size_bytes", "总大小", true),
			col("latest_releasing_date", "最新上映时间", true),
			hiddenCol("agg_mode"),
		},
		SearchColumns: []string{"b.scope_key", "b.level", "b.agg_mode"},
		SortColumns: map[string]string{
			"id": "b.id", "scope_key": "b.scope_key", "level": "b.level", "year": "b.year", "quarter": "b.quarter",
			"month": "b.month", "day": "b.day", "count_all": "b.count_all", "count_owned": "b.count_owned",
			"size_bytes": "b.size_bytes", "agg_mode": "b.agg_mode", "latest_releasing_date": "b.latest_releasing_date",
			"scope": "b.scope_key", "all": "b.count_all", "owned": "b.count_owned", "size": "b.size_bytes", "latest_release": "b.latest_releasing_date", "updated": "b.updated_on",
		},
		OrderColumns: map[string]string{
			"scope_key":             "b.year {order}, b.quarter {order}, b.month {order}, b.day {order}",
			"scope":                 "b.year {order}, b.quarter {order}, b.month {order}, b.day {order}",
			"all":                   "b.count_all",
			"count_all":             "b.count_all",
			"owned":                 "b.count_owned",
			"count_owned":           "b.count_owned",
			"size":                  "b.size_bytes",
			"size_bytes":            "b.size_bytes",
			"latest_release":        "b.latest_releasing_date",
			"latest_releasing_date": "b.latest_releasing_date",
			"updated":               "b.updated_on",
		},
		DefaultOrderBy:  "scope",
		DefaultOrder:    "desc",
		DefaultPageSize: 40,
		DefaultFilters:  map[string]string{"level": "month", "agg_mode": "all"},
		Filters: []*types.PageFilter{
			selectFilter("level", "层级", opt("Year", "year"), opt("Quarter", "quarter"), opt("Month", "month"), opt("Day", "day")),
			selectFilter("agg_mode", "Mode", opt("All", ""), opt("全量", "all"), opt("已拥有", "owned")),
			textFilter("scope_key", "Scope", "2026-04"),
			numberFilter("year", "Year", ""),
			numberFilter("quarter", "Quarter", ""),
			numberFilter("month", "Month", ""),
			numberFilter("day", "Day", ""),
		},
		FilterRules: []*PageFilterRule{
			rule("level", "b.level", "=", "text"),
			rule("agg_mode", "b.agg_mode", "=", "text"),
			rule("scope_key", "b.scope_key", "LIKE", "like"),
			rule("year", "b.year", "=", "int"),
			rule("quarter", "b.quarter", "=", "int"),
			rule("month", "b.month", "=", "int"),
			rule("day", "b.day", "=", "int"),
		},
	}
}

func wMediaBirthOverviewSQL() string {
	return `FROM (
SELECT
  CONCAT('year:', b.scope_key) AS id,
  CONCAT('/w-media-agg/birth?year=', b.year) AS href,
  1 AS sort_group,
  b.year AS rank_no,
  'Year' AS section,
  b.scope_key AS scope_key,
  CONCAT(b.year, ' 年') AS label,
  b.media_count AS media_count,
  b.size_bytes AS size_bytes,
  b.latest_birth_time AS latest_time
FROM w_media_birth_bucket_stat b
WHERE b.level = 'year'
UNION ALL
SELECT
  CONCAT(t.agg_type, ':', t.agg_key) AS id,
  CASE WHEN t.agg_type = 'cast' THEN CONCAT('/cast/', t.agg_id) ELSE '' END AS href,
  CASE t.agg_type WHEN 'cast' THEN 2 WHEN 'director' THEN 3 WHEN 'label' THEN 4 ELSE 5 END AS sort_group,
  t.rank_no AS rank_no,
  CASE t.agg_type WHEN 'cast' THEN 'Top Cast' WHEN 'director' THEN 'Top Director' WHEN 'label' THEN 'Top Label' ELSE 'Top Prefix' END AS section,
  t.scope_key AS scope_key,
  t.agg_name AS label,
  t.media_count AS media_count,
  t.size_bytes AS size_bytes,
  0 AS latest_time
FROM w_media_birth_top_stat t
WHERE t.scope_key = 'root' AND t.rank_no <= 30
) x`
}

func movieReleaseOverviewSQL() string {
	return `FROM (
SELECT
  CONCAT(b.agg_mode, ':year:', b.scope_key) AS id,
  CASE WHEN b.agg_mode = 'owned'
    THEN CONCAT('/movie-agg-all/release?agg_mode=owned&year=', b.year)
    ELSE CONCAT('/movie-agg-all/release?year=', b.year)
  END AS href,
  b.agg_mode AS agg_mode,
  b.scope_key AS scope_key,
  b.level AS level,
  'root' AS page_level,
  b.year AS year,
  b.quarter AS quarter,
  b.month AS month,
  b.day AS day,
  '' AS range_start,
  '' AS range_end,
  '' AS agg_type,
  1 AS sort_group,
  b.year AS rank_no,
  -b.year AS sort_no,
  'Year' AS section,
  CONCAT(b.year, ' 年') AS label,
  b.count_all AS count_all,
  b.count_owned AS count_owned,
  b.size_bytes AS size_bytes,
  b.latest_releasing_date AS latest_time
FROM movie_release_bucket_stat b
WHERE b.level = 'year' AND b.agg_mode IN ('all', 'owned')
UNION ALL
SELECT
  CONCAT(t.agg_mode, ':', t.agg_type, ':', t.agg_key) AS id,
  CASE t.agg_type
    WHEN 'cast' THEN CASE WHEN t.agg_id > 0 THEN CONCAT('/cards?pid=', t.agg_id, '&mowned=', CASE WHEN t.agg_mode = 'owned' THEN '3' ELSE '1' END) ELSE CONCAT('/cards?cn=', t.agg_name, '&mowned=', CASE WHEN t.agg_mode = 'owned' THEN '3' ELSE '1' END) END
    WHEN 'director' THEN CONCAT('/cards?dn=', t.agg_name, '&mowned=', CASE WHEN t.agg_mode = 'owned' THEN '3' ELSE '1' END)
    WHEN 'label' THEN CONCAT('/cards?ln=', t.agg_name, '&mowned=', CASE WHEN t.agg_mode = 'owned' THEN '3' ELSE '1' END)
    ELSE CONCAT('/cards?pn=', t.agg_name, '&mowned=', CASE WHEN t.agg_mode = 'owned' THEN '3' ELSE '1' END)
  END AS href,
  t.agg_mode AS agg_mode,
  t.scope_key AS scope_key,
  t.level AS level,
  'root' AS page_level,
  t.year AS year,
  t.quarter AS quarter,
  t.month AS month,
  0 AS day,
  '' AS range_start,
  '' AS range_end,
  t.agg_type AS agg_type,
  CASE t.agg_type WHEN 'cast' THEN 2 WHEN 'director' THEN 3 WHEN 'label' THEN 4 ELSE 5 END AS sort_group,
  t.rank_no AS rank_no,
  t.rank_no AS sort_no,
  CASE t.agg_type WHEN 'cast' THEN 'Top Cast' WHEN 'director' THEN 'Top Director' WHEN 'label' THEN 'Top Label' ELSE 'Top Prefix' END AS section,
  t.agg_name AS label,
  t.count_all AS count_all,
  t.count_owned AS count_owned,
  0 AS size_bytes,
  0 AS latest_time
FROM movie_release_top_stat t
WHERE t.scope_key = 'root' AND t.agg_mode IN ('all', 'owned') AND t.rank_no <= 30
) x`
}

func crawlerTasksPage() *PageConfig {
	return operationPage("crawler-tasks", "Crawler Tasks", "/crawler/tasks", "Crawler", "爬虫任务列表入口，保留任务监控与跳转。")
}

func triggersPage() *PageConfig {
	return operationPage("triggers", "Crawler Jobs", "/triggers", "Crawler", "爬虫任务总入口，对应 legacy JobsPage。")
}

func crawlerDetailLoopPage() *PageConfig {
	return operationPage("crawler-detail-loop", "Crawler Detail Loop", "/crawler/detail-loop", "Crawler", "详情循环任务控制页。")
}

func dailyBestTriggerPage() *PageConfig {
	return operationPage("triggers-dailybest", "DailyBest Trigger", "/triggers/dailybest", "Crawler", "DailyBest 抓取任务入口。")
}

func seedsTriggerPage() *PageConfig {
	return operationPage("triggers-seeds", "Seeds Trigger", "/triggers/seeds", "Crawler", "Seeds 抓取任务入口。")
}

func postProcessTriggerPage() *PageConfig {
	return operationPage("triggers-post-process", "Post Process Trigger", "/triggers/post-process", "Crawler", "抓取后处理任务入口。")
}

func backfillTriggerPage() *PageConfig {
	return operationPage("triggers-backfill", "Backfill Trigger", "/triggers/backfill", "Crawler", "回填任务入口。")
}

func fetchSiteTriggerPage() *PageConfig {
	return operationPage("triggers-fetch-site", "FetchSite Trigger", "/triggers/fetch-site", "FetchSite", "站点抓取任务入口。")
}

func fetchSehuatangTriggerPage() *PageConfig {
	return operationPage("triggers-fetch-sehuatang", "FetchSehuatang Trigger", "/triggers/fetch-sehuatang", "FetchSite", "色花堂抓取任务入口。")
}

func mediaTriggerPage() *PageConfig {
	return operationPage("triggers-media", "Media Ingest", "/triggers/media", "Media", "媒体入库预处理与提交入口。")
}

func mediaRescanPage() *PageConfig {
	return operationPage("triggers-media-rescan", "Media Rescan", "/triggers/media-rescan", "Media", "媒体元数据重新扫描入口。")
}

func mediaRollbackPage() *PageConfig {
	return operationPage("triggers-media-rollback", "Media Rollback", "/triggers/media-rollback", "Media", "媒体入库回滚入口。")
}

func scMediaMovePage() *PageConfig {
	return operationPage("triggers-sc-media", "SC Media Move", "/triggers/sc-media", "SC", "SC 媒体移动入口。")
}

func scPickSmartPage() *PageConfig {
	return operationPage("sc-pick-smart-media", "SC Pick Smart Media", "/sc-pick-smart-media", "SC", "智能抽片与复制入口。")
}

func scTriggersPage() *PageConfig {
	return operationPage("triggers-sc", "SC Preview", "/triggers/sc", "SC", "SC Add Preview 与提交入口。")
}

func aggTriggersPage() *PageConfig {
	return operationPage("triggers-agg", "Agg Triggers", "/triggers/agg", "MovieAgg", "聚合回填入口。")
}

func operationPage(key, title, legacyPath, group, description string) *PageConfig {
	return &PageConfig{
		Key:         key,
		Title:       title,
		Description: description,
		Group:       group,
		LegacyPath:  legacyPath,
		Kind:        pageKindOperation,
		Columns:     []*types.PageColumn{},
		SortColumns: map[string]string{},
		Actions:     []*types.PageAction{},
		Links: []*types.PageLink{
			link("Crawler Tasks", "/crawler/tasks", "related"),
			link("WAgg Events", "/w-agg-events", "related"),
		},
	}
}
