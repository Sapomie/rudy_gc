package dailybest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

type fetchLogResponse struct {
	ID            int64  `json:"id"`
	FetchRunID    int64  `json:"fetch_run_id"`
	PageNumber    int64  `json:"page_number"`
	AttemptNumber int64  `json:"attempt_number"`
	RequestURL    string `json:"request_url"`
	Status        string `json:"status"`
	HTTPStatus    int64  `json:"http_status"`
	DurationMS    int64  `json:"duration_ms"`
	ErrorMessage  string `json:"error_message"`
	CreatedOn     int64  `json:"created_on"`
	UpdatedOn     int64  `json:"updated_on"`
}

type fetchLogListResponse struct {
	Logs     []fetchLogResponse           `json:"logs"`
	PageInfo *contracts.DailyBestPageInfo `json:"page_info,omitempty"`
}

func (h *DailyBestAPI) ListFetchLogs(c *gin.Context) {
	page, err := parsePageQuery(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}
	fetchRunID, err := parseNonNegativeInt64(c.Query("fetch_run_id"), "fetch_run_id")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_fetch_run_id", err.Error())
		return
	}
	pageNumber, err := parseNonNegativeInt64(c.Query("page_number"), "page_number")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_page_number", err.Error())
		return
	}
	status, err := parseFetchLogStatus(c.Query("status"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_fetch_log_status", err.Error())
		return
	}
	logs, total, err := h.service.ListFetchLogs(c.Request.Context(), dailybesttypes.FetchLogFilter{
		PageQuery: page, FetchRunID: fetchRunID, PageNumber: pageNumber, Status: status,
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list_fetch_logs_failed", err.Error())
		return
	}
	responseLogs := make([]fetchLogResponse, 0, len(logs))
	for _, logRow := range logs {
		responseLogs = append(responseLogs, fetchLogResponse{
			ID: logRow.ID, FetchRunID: logRow.FetchRunID, PageNumber: logRow.PageNumber,
			AttemptNumber: logRow.AttemptNumber, RequestURL: logRow.RequestURL, Status: logRow.Status,
			HTTPStatus: logRow.HTTPStatus, DurationMS: logRow.DurationMS, ErrorMessage: logRow.ErrorMessage,
			CreatedOn: logRow.CreatedOn, UpdatedOn: logRow.UpdatedOn,
		})
	}
	c.JSON(http.StatusOK, fetchLogListResponse{Logs: responseLogs, PageInfo: buildPageInfo(total, page)})
}

func parseFetchLogStatus(raw string) (string, error) {
	status := strings.TrimSpace(raw)
	switch status {
	case "", dailybesttypes.FetchLogStatusRequesting, dailybesttypes.FetchLogStatusSucceeded, dailybesttypes.FetchLogStatusFailed:
		return status, nil
	default:
		return "", fmt.Errorf("status 必须为 requesting、succeeded 或 failed")
	}
}
