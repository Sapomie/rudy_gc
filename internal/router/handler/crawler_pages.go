package handler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/service/loop"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"
)

type CrawlerPages struct {
	runtime   *loop.FetchLoopService
	fetchSite *fetchsite.Service
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
	Title             string
	PageTitle         string
	TaskPanelTitle    string
	PageNote          string
	TaskTableTitle    string
	EventTitle        string
	StorageKey        string
	DefaultTaskType   string
	OverviewExtraMode string
	EmptyStateText    string
	AllowedTaskTypes  []string
	TaskButtons       []crawlerJobsPageTask
	HeaderLinks       []crawlerJobsPageLink
	Labels            crawlerJobsPageLabels
	MovieCardFilter   *movieCardFilterView
	ShowMovieFilters  bool
}

const taskSpiderFetchSiteBoth = "spider_fetch_site_both_resources"

func NewCrawlerPages(deps *svc.Deps) *CrawlerPages {
	return &CrawlerPages{
		runtime:   newCrawlerRuntime(deps),
		fetchSite: newFetchSitePageService(deps),
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
		"Title":          "任务列表",
		"JobID":          strings.TrimSpace(c.Query("job_id")),
		"TriggerPageURL": "/triggers/dailybest",
	})
}

func (h *CrawlerPages) DetailLoopPage(c *gin.Context) {
	c.HTML(http.StatusOK, "page.crawler_detail_loop", gin.H{
		"Title": "详情抓取循环",
	})
}

func (h *CrawlerPages) renderJobsPage(c *gin.Context, cfg crawlerJobsPageConfig) {
	c.HTML(http.StatusOK, "page.admin_triggers", gin.H{
		"Title":             cfg.Title,
		"PageTitle":         cfg.PageTitle,
		"TaskPanelTitle":    cfg.TaskPanelTitle,
		"PageNote":          cfg.PageNote,
		"TaskTableTitle":    cfg.TaskTableTitle,
		"EventTitle":        cfg.EventTitle,
		"StorageKey":        cfg.StorageKey,
		"DefaultTaskType":   cfg.DefaultTaskType,
		"AllowedTaskTypes":  cfg.AllowedTaskTypes,
		"OverviewExtraMode": cfg.OverviewExtraMode,
		"EmptyStateText":    cfg.EmptyStateText,
		"TaskButtons":       cfg.TaskButtons,
		"HeaderLinks":       cfg.HeaderLinks,
		"Labels":            cfg.Labels,
		"MovieCardFilter":   cfg.MovieCardFilter,
		"ShowMovieFilters":  cfg.ShowMovieFilters,
		"ScRootDir":         h.scRootDir,
		"JobID":             strings.TrimSpace(c.Query("job_id")),
	})
}
