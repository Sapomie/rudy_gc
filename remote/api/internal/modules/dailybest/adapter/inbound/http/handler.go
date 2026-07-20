package dailybestremotehttp

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"rudy_gc/remote/api/internal/modules/dailybest/domain"
	contracts "rudy_gc/shared/dailybest"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Ready(c *gin.Context) {
	if err := h.repo.Ping(c.Request.Context()); err != nil {
		writeError(c, http.StatusServiceUnavailable, "database_unavailable", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *Handler) ListSnapshots(c *gin.Context) {
	revision, err := parseNonNegativeInt64(c.Query("after_revision"), "after_revision")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_after_revision", err.Error())
		return
	}
	rows, err := h.repo.ListSnapshotsAfter(c.Request.Context(), revision)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list_snapshots_failed", err.Error())
		return
	}
	latest := revision
	if len(rows) > 0 {
		latest = rows[len(rows)-1].Revision
	}
	c.JSON(http.StatusOK, contracts.DailyBestSnapshotListResponse{Snapshots: rows, LatestRevision: latest})
}

func (h *Handler) GetSnapshot(c *gin.Context) {
	revision, err := parsePositiveInt64(c.Param("revision"), "revision")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_revision", err.Error())
		return
	}
	row, err := h.repo.FindSnapshot(c.Request.Context(), revision, true)
	if err != nil {
		writeRepositoryError(c, "snapshot", err)
		return
	}
	c.JSON(http.StatusOK, row)
}

func (h *Handler) GetSnapshotPages(c *gin.Context) {
	revision, err := parsePositiveInt64(c.Param("revision"), "revision")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_revision", err.Error())
		return
	}
	row, err := h.repo.FindSnapshot(c.Request.Context(), revision, true)
	if err != nil {
		writeRepositoryError(c, "snapshot", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"revision": row.Revision, "pages": row.Pages})
}

func (h *Handler) GetSnapshotByDate(c *gin.Context) {
	businessDate := strings.TrimSpace(c.Param("date"))
	category, err := parseOptionalCategory(c.Query("category"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_category", err.Error())
		return
	}
	row, err := h.repo.FindLatestSnapshotByDate(c.Request.Context(), businessDate, category, true)
	if err != nil {
		writeRepositoryError(c, "snapshot", err)
		return
	}
	c.JSON(http.StatusOK, row)
}

func (h *Handler) StartFetch(c *gin.Context) {
	var req contracts.DailyBestFetchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "请求内容错误")
		return
	}
	run, err := h.collector.Enqueue(c.Request.Context(), req, dailybest.TriggerManual)
	if err != nil {
		writeError(c, http.StatusBadRequest, "fetch_rejected", err.Error())
		return
	}
	c.JSON(http.StatusAccepted, contracts.DailyBestFetchAcceptedResponse{
		RunID: run.ID, Status: run.Status, Message: "DailyBest 抓取任务已创建",
	})
}

func (h *Handler) ListFetchRuns(c *gin.Context) {
	runs, err := h.repo.ListFetchRuns(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list_fetch_runs_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, contracts.DailyBestFetchRunListResponse{Runs: runs})
}

func (h *Handler) GetFetchRun(c *gin.Context) {
	runID, err := parsePositiveInt64(c.Param("id"), "id")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_run_id", err.Error())
		return
	}
	run, err := h.repo.FindFetchRun(c.Request.Context(), runID)
	if err != nil {
		writeRepositoryError(c, "fetch_run", err)
		return
	}
	c.JSON(http.StatusOK, run)
}

func parseOptionalCategory(raw string) (int64, error) {
	if strings.TrimSpace(raw) == "" {
		return contracts.DailyBestCategoryMonth, nil
	}
	value, err := parsePositiveInt64(raw, "category")
	if err != nil {
		return 0, err
	}
	if value != contracts.DailyBestCategoryMonth {
		return 0, errors.New("当前只支持月榜 category=1")
	}
	return value, nil
}

func parseNonNegativeInt64(raw, field string) (int64, error) {
	if strings.TrimSpace(raw) == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value < 0 {
		return 0, errors.New(field + " 必须为非负整数")
	}
	return value, nil
}

func parsePositiveInt64(raw, field string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value <= 0 {
		return 0, errors.New(field + " 必须为正整数")
	}
	return value, nil
}

func writeRepositoryError(c *gin.Context, resource string, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		writeError(c, http.StatusNotFound, resource+"_not_found", "数据不存在")
		return
	}
	writeError(c, http.StatusInternalServerError, resource+"_query_failed", err.Error())
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, contracts.DailyBestAPIError{Code: code, Message: message})
}
