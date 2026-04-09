package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/service/loop"
	"rudy_gc/internal/service/media"
	"rudy_gc/internal/service/sehuatang"
	"rudy_gc/internal/service/wkv"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"
)

type CrawlerPages struct {
	runtime   *loop.FetchLoopService
	fetchSite *fetchsite.Service
	mediaSvc  *media.Service
	sehuatang *sehuatang.Service
	wkvSvc    *wkv.Service
	deps      *svc.Deps
	scRootDir string
}

type crawlerJobsPageTask struct {
	TaskType string
	Label    string
}

type crawlerJobsPageLabels struct {
	Extra   string
	Result  string
	Elapsed string
}

type crawlerJobsPageLink struct {
	Href  string
	Label string
}

type crawlerJobsPageConfig struct {
	Title                     string
	PageTitle                 string
	QuickNavCurrent           string
	TaskPanelTitle            string
	TaskHelperText            string
	PageNote                  string
	TaskTableTitle            string
	EventTitle                string
	StorageKey                string
	DefaultTaskType           string
	OverviewExtraMode         string
	EmptyStateText            string
	AllowedTaskTypes          []string
	TaskButtons               []crawlerJobsPageTask
	HeaderLinks               []crawlerJobsPageLink
	Labels                    crawlerJobsPageLabels
	HideTaskForm              bool
	MovieCardFilter           *movieCardFilterView
	ShowMovieFilters          bool
	ShowFavoriteAlbum         bool
	FavoriteRows              []*favoriteAlbumItemRow
	FetchSehuatangList        string
	FetchSehuatangKey         string
	FetchSehuatangStartPage   int64
	FetchSehuatangEndPage     int64
	FetchSehuatangPersistMode string
}

type favoriteAlbumItemRow struct {
	MovieName   string
	MovieJavID  string
	SourceType  string
	SourceLabel string
	Size        string
}

const (
	crawlerFavoriteAlbumName  = "下载中"
	crawlerFavoriteCardMaxRow = 50
)

const taskSpiderFetchSiteBoth = "spider_fetch_site_both_resources"

func NewCrawlerPages(deps *svc.Deps) *CrawlerPages {
	return &CrawlerPages{
		runtime:   newCrawlerRuntime(deps),
		fetchSite: newFetchSitePageService(deps),
		mediaSvc:  media.NewService(deps),
		sehuatang: sehuatang.NewService(deps),
		wkvSvc:    wkv.NewService(deps),
		deps:      deps,
		scRootDir: deps.Config.Film.ScRootDir,
	}
}

func (h *CrawlerPages) JobsPage(c *gin.Context) {
	target := "/triggers/dailybest"
	jobID := strings.TrimSpace(c.Query("job_id"))
	if jobID != "" {
		target += "?job_id=" + url.QueryEscape(jobID)
	}
	c.Redirect(http.StatusFound, target)
}

func (h *CrawlerPages) DailyBestPage(c *gin.Context) {
	h.renderJobsPage(c, crawlerJobsPageConfig{
		Title:             "DailyBest 任务",
		PageTitle:         "DailyBest 抓取",
		QuickNavCurrent:   "dailybest",
		TaskPanelTitle:    "DailyBest 触发",
		PageNote:          "只处理 DailyBest 抓取与同步任务。",
		TaskTableTitle:    "DailyBest 任务",
		EventTitle:        "DailyBest 事件流",
		StorageKey:        "crawler_jobs_dailybest_selected_job",
		DefaultTaskType:   loop.TaskSpiderDailyBest,
		OverviewExtraMode: "dailybest_stages",
		EmptyStateText:    "等待 DailyBest 任务触发",
		AllowedTaskTypes: []string{
			loop.TaskSpiderDailyBest,
			loop.TaskSpiderDailyBestSync,
		},
		TaskButtons: []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderDailyBest, Label: "DailyBest 抓取"},
			{TaskType: loop.TaskSpiderDailyBestSync, Label: "DailyBest 同步"},
		},
		Labels: crawlerJobsPageLabels{
			Extra:   "当前页数",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
	})
}

func (h *CrawlerPages) SeedsPage(c *gin.Context) {
	h.renderJobsPage(c, crawlerJobsPageConfig{
		Title:             "Seeds 任务",
		PageTitle:         "活跃 Seeds / 按名称抓取 / 刷新最久详情",
		TaskPanelTitle:    "Seeds 抓取",
		PageNote:          "这里只处理活跃 Seeds、按名称抓取和刷新最久详情。",
		TaskTableTitle:    "Seeds 任务",
		EventTitle:        "Seeds 事件流",
		StorageKey:        "crawler_jobs_seeds_selected_job",
		DefaultTaskType:   loop.TaskSpiderSeeds,
		OverviewExtraMode: "seeds_stages",
		EmptyStateText:    "等待 Seeds 任务触发",
		AllowedTaskTypes: []string{
			loop.TaskSpiderSeeds,
			loop.TaskSpiderSeedByName,
			loop.TaskSpiderRefreshOldest,
		},
		TaskButtons: []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderSeeds, Label: "活跃 Seeds"},
			{TaskType: loop.TaskSpiderSeedByName, Label: "按名称抓取"},
			{TaskType: loop.TaskSpiderRefreshOldest, Label: "刷新最久详情"},
		},
		Labels: crawlerJobsPageLabels{
			Extra:   "详情 Loop",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
	})
}

func (h *CrawlerPages) FilmPage(c *gin.Context) {
	h.renderJobsPage(c, crawlerJobsPageConfig{
		Title:             "影片任务",
		PageTitle:         "影片重命名 / 影片处理",
		TaskPanelTitle:    "影片任务",
		PageNote:          "这里只处理影片重命名和影片处理。",
		TaskTableTitle:    "影片任务",
		EventTitle:        "影片事件流",
		StorageKey:        "crawler_jobs_film_selected_job",
		DefaultTaskType:   loop.TaskFilmRename,
		OverviewExtraMode: "fetch_site_summary",
		EmptyStateText:    "等待影片任务触发",
		AllowedTaskTypes: []string{
			loop.TaskFilmRename,
			loop.TaskFilmProcess,
		},
		TaskButtons: []crawlerJobsPageTask{
			{TaskType: loop.TaskFilmRename, Label: "影片重命名"},
			{TaskType: loop.TaskFilmProcess, Label: "影片处理"},
		},
		Labels: crawlerJobsPageLabels{
			Extra:   "任务类型",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
	})
}

func (h *CrawlerPages) PostProcessPage(c *gin.Context) {
	h.renderJobsPage(c, crawlerJobsPageConfig{
		Title:             "后处理任务",
		PageTitle:         "图片抓取 / 标题翻译",
		TaskPanelTitle:    "后处理任务",
		PageNote:          "单独触发封面抓取、标题翻译，或两者一起执行。这里不重跑榜单、种子和详情抓取。",
		TaskTableTitle:    "后处理任务",
		EventTitle:        "后处理事件流",
		StorageKey:        "crawler_jobs_post_process_selected_job",
		DefaultTaskType:   loop.TaskSpiderPostProcess,
		OverviewExtraMode: "post_process_stages",
		EmptyStateText:    "等待后处理任务触发",
		AllowedTaskTypes: []string{
			loop.TaskSpiderDownloadCover,
			loop.TaskSpiderTranslateTitle,
			loop.TaskSpiderPostProcess,
		},
		TaskButtons: []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderDownloadCover, Label: "图片抓取"},
			{TaskType: loop.TaskSpiderTranslateTitle, Label: "标题翻译"},
			{TaskType: loop.TaskSpiderPostProcess, Label: "图片 + 标题翻译"},
		},
		Labels: crawlerJobsPageLabels{
			Extra:   "当前阶段",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
	})
}

func (h *CrawlerPages) MediaPage(c *gin.Context) {
	favoriteRows, _ := h.loadFavoriteAlbumRows(c.Request.Context())
	rootDirs := h.deps.Config.Media.RootDirs
	c.HTML(http.StatusOK, "page.media_ingest", gin.H{
		"Title":        "媒体入库",
		"PageTitle":    "媒体入库（两段执行）",
		"PageNote":     "第一段只做预处理校验与插入前预览；第二段仅插入通过项。",
		"RootDirs":     rootDirs,
		"FavoriteRows": favoriteRows,
	})
}

func (h *CrawlerPages) MediaRollbackPage(c *gin.Context) {
	rootDirs := h.deps.Config.Media.RootDirs
	c.HTML(http.StatusOK, "page.media_rollback", gin.H{
		"Title":     "媒体回滚",
		"PageTitle": "媒体回滚（名称还原）",
		"PageNote":  "将 004_rollback 中的文件名还原后移回 001_ingest_new。",
		"RootDirs":  rootDirs,
	})
}

func (h *CrawlerPages) ScMediaMovePage(c *gin.Context) {
	scName := strings.TrimSpace(c.Query("sc_name"))
	rootDirs := h.deps.Config.Media.RootDirs
	scDetailHref := ""
	if scName != "" {
		scDetailHref = "/sc-events/" + url.PathEscape(scName)
	}

	c.HTML(http.StatusOK, "page.sc_media_move", gin.H{
		"Title":        "SC Media 移动",
		"PageTitle":    "SC Media 移动（两段执行）",
		"PageNote":     "第一段只做预处理预览；第二段仅移动预处理通过的 WMedia 到 watched 目录。",
		"RootDirs":     rootDirs,
		"ScName":       scName,
		"ScDetailHref": scDetailHref,
	})
}

func (h *CrawlerPages) FilmMovePage(c *gin.Context) {
	base := types.ListMovieFullRequest{
		OrderBy:  consts.OrderByReleasingDate,
		Page:     1,
		PageSize: maxPageSize,
	}
	req, err := parseMovieCardRequest(c, base)
	if err != nil {
		req = base
		req.Order = normalizeOrder(c.Query("order"))
	}

	filterView := buildMovieCardFilterView(c, req, req.OrderBy, nil)
	filterView.Action = "/triggers/film-move"
	filterView.ClearHref = "/triggers/film-move"

	c.HTML(http.StatusOK, "page.film_move", gin.H{
		"Title":           "影片批量移动",
		"PageTitle":       "影片批量移动（两步执行）",
		"PageNote":        "第一步按 Cards 同款筛选列出影片；第二步将第一步结果移动到 SC Movie 同款目标目录。",
		"MovieCardFilter": filterView,
	})
}

func (h *CrawlerPages) MediaRescanPage(c *gin.Context) {
	rootOptions, err := h.mediaSvc.ListLibraryRescanRootOptions(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "加载媒体重扫页面失败: %v", err)
		return
	}
	c.HTML(http.StatusOK, "page.media_rescan", gin.H{
		"Title":       "媒体重扫",
		"PageTitle":   "媒体重扫（位置与删除状态同步）",
		"PageNote":    "同一个 root 下可分别勾选 media 与 watched；若勾选二级目录，则只扫描该子树。未挂载或不存在的范围会直接跳过，并保持现有 w_media 不变。这里也可以手动回填旧的 WMedia 时间聚合数据和上映日时间聚合数据。",
		"RootOptions": rootOptions,
	})
}

func (h *CrawlerPages) FetchSitePage(c *gin.Context) {
	req, err := parseMovieCardRequest(c, types.ListMovieFullRequest{
		OrderBy: consts.OrderByReleasingDate,
	})
	if err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}
	filterView := buildMovieCardFilterView(c, req, req.OrderBy, nil)
	filterView.HideDirs = true
	filterView.TextDateInputs = true
	h.renderJobsPage(c, crawlerJobsPageConfig{
		Title:             "抓取站点任务",
		PageTitle:         "JavBus / Sukebei 资源抓取",
		TaskPanelTitle:    "抓取站点任务",
		PageNote:          "JavBus / Sukebei 抓取目标由 cards 同源筛选决定，数量参数控制本次最多处理条数。",
		TaskTableTitle:    "抓取站点任务",
		EventTitle:        "抓取站点事件流",
		StorageKey:        "crawler_jobs_fetch_site_selected_job",
		DefaultTaskType:   loop.TaskSpiderFetchJavbus,
		OverviewExtraMode: "fetch_site_summary",
		EmptyStateText:    "等待抓取站点任务触发",
		AllowedTaskTypes: []string{
			loop.TaskSpiderFetchJavbus,
			loop.TaskSpiderFetchSukebei,
		},
		TaskButtons: []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderFetchJavbus, Label: "JavBus 资源抓取"},
			{TaskType: loop.TaskSpiderFetchSukebei, Label: "Sukebei 资源抓取"},
			{TaskType: taskSpiderFetchSiteBoth, Label: "JavBus + Sukebei 同时抓取"},
		},
		HeaderLinks: []crawlerJobsPageLink{
			{Href: "/fetch-site-javbus-list", Label: "JavBus 列表"},
			{Href: "/fetch-site-sukebei-list", Label: "Sukebei 列表"},
		},
		Labels: crawlerJobsPageLabels{
			Extra:   "任务类型",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		MovieCardFilter:  filterView,
		ShowMovieFilters: true,
	})
}

func (h *CrawlerPages) FetchSiteSukebeiFilteredPage(c *gin.Context) {
	h.renderJobsPage(c, crawlerJobsPageConfig{
		Title:             "Sukebei 筛选抓取任务",
		PageTitle:         "Sukebei 列表筛选抓取",
		TaskPanelTitle:    "筛选结果抓取任务",
		TaskHelperText:    "本页用于展示从 Sukebei 列表页按当前筛选结果触发的批量抓取任务。请在列表页点击“触发当前筛选结果”后回到这里查看队列、进度和日志流。",
		PageNote:          "队列会按列表筛选快照固定下来，并实时展示待入列、处理中、已完成的影片。",
		TaskTableTitle:    "Sukebei 筛选抓取任务",
		EventTitle:        "Sukebei 筛选抓取事件流",
		StorageKey:        "crawler_jobs_fetch_site_sukebei_filtered_selected_job",
		DefaultTaskType:   loop.TaskSpiderFetchSukebeiFilter,
		OverviewExtraMode: "sukebei_filtered_queue",
		EmptyStateText:    "等待 Sukebei 列表页触发筛选抓取任务",
		AllowedTaskTypes: []string{
			loop.TaskSpiderFetchSukebeiFilter,
		},
		HeaderLinks: []crawlerJobsPageLink{
			{Href: "/fetch-site-sukebei-list", Label: "返回 Sukebei 列表"},
			{Href: "/fetch-site-javbus-list", Label: "JavBus 列表"},
		},
		Labels: crawlerJobsPageLabels{
			Extra:   "队列",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		HideTaskForm: true,
	})
}

func (h *CrawlerPages) FetchSiteJavbusFilteredPage(c *gin.Context) {
	h.renderJobsPage(c, crawlerJobsPageConfig{
		Title:             "JavBus 筛选抓取任务",
		PageTitle:         "JavBus 列表筛选抓取",
		TaskPanelTitle:    "筛选结果抓取任务",
		TaskHelperText:    "本页用于展示从 JavBus 列表页按当前筛选结果触发的批量抓取任务。请在列表页点击“触发筛选结果抓取”后回到这里查看队列、进度和日志流。",
		PageNote:          "队列会按列表筛选快照固定下来，并实时展示待入列和已完成的影片。",
		TaskTableTitle:    "JavBus 筛选抓取任务",
		EventTitle:        "JavBus 筛选抓取事件流",
		StorageKey:        "crawler_jobs_fetch_site_javbus_filtered_selected_job",
		DefaultTaskType:   loop.TaskSpiderFetchJavbusFilter,
		OverviewExtraMode: "javbus_filtered_queue",
		EmptyStateText:    "等待 JavBus 列表页触发筛选抓取任务",
		AllowedTaskTypes: []string{
			loop.TaskSpiderFetchJavbusFilter,
		},
		HeaderLinks: []crawlerJobsPageLink{
			{Href: "/fetch-site-javbus-list", Label: "返回 JavBus 列表"},
			{Href: "/fetch-site-sukebei-list", Label: "Sukebei 列表"},
		},
		Labels: crawlerJobsPageLabels{
			Extra:   "队列",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		HideTaskForm: true,
	})
}

func (h *CrawlerPages) BackfillPage(c *gin.Context) {
	h.renderJobsPage(c, crawlerJobsPageConfig{
		Title:             "回填任务",
		PageTitle:         "人物回填 / 演员回填 / 周期排行回填 / SC 统计回填 / SC 影片移动",
		TaskPanelTitle:    "回填任务",
		PageNote:          "这里只处理 person 回填、演员 Rank 回填、周期排行回填、SC 统计回填和 SC 影片移动。",
		TaskTableTitle:    "回填任务",
		EventTitle:        "回填事件流",
		StorageKey:        "crawler_jobs_backfill_selected_job",
		DefaultTaskType:   loop.TaskSpiderBackfillRankPeriod,
		OverviewExtraMode: "task_type",
		EmptyStateText:    "等待回填任务触发",
		AllowedTaskTypes: []string{
			loop.TaskSpiderBackfillPerson,
			loop.TaskSpiderBackfillRankPeriod,
			loop.TaskSpiderBackfillFetchSite,
			loop.TaskSpiderRebuildCastRank,
			loop.TaskSpiderRebuildActorRank,
			loop.TaskScRebuildStats,
			loop.TaskScMove,
		},
		TaskButtons: []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderBackfillPerson, Label: "person 回填"},
			{TaskType: loop.TaskSpiderBackfillRankPeriod, Label: "周期排行回填"},
			{TaskType: loop.TaskSpiderBackfillFetchSite, Label: "外站任务回填"},
			{TaskType: loop.TaskSpiderRebuildCastRank, Label: "演员 Rank 回填"},
			{TaskType: loop.TaskSpiderRebuildActorRank, Label: "单演员 Rank"},
			{TaskType: loop.TaskScRebuildStats, Label: "SC 统计回填"},
		},
		Labels: crawlerJobsPageLabels{
			Extra:   "任务类型",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
	})
}

func (h *CrawlerPages) TasksPage(c *gin.Context) {
	c.HTML(http.StatusOK, "page.crawler_tasks", gin.H{
		"Title":           "任务列表",
		"QuickNavCurrent": "crawler_tasks",
		"JobID":           strings.TrimSpace(c.Query("job_id")),
		"TriggerPageURL":  "/triggers/dailybest",
	})
}

func (h *CrawlerPages) DetailLoopPage(c *gin.Context) {
	c.HTML(http.StatusOK, "page.crawler_detail_loop", gin.H{
		"Title": "详情抓取循环",
	})
}

func (h *CrawlerPages) renderJobsPage(c *gin.Context, cfg crawlerJobsPageConfig) {
	c.HTML(http.StatusOK, "page.admin_triggers", gin.H{
		"Title":                     cfg.Title,
		"PageTitle":                 cfg.PageTitle,
		"QuickNavCurrent":           cfg.QuickNavCurrent,
		"TaskPanelTitle":            cfg.TaskPanelTitle,
		"TaskHelperText":            cfg.TaskHelperText,
		"PageNote":                  cfg.PageNote,
		"TaskTableTitle":            cfg.TaskTableTitle,
		"EventTitle":                cfg.EventTitle,
		"StorageKey":                cfg.StorageKey,
		"DefaultTaskType":           cfg.DefaultTaskType,
		"AllowedTaskTypes":          cfg.AllowedTaskTypes,
		"OverviewExtraMode":         cfg.OverviewExtraMode,
		"EmptyStateText":            cfg.EmptyStateText,
		"TaskButtons":               cfg.TaskButtons,
		"HeaderLinks":               cfg.HeaderLinks,
		"Labels":                    cfg.Labels,
		"HideTaskForm":              cfg.HideTaskForm,
		"MovieCardFilter":           cfg.MovieCardFilter,
		"ShowMovieFilters":          cfg.ShowMovieFilters,
		"ShowFavoriteAlbum":         cfg.ShowFavoriteAlbum,
		"FavoriteRows":              cfg.FavoriteRows,
		"FetchSehuatangList":        cfg.FetchSehuatangList,
		"FetchSehuatangKey":         cfg.FetchSehuatangKey,
		"FetchSehuatangStartPage":   cfg.FetchSehuatangStartPage,
		"FetchSehuatangEndPage":     cfg.FetchSehuatangEndPage,
		"FetchSehuatangPersistMode": cfg.FetchSehuatangPersistMode,
		"ScRootDir":                 h.scRootDir,
		"JobID":                     strings.TrimSpace(c.Query("job_id")),
	})
}

func (h *CrawlerPages) loadFavoriteAlbumRows(ctx context.Context) ([]*favoriteAlbumItemRow, error) {
	album, err := h.deps.AlbumModel.FindOneByName(ctx, crawlerFavoriteAlbumName)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return []*favoriteAlbumItemRow{}, nil
		}
		return nil, err
	}

	rows, err := h.deps.AlbumItemModel.ListPageRows(ctx, album.Id, 0, crawlerFavoriteCardMaxRow, "`created_on` DESC, `id` DESC", moviex.AlbumItemPageFilter{})
	if err != nil {
		return nil, err
	}

	out := make([]*favoriteAlbumItemRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		movieName := strings.TrimSpace(row.MovieName)
		out = append(out, &favoriteAlbumItemRow{
			MovieName:   movieName,
			MovieJavID:  strings.TrimSpace(row.MovieJavId),
			SourceType:  strings.TrimSpace(row.SourceType),
			SourceLabel: favoriteSourceTypeLabel(row.SourceType),
			Size:        strings.TrimSpace(row.Size),
		})
	}
	return out, nil
}

func favoriteSourceTypeLabel(sourceType string) string {
	switch strings.TrimSpace(sourceType) {
	case "javbus_magnet":
		return "JavBus"
	case "sukebei_torrent":
		return "Sukebei"
	case "sehuatang_magnet":
		return "Sehuatang"
	default:
		return "未知"
	}
}
