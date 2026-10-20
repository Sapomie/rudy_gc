package crawler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/fetchsite"
)

type FetchSiteConfigAPI struct {
	service *fetchsite.Service
}

type fetchSiteConfigResponse struct {
	ID             int64  `json:"id"`
	SiteCode       string `json:"site_code"`
	SiteName       string `json:"site_name"`
	BaseURL        string `json:"base_url"`
	UserAgent      string `json:"user_agent"`
	Cookie         string `json:"cookie"`
	Proxy          string `json:"proxy"`
	TimeoutSeconds int64  `json:"timeout_seconds"`
	RequestSleepMs int64  `json:"request_sleep_ms"`
	RetrySleepMs   int64  `json:"retry_sleep_ms"`
	MaxRetryTimes  int64  `json:"max_retry_times"`
	Status         int64  `json:"status"`
	CreatedOn      int64  `json:"created_on"`
	UpdatedOn      int64  `json:"updated_on"`
	Persisted      bool   `json:"persisted"`
}

type updateFetchSiteConfigRequest struct {
	SiteName       string `json:"site_name"`
	BaseURL        string `json:"base_url"`
	UserAgent      string `json:"user_agent"`
	Cookie         string `json:"cookie"`
	Proxy          string `json:"proxy"`
	TimeoutSeconds int64  `json:"timeout_seconds"`
	RequestSleepMs int64  `json:"request_sleep_ms"`
	RetrySleepMs   int64  `json:"retry_sleep_ms"`
	MaxRetryTimes  int64  `json:"max_retry_times"`
	Status         int64  `json:"status"`
}

func NewFetchSiteConfigAPI(deps *dep.Dep) *FetchSiteConfigAPI {
	return &FetchSiteConfigAPI{service: fetchsite.NewService(deps)}
}

func (h *FetchSiteConfigAPI) List(c *gin.Context) {
	items, err := h.service.ListConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	out := make([]fetchSiteConfigResponse, 0, len(items))
	for _, item := range items {
		out = append(out, fetchSiteConfigResponseFromView(item))
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "items": out})
}

func (h *FetchSiteConfigAPI) Update(c *gin.Context) {
	siteCode := strings.TrimSpace(c.Param("site_code"))
	if siteCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 site_code 参数"})
		return
	}

	var req updateFetchSiteConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	item, err := h.service.UpdateConfig(c.Request.Context(), fetchsite.UpdateConfigInput{
		SiteCode:       siteCode,
		SiteName:       req.SiteName,
		BaseURL:        req.BaseURL,
		UserAgent:      req.UserAgent,
		Cookie:         req.Cookie,
		Proxy:          req.Proxy,
		TimeoutSeconds: req.TimeoutSeconds,
		RequestSleepMs: req.RequestSleepMs,
		RetrySleepMs:   req.RetrySleepMs,
		MaxRetryTimes:  req.MaxRetryTimes,
		Status:         req.Status,
	})
	if err != nil {
		if errors.Is(err, fetchsite.ErrInvalidConfig) {
			message := strings.TrimPrefix(err.Error(), fetchsite.ErrInvalidConfig.Error()+": ")
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "item": fetchSiteConfigResponseFromView(item)})
}

func fetchSiteConfigResponseFromView(view fetchsite.ConfigView) fetchSiteConfigResponse {
	cfg := view.Config
	return fetchSiteConfigResponse{
		ID:             view.ID,
		SiteCode:       cfg.SiteCode,
		SiteName:       cfg.SiteName,
		BaseURL:        cfg.BaseURL,
		UserAgent:      cfg.UserAgent,
		Cookie:         cfg.Cookie,
		Proxy:          cfg.Proxy,
		TimeoutSeconds: int64(cfg.Timeout.Seconds()),
		RequestSleepMs: cfg.RequestSleep.Milliseconds(),
		RetrySleepMs:   cfg.RetrySleep.Milliseconds(),
		MaxRetryTimes:  cfg.MaxRetryTimes,
		Status:         cfg.Status,
		CreatedOn:      view.CreatedOn,
		UpdatedOn:      view.UpdatedOn,
		Persisted:      view.Persisted,
	}
}
