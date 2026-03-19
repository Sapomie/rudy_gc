package handler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/loop"
	"rudy_gc/internal/svc"
)

type CrawlerPages struct {
	runtime   *loop.FetchLoopService
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
	Labels            crawlerJobsPageLabels
}

func NewCrawlerPages(deps *svc.Deps) *CrawlerPages {
	return &CrawlerPages{
		runtime:   newCrawlerRuntime(deps),
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
		OverviewExtraMode: "task_type",
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

func (h *CrawlerPages) BackfillPage(c *gin.Context) {
	h.renderJobsPage(c, crawlerJobsPageConfig{
		Title:             "回填任务",
		PageTitle:         "演员回填 / 周期排行回填 / SC 统计回填",
		TaskPanelTitle:    "回填任务",
		PageNote:          "这里只处理演员 Rank 回填、周期排行回填和 SC 统计回填。",
		TaskTableTitle:    "回填任务",
		EventTitle:        "回填事件流",
		StorageKey:        "crawler_jobs_backfill_selected_job",
		DefaultTaskType:   loop.TaskSpiderBackfillRankPeriod,
		OverviewExtraMode: "task_type",
		EmptyStateText:    "等待回填任务触发",
		AllowedTaskTypes: []string{
			loop.TaskSpiderBackfillRankPeriod,
			loop.TaskSpiderRebuildCastRank,
			loop.TaskSpiderRebuildActorRank,
			loop.TaskScRebuildStats,
		},
		TaskButtons: []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderBackfillRankPeriod, Label: "周期排行回填"},
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
		"Labels":            cfg.Labels,
		"ScRootDir":         h.scRootDir,
		"JobID":             strings.TrimSpace(c.Query("job_id")),
	})
}
