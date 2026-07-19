package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/dep"
	moviedomain "rudy_gc/internal/domain/movie"
	"rudy_gc/internal/repository/moviex"
	"rudy_gc/internal/service/loop"
)

type CrawlerAPI struct {
	runtime *loop.FetchLoopService
	deps    *dep.Dep
}

func NewCrawlerAPI(deps *dep.Dep) *CrawlerAPI {
	return &CrawlerAPI{
		runtime: newCrawlerRuntime(deps),
		deps:    deps,
	}
}

func (h *CrawlerAPI) Start(c *gin.Context) {
	var req loop.StartTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeCrawlerError(c, http.StatusBadRequest, "invalid request")
		return
	}
	jobID, err := h.runtime.StartTask(req)
	if err != nil {
		writeCrawlerError(c, http.StatusBadRequest, err.Error())
		return
	}
	writeCrawlerJobStarted(c, jobID, strings.TrimSpace(req.TaskType))
}

func (h *CrawlerAPI) PreviewFetchSite(c *gin.Context) {
	var req loop.StartTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeCrawlerError(c, http.StatusBadRequest, "invalid request")
		return
	}

	result, err := h.runtime.PreviewFetchSiteTargets(c.Request.Context(), req)
	if err != nil {
		writeCrawlerError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CrawlerAPI) PreviewSehuatangSHThreadDetails(c *gin.Context) {
	var req loop.StartTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeCrawlerError(c, http.StatusBadRequest, "invalid request")
		return
	}
	detailStatuses := normalizeSHDetailStatuses(req.DetailStatuses)
	if len(detailStatuses) == 0 {
		detailStatus := strings.TrimSpace(req.DetailStatus)
		if detailStatus == "" {
			detailStatus = strings.TrimSpace(req.Status)
		}
		if detailStatus != "" {
			detailStatuses = []string{detailStatus}
		}
	}
	movieTypes := moviedomain.NormalizeSehuatangThreadMovieTypes(append(req.MovieTypes, req.MovieType)...)
	if req.Limit < 0 {
		writeCrawlerError(c, http.StatusBadRequest, "limit 必须大于等于 0")
		return
	}
	targetForumIDs, err := normalizeSHFillTargetForumIDs(req.TargetForumIDs)
	if err != nil {
		writeCrawlerError(c, http.StatusBadRequest, err.Error())
		return
	}
	filter := moviex.ShSehuatangThreadListFilter{
		Keyword:        strings.TrimSpace(req.Keyword),
		MovieTypes:     movieTypes,
		SourceForumIDs: targetForumIDs,
		DetailStatuses: detailStatuses,
	}
	if ts, ok, err := parseOptionalDateStart(req.PostTimeFrom); err != nil {
		writeCrawlerError(c, http.StatusBadRequest, fmt.Sprintf("发帖时间开始日期错误: %v", err))
		return
	} else if ok {
		filter.PostTimeFrom = ts
		filter.HasPostTimeFrom = true
	}
	if ts, ok, err := parseOptionalDateEnd(req.PostTimeTo); err != nil {
		writeCrawlerError(c, http.StatusBadRequest, fmt.Sprintf("发帖时间结束日期错误: %v", err))
		return
	} else if ok {
		filter.PostTimeTo = ts
		filter.HasPostTimeTo = true
	}

	aggregate, err := h.deps.ShSehuatangThreadModel.Aggregate(c.Request.Context(), filter)
	if err != nil {
		writeCrawlerError(c, http.StatusInternalServerError, err.Error())
		return
	}
	rows, err := h.deps.ShSehuatangThreadModel.ListPage(c.Request.Context(), 0, req.Limit, "`post_time` DESC, `id` DESC", filter)
	if err != nil {
		writeCrawlerError(c, http.StatusInternalServerError, err.Error())
		return
	}
	threadTIDs := make([]string, 0, len(rows))
	seenThreadTIDs := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		threadTID := strings.TrimSpace(row.ThreadTID)
		if threadTID == "" {
			continue
		}
		if _, ok := seenThreadTIDs[threadTID]; ok {
			continue
		}
		seenThreadTIDs[threadTID] = struct{}{}
		threadTIDs = append(threadTIDs, threadTID)
	}
	c.JSON(http.StatusOK, gin.H{
		"total":          aggregate.Total,
		"selected_count": len(threadTIDs),
		"limit":          req.Limit,
		"thread_tids":    threadTIDs,
		"aggregate": gin.H{
			"total":                aggregate.Total,
			"pending_detail_count": aggregate.PendingDetailCount,
			"fetched_detail_count": aggregate.FetchedDetailCount,
			"failed_detail_count":  aggregate.FailedDetailCount,
			"forum_counts": gin.H{
				"36":  aggregate.Forum36Count,
				"103": aggregate.Forum103Count,
			},
		},
	})
}

func normalizeSHFillTargetForumIDs(values []string) ([]string, error) {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("请选择目标版块")
	}
	return out, nil
}

func normalizeSHDetailStatuses(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func (h *CrawlerAPI) ListJobs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"jobs":         h.runtime.ListAllJobs(),
		"running_jobs": h.runtime.ListJobs(),
		"detail_loop":  h.runtime.GetDetailLoopSnapshot(),
	})
}

func (h *CrawlerAPI) GetJob(c *gin.Context) {
	jobID := strings.TrimSpace(c.Param("jobID"))
	if jobID == "" {
		writeCrawlerError(c, http.StatusBadRequest, "job_id is required")
		return
	}

	job, err := h.runtime.GetJob(jobID)
	if err != nil {
		writeCrawlerError(c, http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, job)
}

func (h *CrawlerAPI) GetJobEvents(c *gin.Context) {
	jobID := strings.TrimSpace(c.Param("jobID"))
	if jobID == "" {
		writeCrawlerError(c, http.StatusBadRequest, "job_id is required")
		return
	}

	events, err := h.runtime.GetJobEvents(jobID, parseCrawlerLastEventID(c))
	if err != nil {
		writeCrawlerError(c, http.StatusNotFound, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"events": events,
	})
}

func (h *CrawlerAPI) Pause(c *gin.Context) {
	h.controlJob(c, "任务已暂停", func(jobID string) error {
		return h.runtime.PauseJob(jobID)
	})
}

func (h *CrawlerAPI) Resume(c *gin.Context) {
	h.controlJob(c, "任务已继续", func(jobID string) error {
		return h.runtime.ResumeJob(jobID)
	})
}

func (h *CrawlerAPI) Stop(c *gin.Context) {
	h.controlJob(c, "任务已停止", func(jobID string) error {
		return h.runtime.StopJob(jobID)
	})
}

func (h *CrawlerAPI) Stream(c *gin.Context) {
	jobID := strings.TrimSpace(c.Param("jobID"))
	if jobID == "" {
		writeCrawlerError(c, http.StatusBadRequest, "job_id is required")
		return
	}

	history, eventCh, cancel, err := h.runtime.SubscribeJob(jobID, parseCrawlerLastEventID(c))
	if err != nil {
		writeCrawlerError(c, http.StatusNotFound, err.Error())
		return
	}
	defer cancel()

	streamCrawlerJobEvents(c, history, eventCh, true)
}

func (h *CrawlerAPI) StreamAll(c *gin.Context) {
	history, eventCh, cancel := h.runtime.SubscribeAllJobs(parseCrawlerLastEventID(c))
	defer cancel()

	streamCrawlerJobEvents(c, history, eventCh, false)
}

func streamCrawlerJobEvents(c *gin.Context, history []loop.JobEvent, eventCh <-chan loop.JobEvent, stopOnDone bool) {
	flusher, ok := prepareCrawlerEventStream(c)
	if !ok {
		return
	}

	fmt.Fprint(c.Writer, ": connected\n\n")
	flusher.Flush()

	for _, event := range history {
		if !writeCrawlerJobEvent(c.Writer, flusher, event) {
			continue
		}
		if stopOnDone && event.Done {
			return
		}
	}

	pingTicker := time.NewTicker(15 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case event, ok := <-eventCh:
			if !ok {
				return
			}
			if !writeCrawlerJobEvent(c.Writer, flusher, event) {
				continue
			}
			if stopOnDone && event.Done {
				return
			}
		case <-pingTicker.C:
			fmt.Fprint(c.Writer, ": ping\n\n")
			flusher.Flush()
		}
	}
}

func prepareCrawlerEventStream(c *gin.Context) (http.Flusher, bool) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	_ = http.NewResponseController(c.Writer).SetWriteDeadline(time.Time{})

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		writeCrawlerError(c, http.StatusInternalServerError, "streaming unsupported")
		return nil, false
	}
	return flusher, true
}

func writeCrawlerJobEvent(w http.ResponseWriter, flusher http.Flusher, event loop.JobEvent) bool {
	payload, err := json.Marshal(event)
	if err != nil {
		return false
	}
	if event.ID > 0 {
		fmt.Fprintf(w, "id: %d\n", event.ID)
	}
	fmt.Fprintf(w, "data: %s\n\n", payload)
	flusher.Flush()
	return true
}

func parseCrawlerLastEventID(c *gin.Context) int64 {
	if c == nil || c.Request == nil {
		return 0
	}
	raw := strings.TrimSpace(c.Request.Header.Get("Last-Event-ID"))
	if raw == "" {
		raw = strings.TrimSpace(c.Query("last_event_id"))
	}
	if raw == "" {
		return 0
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func (h *CrawlerAPI) GetDetailLoop(c *gin.Context) {
	c.JSON(http.StatusOK, h.runtime.GetDetailLoopSnapshot())
}

func (h *CrawlerAPI) StartDetailLoop(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":     "详情抓取 loop 已启动",
		"detail_loop": h.runtime.StartManagedDetailLoop(),
	})
}

func (h *CrawlerAPI) StopDetailLoop(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":     "详情抓取 loop 已停止",
		"detail_loop": h.runtime.StopManagedDetailLoop(),
	})
}

func (h *CrawlerAPI) StreamDetailLoop(c *gin.Context) {
	history, eventCh, cancel := h.runtime.SubscribeDetailLoop()
	defer cancel()

	flusher, ok := prepareCrawlerEventStream(c)
	if !ok {
		return
	}

	fmt.Fprint(c.Writer, ": connected\n\n")
	flusher.Flush()

	for _, event := range history {
		payload, err := json.Marshal(event)
		if err != nil {
			continue
		}
		fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
		flusher.Flush()
	}

	pingTicker := time.NewTicker(15 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case event := <-eventCh:
			payload, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", payload)
			flusher.Flush()
		case <-pingTicker.C:
			fmt.Fprint(c.Writer, ": ping\n\n")
			flusher.Flush()
		}
	}
}

func (h *CrawlerAPI) controlJob(c *gin.Context, message string, fn func(string) error) {
	jobID := strings.TrimSpace(c.Param("jobID"))
	if jobID == "" {
		writeCrawlerError(c, http.StatusBadRequest, "job_id is required")
		return
	}

	if err := fn(jobID); err != nil {
		writeCrawlerError(c, http.StatusBadRequest, err.Error())
		return
	}
	writeCrawlerJobAction(c, jobID, message)
}
