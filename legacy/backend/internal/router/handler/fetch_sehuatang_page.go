package handler

import (
	"rudy_gc/internal/service/fetchsehuatang"
	"rudy_gc/internal/service/loop"

	"github.com/gin-gonic/gin"
)

func (h *CrawlerPages) FetchSehuatangPage(c *gin.Context) {
	h.renderJobsPage(c, crawlerJobsPageConfig{
		Title:             "色花堂抓取任务",
		PageTitle:         "色花堂磁力抓取（日志流）",
		QuickNavCurrent:   "fetch_sehuatang",
		TaskPanelTitle:    "色花堂抓取任务",
		PageNote:          "抓取列表普通帖子并直接入库，自动跳过置顶/公告帖；支持按起始页到结束页范围抓取。",
		TaskTableTitle:    "色花堂抓取任务",
		EventTitle:        "色花堂抓取事件流",
		StorageKey:        "crawler_jobs_fetch_sehuatang_selected_job",
		DefaultTaskType:   loop.TaskSpiderFetchSehuatang,
		OverviewExtraMode: "sehuatang_progress",
		EmptyStateText:    "等待 色花堂抓取任务触发",
		AllowedTaskTypes: []string{
			loop.TaskSpiderFetchSehuatang,
		},
		TaskButtons: []crawlerJobsPageTask{
			{TaskType: loop.TaskSpiderFetchSehuatang, Label: "色花堂磁力抓取"},
		},
		HeaderLinks: []crawlerJobsPageLink{
			{Href: "/triggers/fetch-site", Label: "FetchSite"},
			{Href: "/fetch-site-sehuatang-list", Label: "Sehuatang 列表"},
		},
		Labels: crawlerJobsPageLabels{
			Extra:   "任务类型",
			Result:  "成功/失败",
			Elapsed: "已运行时长",
		},
		FetchSehuatangList:        fetchsehuatang.DefaultListURL,
		FetchSehuatangKey:         fetchsehuatang.DefaultKeyword,
		FetchSehuatangStartPage:   fetchsehuatang.DefaultStartPage,
		FetchSehuatangEndPage:     fetchsehuatang.DefaultEndPage,
		FetchSehuatangPersistMode: fetchsehuatang.DefaultPersistMode,
	})
}
