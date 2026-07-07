package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/fetchsehuatang"
	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/service/loop"
	"rudy_gc/internal/service/sehuatang"
	"rudy_gc/internal/service/wkv"
	"rudy_gc/internal/types"
)

type CrawlerPagesAPI struct {
	deps      *dep.Dep
	fetchSite *fetchsite.Service
	sehuatang *sehuatang.Service
	wkvSvc    *wkv.Service
	scRootDir string
}

type crawlerJobsPageTask struct {
	TaskType string `json:"TaskType"`
	Label    string `json:"Label"`
}

type crawlerJobsPageLabels struct {
	Extra   string `json:"Extra"`
	Result  string `json:"Result"`
	Elapsed string `json:"Elapsed"`
}

type crawlerJobsPageLink struct {
	Href  string `json:"Href"`
	Label string `json:"Label"`
}

func NewCrawlerPagesAPI(deps *dep.Dep) *CrawlerPagesAPI {
	return &CrawlerPagesAPI{
		deps:      deps,
		fetchSite: fetchsite.NewService(deps),
		sehuatang: sehuatang.NewService(deps),
		wkvSvc:    wkv.NewService(deps),
		scRootDir: deps.Config.Film.ScRootDir,
	}
}

func (h *CrawlerPagesAPI) DailyBestPage(c *gin.Context) {
	defer withFrontendPath(c, "/triggers/dailybest")()

	c.JSON(http.StatusOK, gin.H{
		"Title":             "DailyBest 任务",
		"ViewType":          "crawler-jobs",
		"PageTitle":         "DailyBest 抓取",
		"QuickNavCurrent":   "dailybest",
		"TaskPanelTitle":    "DailyBest 触发",
		"PageNote":          "只处理 DailyBest 抓取与同步任务。",
		"TaskTableTitle":    "DailyBest 任务",
		"EventTitle":        "DailyBest 事件流",
		"StorageKey":        "crawler_jobs_dailybest_selected_job",
		"DefaultTaskType":   loop.TaskSpiderDailyBest,
		"OverviewExtraMode": "dailybest_stages",
		"EmptyStateText":    "等待 DailyBest 任务触发",
		"AllowedTaskTypes": []string{
			loop.TaskSpiderDailyBest,
			loop.TaskSpiderDailyBestSync,
		},
		"TaskButtons": []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderDailyBest, Label: "DailyBest 抓取"},
			{TaskType: loop.TaskSpiderDailyBestSync, Label: "DailyBest 同步"},
		},
		"HeaderLinks": []crawlerJobsPageLink{},
		"Labels": crawlerJobsPageLabels{
			Extra:   "当前页数",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		"HideTaskForm":              false,
		"ShowMovieFilters":          false,
		"ShowFavoriteAlbum":         false,
		"FetchSehuatangList":        "",
		"FetchSehuatangKey":         "",
		"FetchSehuatangStartPage":   int64(0),
		"FetchSehuatangEndPage":     int64(0),
		"FetchSehuatangPersistMode": "",
		"ScRootDir":                 h.scRootDir,
		"JobID":                     strings.TrimSpace(c.Query("job_id")),
	})
}

func (h *CrawlerPagesAPI) SeedsPage(c *gin.Context) {
	defer withFrontendPath(c, "/triggers/seeds")()

	c.JSON(http.StatusOK, gin.H{
		"Title":             "Seeds 任务",
		"ViewType":          "crawler-jobs",
		"PageTitle":         "活跃 Seeds / 按名称抓取 / 刷新最久详情",
		"TaskPanelTitle":    "Seeds 抓取",
		"PageNote":          "这里只处理活跃 Seeds、按名称抓取和刷新最久详情。",
		"TaskTableTitle":    "Seeds 任务",
		"EventTitle":        "Seeds 事件流",
		"StorageKey":        "crawler_jobs_seeds_selected_job",
		"DefaultTaskType":   loop.TaskSpiderSeeds,
		"OverviewExtraMode": "seeds_stages",
		"EmptyStateText":    "等待 Seeds 任务触发",
		"AllowedTaskTypes": []string{
			loop.TaskSpiderSeeds,
			loop.TaskSpiderSeedByName,
			loop.TaskSpiderRefreshOldest,
		},
		"TaskButtons": []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderSeeds, Label: "活跃 Seeds"},
			{TaskType: loop.TaskSpiderSeedByName, Label: "按名称抓取"},
			{TaskType: loop.TaskSpiderRefreshOldest, Label: "刷新最久详情"},
		},
		"HeaderLinks": []crawlerJobsPageLink{},
		"Labels": crawlerJobsPageLabels{
			Extra:   "详情 Loop",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		"HideTaskForm":              false,
		"ShowMovieFilters":          false,
		"ShowFavoriteAlbum":         false,
		"FetchSehuatangList":        "",
		"FetchSehuatangKey":         "",
		"FetchSehuatangStartPage":   int64(0),
		"FetchSehuatangEndPage":     int64(0),
		"FetchSehuatangPersistMode": "",
		"ScRootDir":                 h.scRootDir,
		"JobID":                     strings.TrimSpace(c.Query("job_id")),
	})
}

func (h *CrawlerPagesAPI) PostProcessPage(c *gin.Context) {
	defer withFrontendPath(c, "/triggers/post-process")()

	c.JSON(http.StatusOK, gin.H{
		"Title":             "后处理任务",
		"ViewType":          "crawler-jobs",
		"PageTitle":         "图片抓取 / 标题翻译",
		"TaskPanelTitle":    "后处理任务",
		"PageNote":          "单独触发封面抓取、标题翻译，或两者一起执行。这里不重跑榜单、种子和详情抓取。",
		"TaskTableTitle":    "后处理任务",
		"EventTitle":        "后处理事件流",
		"StorageKey":        "crawler_jobs_post_process_selected_job",
		"DefaultTaskType":   loop.TaskSpiderPostProcess,
		"OverviewExtraMode": "post_process_stages",
		"EmptyStateText":    "等待后处理任务触发",
		"AllowedTaskTypes": []string{
			loop.TaskSpiderDownloadCover,
			loop.TaskSpiderTranslateTitle,
			loop.TaskSpiderPostProcess,
		},
		"TaskButtons": []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderDownloadCover, Label: "图片抓取"},
			{TaskType: loop.TaskSpiderTranslateTitle, Label: "标题翻译"},
			{TaskType: loop.TaskSpiderPostProcess, Label: "图片 + 标题翻译"},
		},
		"HeaderLinks": []crawlerJobsPageLink{},
		"Labels": crawlerJobsPageLabels{
			Extra:   "当前阶段",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		"HideTaskForm":              false,
		"ShowMovieFilters":          false,
		"ShowFavoriteAlbum":         false,
		"FetchSehuatangList":        "",
		"FetchSehuatangKey":         "",
		"FetchSehuatangStartPage":   int64(0),
		"FetchSehuatangEndPage":     int64(0),
		"FetchSehuatangPersistMode": "",
		"ScRootDir":                 h.scRootDir,
		"JobID":                     strings.TrimSpace(c.Query("job_id")),
	})
}

func (h *CrawlerPagesAPI) BackfillPage(c *gin.Context) {
	defer withFrontendPath(c, "/triggers/backfill")()

	c.JSON(http.StatusOK, gin.H{
		"Title":             "回填任务",
		"ViewType":          "crawler-jobs",
		"PageTitle":         "人物回填 / 演员回填 / 周期排行回填 / SC 统计回填",
		"TaskPanelTitle":    "回填任务",
		"PageNote":          "这里只处理 person 回填、演员 Rank 回填、周期排行回填和 SC 统计回填。",
		"TaskTableTitle":    "回填任务",
		"EventTitle":        "回填事件流",
		"StorageKey":        "crawler_jobs_backfill_selected_job",
		"DefaultTaskType":   loop.TaskSpiderBackfillRankPeriod,
		"OverviewExtraMode": "task_type",
		"EmptyStateText":    "等待回填任务触发",
		"AllowedTaskTypes": []string{
			loop.TaskSpiderBackfillPerson,
			loop.TaskSpiderBackfillRankPeriod,
			loop.TaskSpiderBackfillFetchSite,
			loop.TaskSpiderRebuildCastRank,
			loop.TaskSpiderRebuildActorRank,
			loop.TaskScRebuildStats,
		},
		"TaskButtons": []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderBackfillPerson, Label: "person 回填"},
			{TaskType: loop.TaskSpiderBackfillRankPeriod, Label: "周期排行回填"},
			{TaskType: loop.TaskSpiderBackfillFetchSite, Label: "外站任务回填"},
			{TaskType: loop.TaskSpiderRebuildCastRank, Label: "演员 Rank 回填"},
			{TaskType: loop.TaskSpiderRebuildActorRank, Label: "单演员 Rank"},
			{TaskType: loop.TaskScRebuildStats, Label: "SC 统计回填"},
		},
		"HeaderLinks": []crawlerJobsPageLink{},
		"Labels": crawlerJobsPageLabels{
			Extra:   "任务类型",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		"HideTaskForm":              false,
		"ShowMovieFilters":          false,
		"ShowFavoriteAlbum":         false,
		"FetchSehuatangList":        "",
		"FetchSehuatangKey":         "",
		"FetchSehuatangStartPage":   int64(0),
		"FetchSehuatangEndPage":     int64(0),
		"FetchSehuatangPersistMode": "",
		"ScRootDir":                 h.scRootDir,
		"JobID":                     strings.TrimSpace(c.Query("job_id")),
	})
}

func (h *CrawlerPagesAPI) TasksPage(c *gin.Context) {
	defer withFrontendPath(c, "/crawler/tasks")()

	c.JSON(http.StatusOK, gin.H{
		"Title":           "任务列表",
		"ViewType":        "crawler-tasks",
		"QuickNavCurrent": "crawler_tasks",
		"JobID":           strings.TrimSpace(c.Query("job_id")),
		"TriggerPageURL":  "/triggers/dailybest",
	})
}

func (h *CrawlerPagesAPI) DetailLoopPage(c *gin.Context) {
	defer withFrontendPath(c, "/crawler/detail-loop")()

	c.JSON(http.StatusOK, gin.H{
		"Title":    "详情抓取循环",
		"ViewType": "crawler-detail-loop",
		"ApiBase":  "/api/crawler/detail-loop",
	})
}

func (h *CrawlerPagesAPI) FetchSitePage(c *gin.Context) {
	defer withFrontendPath(c, "/triggers/fetch-site")()

	req, err := parseMovieCardRequest(c, types.ListMovieFullRequest{
		OrderBy: consts.OrderByReleasingDate,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数解析错误: " + err.Error()})
		return
	}
	filterView := buildMovieCardFilterView(c, req, req.OrderBy, nil)
	filterView.HideDirs = true
	filterView.TextDateInputs = true

	c.JSON(http.StatusOK, gin.H{
		"Title":             "抓取站点任务",
		"ViewType":          "crawler-jobs",
		"PageTitle":         "JavBus / Sukebei 资源抓取",
		"QuickNavCurrent":   "fetch-site",
		"TaskPanelTitle":    "抓取站点任务",
		"PageNote":          "JavBus / Sukebei 抓取目标由 cards 同源筛选决定，数量参数控制本次最多处理条数。",
		"TaskTableTitle":    "抓取站点任务",
		"EventTitle":        "抓取站点事件流",
		"StorageKey":        "crawler_jobs_fetch_site_selected_job",
		"DefaultTaskType":   loop.TaskSpiderFetchJavbus,
		"OverviewExtraMode": "fetch_site_summary",
		"EmptyStateText":    "等待抓取站点任务触发",
		"AllowedTaskTypes": []string{
			loop.TaskSpiderFetchJavbus,
			loop.TaskSpiderFetchSukebei,
		},
		"TaskButtons": []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderFetchJavbus, Label: "JavBus 资源抓取"},
			{TaskType: loop.TaskSpiderFetchSukebei, Label: "Sukebei 资源抓取"},
			{TaskType: loop.TaskSpiderFetchSiteBoth, Label: "JavBus + Sukebei 同时抓取"},
		},
		"HeaderLinks": []crawlerJobsPageLink{
			{Href: "/fetch-site-javbus-list", Label: "JavBus 列表"},
			{Href: "/fetch-site-sukebei-list", Label: "Sukebei 列表"},
		},
		"Labels": crawlerJobsPageLabels{
			Extra:   "任务类型",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		"HideTaskForm":              false,
		"MovieCardFilter":           filterView,
		"ShowMovieFilters":          true,
		"ShowFavoriteAlbum":         false,
		"FetchSehuatangList":        "",
		"FetchSehuatangKey":         "",
		"FetchSehuatangStartPage":   int64(0),
		"FetchSehuatangEndPage":     int64(0),
		"FetchSehuatangPersistMode": "",
		"ScRootDir":                 h.scRootDir,
		"JobID":                     strings.TrimSpace(c.Query("job_id")),
	})
}

func (h *CrawlerPagesAPI) FetchSehuatangPage(c *gin.Context) {
	defer withFrontendPath(c, "/triggers/fetch-sehuatang")()

	c.JSON(http.StatusOK, gin.H{
		"Title":             "色花堂抓取任务",
		"ViewType":          "crawler-jobs",
		"PageTitle":         "色花堂磁力抓取（日志流）",
		"QuickNavCurrent":   "fetch_sehuatang",
		"TaskPanelTitle":    "色花堂抓取任务",
		"PageNote":          "抓取列表普通帖子并直接入库，自动跳过置顶/公告帖；支持按起始页到结束页范围抓取。",
		"TaskTableTitle":    "色花堂抓取任务",
		"EventTitle":        "色花堂抓取事件流",
		"StorageKey":        "crawler_jobs_fetch_sehuatang_selected_job",
		"DefaultTaskType":   loop.TaskSpiderFetchSehuatang,
		"OverviewExtraMode": "sehuatang_progress",
		"EmptyStateText":    "等待 色花堂抓取任务触发",
		"AllowedTaskTypes": []string{
			loop.TaskSpiderFetchSehuatang,
		},
		"TaskButtons": []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderFetchSehuatang, Label: "色花堂磁力抓取"},
		},
		"HeaderLinks": []crawlerJobsPageLink{
			{Href: "/triggers/fetch-site", Label: "FetchSite"},
			{Href: "/fetch-sites", Label: "FetchSites 配置"},
			{Href: "/fetch-site-sehuatang-list", Label: "Sehuatang 列表"},
		},
		"Labels": crawlerJobsPageLabels{
			Extra:   "任务类型",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		"HideTaskForm":                 false,
		"ShowMovieFilters":             false,
		"ShowFavoriteAlbum":            false,
		"FetchSehuatangList":           sehuatangListURLFromFetchSiteConfig(h.deps),
		"FetchSehuatangKey":            fetchsehuatang.DefaultKeyword,
		"FetchSehuatangStartPage":      fetchsehuatang.DefaultStartPage,
		"FetchSehuatangEndPage":        fetchsehuatang.DefaultEndPage,
		"FetchSehuatangPersistMode":    fetchsehuatang.DefaultPersistMode,
		"FetchSehuatangStopOnExisting": false,
		"ScRootDir":                    h.scRootDir,
		"JobID":                        strings.TrimSpace(c.Query("job_id")),
	})
}

func (h *CrawlerPagesAPI) FetchSehuatangSHPage(c *gin.Context) {
	defer withFrontendPath(c, "/triggers/fetch-sehuatang-sh")()

	c.JSON(http.StatusOK, gin.H{
		"Title":             "色花堂帖子新流程",
		"ViewType":          "crawler-jobs",
		"PageTitle":         "色花堂帖子新流程（sh 表）",
		"QuickNavCurrent":   "fetch_sehuatang_sh",
		"TaskPanelTitle":    "色花堂帖子新流程",
		"PageNote":          "抓取搜索结果或指定版块普通帖，详情写入 sh_sehuatang_thread，并记录 sh_sehuatang_fetch_run / sh_sehuatang_fetch_item。",
		"TaskTableTitle":    "色花堂帖子新流程任务",
		"EventTitle":        "色花堂帖子新流程事件流",
		"StorageKey":        "crawler_jobs_fetch_sehuatang_sh_selected_job",
		"DefaultTaskType":   loop.TaskSpiderFetchSehuatangSH,
		"OverviewExtraMode": "sehuatang_progress",
		"EmptyStateText":    "等待 色花堂帖子新流程任务触发",
		"AllowedTaskTypes": []string{
			loop.TaskSpiderFetchSehuatangSH,
		},
		"TaskButtons": []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderFetchSehuatangSH, Label: "色花堂帖子新流程"},
		},
		"HeaderLinks": []crawlerJobsPageLink{
			{Href: "/triggers/fetch-sehuatang", Label: "旧磁力抓取"},
			{Href: "/fetch-sites", Label: "FetchSites 配置"},
			{Href: "/fetch-site-sehuatang-list", Label: "Sehuatang 列表"},
		},
		"Labels": crawlerJobsPageLabels{
			Extra:   "任务类型",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		"HideTaskForm":                 false,
		"ShowMovieFilters":             false,
		"ShowFavoriteAlbum":            false,
		"ShowSehuatangSH":              true,
		"FetchSehuatangSHMode":         fetchsehuatang.SHFetchModeSearch,
		"FetchSehuatangSHTargetForums": []string{fetchsehuatang.SHForumAsiaUncensoredOriginal, fetchsehuatang.SHForumHDChineseSubtitle},
		"FetchSehuatangList":           "",
		"FetchSehuatangKey":            "fc2ppv",
		"FetchSehuatangStartPage":      fetchsehuatang.DefaultStartPage,
		"FetchSehuatangEndPage":        fetchsehuatang.DefaultEndPage,
		"FetchSehuatangPersistMode":    "",
		"FetchSehuatangStopOnExisting": false,
		"ScRootDir":                    h.scRootDir,
		"JobID":                        strings.TrimSpace(c.Query("job_id")),
	})
}

func (h *CrawlerPagesAPI) FetchSehuatangSHThreadListPage(c *gin.Context) {
	defer withFrontendPath(c, "/triggers/fetch-sehuatang-sh-thread-list")()

	c.JSON(http.StatusOK, gin.H{
		"Title":             "色花堂帖子列表抓取",
		"ViewType":          "crawler-jobs",
		"PageTitle":         "色花堂帖子列表抓取",
		"QuickNavCurrent":   "fetch_sehuatang_sh_thread_list",
		"TaskPanelTitle":    "色花堂帖子列表抓取",
		"PageNote":          "只翻 forum-36 / forum-103 列表页，只保存帖子基础信息，不请求帖子详情页。",
		"TaskTableTitle":    "色花堂帖子列表抓取任务",
		"EventTitle":        "色花堂帖子列表抓取事件流",
		"StorageKey":        "crawler_jobs_fetch_sehuatang_sh_thread_list_selected_job",
		"DefaultTaskType":   loop.TaskSpiderFetchSehuatangSHList,
		"OverviewExtraMode": "sehuatang_progress",
		"EmptyStateText":    "等待 色花堂帖子列表抓取任务触发",
		"AllowedTaskTypes": []string{
			loop.TaskSpiderFetchSehuatangSHList,
		},
		"TaskButtons": []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderFetchSehuatangSHList, Label: "色花堂帖子列表抓取"},
		},
		"HeaderLinks": []crawlerJobsPageLink{
			{Href: "/triggers/fetch-sehuatang-sh-thread-detail", Label: "详情补全"},
			{Href: "/fetch-site-sehuatang-sh-list", Label: "帖子池"},
		},
		"Labels": crawlerJobsPageLabels{
			Extra:   "任务类型",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		"HideTaskForm":                 false,
		"ShowMovieFilters":             false,
		"ShowFavoriteAlbum":            false,
		"ShowSehuatangSHThreadList":    true,
		"FetchSehuatangSHTargetForums": []string{fetchsehuatang.SHForumAsiaUncensoredOriginal, fetchsehuatang.SHForumHDChineseSubtitle},
		"FetchSehuatangStartPage":      fetchsehuatang.DefaultStartPage,
		"FetchSehuatangEndPage":        fetchsehuatang.DefaultEndPage,
		"ScRootDir":                    h.scRootDir,
		"JobID":                        strings.TrimSpace(c.Query("job_id")),
	})
}

func (h *CrawlerPagesAPI) FetchSehuatangSHThreadDetailPage(c *gin.Context) {
	defer withFrontendPath(c, "/triggers/fetch-sehuatang-sh-thread-detail")()

	c.JSON(http.StatusOK, gin.H{
		"Title":             "色花堂帖子详情补全",
		"ViewType":          "crawler-jobs",
		"PageTitle":         "色花堂帖子详情补全",
		"QuickNavCurrent":   "fetch_sehuatang_sh_thread_detail",
		"TaskPanelTitle":    "色花堂帖子详情补全",
		"PageNote":          "先从 sh_sehuatang_thread 筛选并预览帖子，再抓取页面上选中的帖子详情。",
		"TaskTableTitle":    "色花堂帖子详情补全任务",
		"EventTitle":        "色花堂帖子详情补全事件流",
		"StorageKey":        "crawler_jobs_fetch_sehuatang_sh_thread_detail_selected_job",
		"DefaultTaskType":   loop.TaskSpiderFetchSehuatangSHDetail,
		"OverviewExtraMode": "sehuatang_progress",
		"EmptyStateText":    "等待 色花堂帖子详情补全任务触发",
		"AllowedTaskTypes": []string{
			loop.TaskSpiderFetchSehuatangSHDetail,
		},
		"TaskButtons": []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderFetchSehuatangSHDetail, Label: "抓取选中详情"},
		},
		"HeaderLinks": []crawlerJobsPageLink{
			{Href: "/triggers/fetch-sehuatang-sh-thread-list", Label: "列表抓取"},
			{Href: "/fetch-site-sehuatang-sh-list", Label: "帖子池"},
		},
		"Labels": crawlerJobsPageLabels{
			Extra:   "任务类型",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		"HideTaskForm":                 false,
		"ShowMovieFilters":             false,
		"ShowFavoriteAlbum":            false,
		"ShowSehuatangSHThreadDetail":  true,
		"FetchSehuatangSHTargetForums": []string{fetchsehuatang.SHForumAsiaUncensoredOriginal, fetchsehuatang.SHForumHDChineseSubtitle},
		"FetchSehuatangSHDetailStatus": fetchsehuatang.SHDetailStatusPending,
		"ScRootDir":                    h.scRootDir,
		"JobID":                        strings.TrimSpace(c.Query("job_id")),
	})
}

func (h *CrawlerPagesAPI) FetchSiteSukebeiFilteredPage(c *gin.Context) {
	defer withFrontendPath(c, "/triggers/fetch-site-sukebei-filtered")()

	c.JSON(http.StatusOK, gin.H{
		"Title":             "Sukebei 筛选抓取任务",
		"ViewType":          "crawler-jobs",
		"PageTitle":         "Sukebei 列表筛选抓取",
		"TaskPanelTitle":    "筛选结果抓取任务",
		"TaskHelperText":    "本页用于展示从 Sukebei 列表页按当前筛选结果触发的批量抓取任务。请在列表页点击“触发当前筛选结果”后回到这里查看队列、进度和日志流。",
		"PageNote":          "队列会按列表筛选快照固定下来，并实时展示待入列、处理中、已完成的影片。",
		"TaskTableTitle":    "Sukebei 筛选抓取任务",
		"EventTitle":        "Sukebei 筛选抓取事件流",
		"StorageKey":        "crawler_jobs_fetch_site_sukebei_filtered_selected_job",
		"DefaultTaskType":   loop.TaskSpiderFetchSukebeiFilter,
		"OverviewExtraMode": "sukebei_filtered_queue",
		"EmptyStateText":    "等待 Sukebei 列表页触发筛选抓取任务",
		"AllowedTaskTypes": []string{
			loop.TaskSpiderFetchSukebeiFilter,
		},
		"TaskButtons": []crawlerJobsPageTask{},
		"HeaderLinks": []crawlerJobsPageLink{
			{Href: "/fetch-site-sukebei-list", Label: "返回 Sukebei 列表"},
			{Href: "/fetch-site-javbus-list", Label: "JavBus 列表"},
		},
		"Labels": crawlerJobsPageLabels{
			Extra:   "队列",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		"HideTaskForm":              true,
		"ShowMovieFilters":          false,
		"ShowFavoriteAlbum":         false,
		"FetchSehuatangList":        "",
		"FetchSehuatangKey":         "",
		"FetchSehuatangStartPage":   int64(0),
		"FetchSehuatangEndPage":     int64(0),
		"FetchSehuatangPersistMode": "",
		"ScRootDir":                 h.scRootDir,
		"JobID":                     strings.TrimSpace(c.Query("job_id")),
	})
}

func (h *CrawlerPagesAPI) FetchSiteJavbusFilteredPage(c *gin.Context) {
	defer withFrontendPath(c, "/triggers/fetch-site-javbus-filtered")()

	c.JSON(http.StatusOK, gin.H{
		"Title":             "JavBus 筛选抓取任务",
		"ViewType":          "crawler-jobs",
		"PageTitle":         "JavBus 列表筛选抓取",
		"TaskPanelTitle":    "筛选结果抓取任务",
		"TaskHelperText":    "本页用于展示从 JavBus 列表页按当前筛选结果触发的批量抓取任务。请在列表页点击“触发筛选结果抓取”后回到这里查看队列、进度和日志流。",
		"PageNote":          "队列会按列表筛选快照固定下来，并实时展示待入列和已完成的影片。",
		"TaskTableTitle":    "JavBus 筛选抓取任务",
		"EventTitle":        "JavBus 筛选抓取事件流",
		"StorageKey":        "crawler_jobs_fetch_site_javbus_filtered_selected_job",
		"DefaultTaskType":   loop.TaskSpiderFetchJavbusFilter,
		"OverviewExtraMode": "javbus_filtered_queue",
		"EmptyStateText":    "等待 JavBus 列表页触发筛选抓取任务",
		"AllowedTaskTypes": []string{
			loop.TaskSpiderFetchJavbusFilter,
		},
		"TaskButtons": []crawlerJobsPageTask{},
		"HeaderLinks": []crawlerJobsPageLink{
			{Href: "/fetch-site-javbus-list", Label: "返回 JavBus 列表"},
			{Href: "/fetch-site-sukebei-list", Label: "Sukebei 列表"},
		},
		"Labels": crawlerJobsPageLabels{
			Extra:   "队列",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		"HideTaskForm":              true,
		"ShowMovieFilters":          false,
		"ShowFavoriteAlbum":         false,
		"FetchSehuatangList":        "",
		"FetchSehuatangKey":         "",
		"FetchSehuatangStartPage":   int64(0),
		"FetchSehuatangEndPage":     int64(0),
		"FetchSehuatangPersistMode": "",
		"ScRootDir":                 h.scRootDir,
		"JobID":                     strings.TrimSpace(c.Query("job_id")),
	})
}
