package modelx

import "rudy-gc-api/internal/types"

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
			linkCol("id", "ID", "/cast/{id}", true),
			col("name", "Name", true),
			col("chinese", "Chinese", true),
			col("movie_number", "Movies", true),
			col("owned_movie_number", "Owned", true),
			col("owned_w_media_number", "WMedia", true),
			col("sc_times", "SC", true),
			col("come_times", "Come", true),
			col("last_sc_time", "Last SC", true),
			col("highest_rank", "Highest Rank", true),
		},
		SearchColumns: []string{"p.name", "p.chinese", "p.alias"},
		SortColumns: map[string]string{
			"id": "p.id", "name": "p.name", "chinese": "p.chinese", "movie_number": "p.movie_number",
			"owned_movie_number": "p.owned_movie_number", "owned_w_media_number": "p.owned_w_media_number",
			"sc_times": "p.sc_times", "come_times": "p.come_times", "last_sc_time": "p.last_sc_time",
			"highest_rank": "p.highest_rank",
		},
		DefaultOrderBy: "owned_w_media_number",
		DefaultOrder:   "desc",
		Links:          []*types.PageLink{link("SC Events", "/sc-events", "related"), link("Movie Cards", "/cards", "related")},
	}
}

func castDetailPage() *PageConfig {
	cfg := castsPage()
	cfg.Key = "cast-detail"
	cfg.Title = "Cast Detail"
	cfg.Description = "演员详情基础信息，前端会联动 cards 的 pid 筛选展示影片。"
	cfg.LegacyPath = "/cast/:id"
	cfg.Filters = []*types.PageFilter{textFilter("id", "演员 ID", "例如 1")}
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
			col("id", "ID", true),
			col("name", "Name", true),
			col("type", "Type", true),
			col("start_time", "Start", true),
			col("end_time", "End", true),
			col("detail_number", "Detail", true),
		},
		SearchColumns: []string{"r.name", "r.type"},
		SortColumns: map[string]string{
			"id": "r.id", "name": "r.name", "type": "r.type", "start_time": "r.start_time",
			"end_time": "r.end_time", "detail_number": "r.detail_number",
		},
		DefaultOrderBy: "start_time",
		DefaultOrder:   "desc",
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
			linkCol("name", "Name", "/sc-events/{name}", true),
			linkCol("come_movie_name", "Come Movie", "/movie/{come_movie_name}", true),
			col("movie_cast", "Movie Cast", true),
			col("sc_time", "SC Time", true),
			col("cooldown", "Cooldown", true),
			col("duration", "Duration", true),
			col("kind", "Kind", true),
			col("remarks", "Remarks", false),
		},
		SearchColumns: []string{"s.name", "s.come_movie_name", "s.movie_cast", "s.remarks", "s.kind"},
		SortColumns: map[string]string{
			"id": "s.id", "name": "s.name", "come_movie_name": "s.come_movie_name", "movie_cast": "s.movie_cast",
			"sc_time": "s.sc_time", "cooldown": "s.cooldown", "duration": "s.duration", "kind": "s.kind", "remarks": "s.remarks",
		},
		DefaultOrderBy: "sc_time",
		DefaultOrder:   "desc",
		Links:          []*types.PageLink{link("Smart Pick", "/sc-pick-smart-media", "operation"), link("SC Preview", "/triggers/sc", "operation")},
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
		BaseSQL:     "FROM w_media m",
		Columns: []*types.PageColumn{
			col("id", "ID", true),
			linkCol("movie_name", "Movie", "/movie/{movie_name}", true),
			col("file_name", "File", true),
			col("full_dir", "Directory", true),
			col("size", "Size", true),
			col("height", "Height", true),
			col("duration", "Duration", true),
			col("birth_time", "Birth", true),
			col("has_sub", "Sub", true),
			col("is_removed", "Removed", true),
		},
		SearchColumns: []string{"m.movie_name", "m.movie_jav_id", "m.file_name", "m.full_dir", "m.source_torrent_hash"},
		SortColumns: map[string]string{
			"id": "m.id", "movie_name": "m.movie_name", "file_name": "m.file_name", "full_dir": "m.full_dir",
			"size": "m.size", "height": "m.height", "duration": "m.duration", "birth_time": "m.birth_time",
			"has_sub": "m.has_sub", "is_removed": "m.is_removed",
		},
		DefaultOrderBy: "birth_time",
		DefaultOrder:   "desc",
		Links:          []*types.PageLink{link("Media Ingest", "/triggers/media", "operation"), link("Media Rescan", "/triggers/media-rescan", "operation")},
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
			col("id", "ID", true),
			col("agg_key", "Agg Key", true),
			col("flow_key", "Flow", true),
			col("status", "Status", true),
			col("scope_count", "Scope", true),
			col("bucket_count", "Bucket", true),
			col("top_count", "Top", true),
			col("duration_ms", "Duration", true),
			col("error_message", "Error", false),
		},
		SearchColumns: []string{"e.agg_key", "e.flow_key", "e.status", "e.error_message"},
		SortColumns: map[string]string{
			"id": "e.id", "agg_key": "e.agg_key", "flow_key": "e.flow_key", "status": "e.status",
			"scope_count": "e.scope_count", "bucket_count": "e.bucket_count", "top_count": "e.top_count",
			"duration_ms": "e.duration_ms", "error_message": "e.error_message",
		},
		DefaultOrderBy: "id",
		DefaultOrder:   "desc",
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
			col("has_chinese", "Chinese", true),
			col("detail_need_scan", "Need Scan", true),
			col("last_query_detail_time", "Last Query", true),
		},
		SearchColumns: []string{"i.name", "i.jav_id", "i.prefix", "i.search_by"},
		SortColumns: map[string]string{
			"id": "i.id", "name": "i.name", "jav_id": "i.jav_id", "prefix": "i.prefix", "search_type": "i.search_type",
			"has_detail": "i.has_detail", "has_chinese": "i.has_chinese", "detail_need_scan": "i.detail_need_scan",
			"last_query_detail_time": "i.last_query_detail_time", "updated_on": "i.updated_on",
		},
		DefaultOrderBy: "updated_on",
		DefaultOrder:   "desc",
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
		},
		SearchColumns: []string{"a.name", "i.movie_name", "i.movie_jav_id", "i.info_hash", "i.source_type"},
		SortColumns: map[string]string{
			"id": "i.id", "album_name": "a.name", "movie_name": "i.movie_name", "source_type": "i.source_type",
			"info_hash": "i.info_hash", "size": "i.size", "publish_time": "i.publish_time",
		},
		DefaultOrderBy: "id",
		DefaultOrder:   "desc",
		Filters:        []*types.PageFilter{textFilter("album", "相册", "下载中 / 待下载 / Media")},
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
		},
		SearchColumns: []string{"a.name", "i.movie_name", "i.movie_jav_id"},
		SortColumns: map[string]string{
			"id": "i.id", "album_name": "a.name", "movie_name": "i.movie_name", "movie_jav_id": "i.movie_jav_id",
			"releasing_date": "i.releasing_date", "sort_no": "i.sort_no",
		},
		DefaultOrderBy: "sort_no",
		DefaultOrder:   "asc",
		Filters:        []*types.PageFilter{textFilter("album", "相册", "相册名")},
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
			col("page_now", "Page", true),
			col("last_status", "Status", true),
			col("movie_total", "Movies", true),
			col("movie_latest_releasing_movie_name", "Latest Movie", true),
			col("last_error", "Error", false),
		},
		SearchColumns: []string{"s.name", "s.last_error", "s.movie_latest_releasing_movie_name"},
		SortColumns: map[string]string{
			"id": "s.id", "name": "s.name", "active": "s.active", "search_type": "s.search_type", "page_now": "s.page_now",
			"last_status": "s.last_status", "movie_total": "s.movie_total",
			"movie_latest_releasing_movie_name": "s.movie_latest_releasing_movie_name", "last_error": "s.last_error", "updated_on": "s.updated_on",
		},
		DefaultOrderBy: "updated_on",
		DefaultOrder:   "desc",
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
			col("fetch_status", "Status", true),
			col("try_count", "Try", true),
			col("last_fetch_time", "Last Fetch", true),
			col("torrent_hash_count", "Hash Count", true),
			col("latest_publish_time", "Latest Publish", true),
			col("last_error", "Error", false),
		},
		SearchColumns: []string{"f.movie_name", "f.movie_jav_id", "f.last_error", "f.source_url"},
		SortColumns: map[string]string{
			"id": "f.id", "movie_name": "f.movie_name", "movie_jav_id": "f.movie_jav_id",
			"fetch_status": "f.fetch_status", "try_count": "f.try_count", "last_fetch_time": "f.last_fetch_time",
			"torrent_hash_count": "f.torrent_hash_count", "latest_publish_time": "f.latest_publish_time", "last_error": "f.last_error", "status": "f.fetch_status",
		},
		DefaultOrderBy: "last_fetch_time",
		DefaultOrder:   "desc",
		Filters: []*types.PageFilter{
			selectFilter("status", "状态", opt("全部", ""), opt("待抓取", "1"), opt("抓取中", "2"), opt("成功", "3"), opt("失败", "4")),
		},
	}
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
			col("thread_url", "URL", false),
			col("info_hash", "Hash", true),
			col("post_time", "Post Time", true),
			col("last_seen_time", "Last Seen", true),
		},
		SearchColumns: []string{"m.movie_name", "m.movie_jav_id", "m.thread_title", "m.info_hash", "m.tag"},
		SortColumns: map[string]string{
			"id": "m.id", "movie_name": "m.movie_name", "tag": "m.tag", "thread_title": "m.thread_title",
			"thread_url": "m.thread_url", "info_hash": "m.info_hash", "post_time": "m.post_time", "last_seen_time": "m.last_seen_time",
		},
		DefaultOrderBy: "post_time",
		DefaultOrder:   "desc",
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
		DefaultOrderBy: "path",
		DefaultOrder:   "asc",
		Filters:        []*types.PageFilter{textFilter("id", "目录 ID", "留空查看根目录列表")},
	}
}

func wMediaBirthPage() *PageConfig {
	cfg := wMediaBirthBucketPage()
	cfg.Key = "w-media-agg-birth"
	cfg.Title = "WMedia Birth Agg"
	cfg.LegacyPath = "/w-media-agg/birth"
	cfg.Description = "w_media 下载时间聚合总览。"
	return cfg
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
		},
		DefaultOrderBy: "latest_birth_time",
		DefaultOrder:   "desc",
		Filters:        []*types.PageFilter{textFilter("level", "层级", "year / quarter / month / day")},
		Links:          []*types.PageLink{link("Agg Triggers", "/triggers/agg", "operation")},
	}
}

func movieReleaseAggPage() *PageConfig {
	cfg := movieReleaseBucketPage()
	cfg.Key = "movie-agg-all-release"
	cfg.Title = "Movie Release Agg"
	cfg.LegacyPath = "/movie-agg-all/release"
	cfg.Description = "电影上映时间聚合总览。"
	return cfg
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
			col("id", "ID", true),
			col("scope_key", "Scope", true),
			col("level", "Level", true),
			col("year", "Year", true),
			col("quarter", "Quarter", true),
			col("month", "Month", true),
			col("day", "Day", true),
			col("count_all", "All", true),
			col("count_owned", "Owned", true),
			col("size_bytes", "Size", true),
			col("agg_mode", "Mode", true),
		},
		SearchColumns: []string{"b.scope_key", "b.level", "b.agg_mode"},
		SortColumns: map[string]string{
			"id": "b.id", "scope_key": "b.scope_key", "level": "b.level", "year": "b.year", "quarter": "b.quarter",
			"month": "b.month", "day": "b.day", "count_all": "b.count_all", "count_owned": "b.count_owned",
			"size_bytes": "b.size_bytes", "agg_mode": "b.agg_mode", "latest_releasing_date": "b.latest_releasing_date",
		},
		DefaultOrderBy: "latest_releasing_date",
		DefaultOrder:   "desc",
		Filters:        []*types.PageFilter{textFilter("level", "层级", "year / quarter / month / day")},
	}
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
