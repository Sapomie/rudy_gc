package dailybest

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	dailybestservice "rudy_gc/remote/api/internal/service/dailybest"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

type targetProbeListResponse struct {
	Probes   []dailybesttypes.TargetProbe `json:"probes"`
	PageInfo *contracts.DailyBestPageInfo `json:"page_info,omitempty"`
}

type runtimeEventListResponse struct {
	Events   []dailybesttypes.RuntimeEvent `json:"events"`
	PageInfo *contracts.DailyBestPageInfo  `json:"page_info,omitempty"`
}

func (h *DailyBestAPI) RuntimeStatus(c *gin.Context) {
	c.JSON(http.StatusOK, h.service.RuntimeStatus(c.Request.Context()))
}

func (h *DailyBestAPI) StartTargetProbe(c *gin.Context) {
	probe, err := h.service.TriggerTargetProbe(c.Request.Context())
	if err != nil {
		if errors.Is(err, dailybestservice.ErrActiveTargetProbe) {
			writeError(c, http.StatusConflict, "active_target_probe_exists", "当前已经存在目标站探测")
			return
		}
		writeError(c, http.StatusServiceUnavailable, "target_probe_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, probe)
}

func (h *DailyBestAPI) ListTargetProbes(c *gin.Context) {
	page, err := parsePageQuery(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}
	status, err := parseTargetProbeStatus(c.Query("status"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_target_probe_status", err.Error())
		return
	}
	startedFrom, err := parseNonNegativeInt64(c.Query("started_from"), "started_from")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_started_from", err.Error())
		return
	}
	startedTo, err := parseNonNegativeInt64(c.Query("started_to"), "started_to")
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_started_to", err.Error())
		return
	}
	if startedFrom > 0 && startedTo > 0 && startedFrom > startedTo {
		writeError(c, http.StatusBadRequest, "invalid_started_range", "started_from 不能晚于 started_to")
		return
	}
	probes, total, err := h.service.ListTargetProbes(c.Request.Context(), dailybesttypes.TargetProbeFilter{
		PageQuery: page, Status: status, StartedFrom: startedFrom, StartedTo: startedTo,
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list_target_probes_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, targetProbeListResponse{Probes: probes, PageInfo: buildPageInfo(total, page)})
}

func (h *DailyBestAPI) ListRuntimeEvents(c *gin.Context) {
	page, err := parsePageQuery(c)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}
	component := strings.TrimSpace(c.Query("component"))
	level := strings.TrimSpace(c.Query("level"))
	if level != "" && level != dailybesttypes.RuntimeLevelInfo && level != dailybesttypes.RuntimeLevelError {
		writeError(c, http.StatusBadRequest, "invalid_runtime_level", "level 必须为 info 或 error")
		return
	}
	events, total, err := h.service.ListRuntimeEvents(c.Request.Context(), dailybesttypes.RuntimeEventFilter{
		PageQuery: page, Component: component, Level: level,
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list_runtime_events_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, runtimeEventListResponse{Events: events, PageInfo: buildPageInfo(total, page)})
}

func parseTargetProbeStatus(raw string) (string, error) {
	status := strings.TrimSpace(raw)
	switch status {
	case "", dailybesttypes.TargetProbeStatusRunning, dailybesttypes.TargetProbeStatusSucceeded,
		dailybesttypes.TargetProbeStatusFailed, dailybesttypes.TargetProbeStatusInterrupted:
		return status, nil
	default:
		return "", fmt.Errorf("status 必须为 running、succeeded、failed 或 interrupted")
	}
}
