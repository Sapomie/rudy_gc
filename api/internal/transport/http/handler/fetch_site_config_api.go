package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/repository/moviex"
	"rudy_gc/internal/svc"
)

type FetchSiteConfigAPI struct {
	deps *dep.Dep
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
	return &FetchSiteConfigAPI{deps: deps}
}

func (h *FetchSiteConfigAPI) List(c *gin.Context) {
	rowsByCode := make(map[string]*moviex.TFetchSite)
	if h.deps != nil && h.deps.FetchSiteModel != nil {
		rows, err := h.deps.FetchSiteModel.ListAll(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
			return
		}
		for _, row := range rows {
			if row == nil {
				continue
			}
			rowsByCode[strings.TrimSpace(row.SiteCode)] = row
		}
	}

	out := make([]fetchSiteConfigResponse, 0, len(rowsByCode)+len(h.deps.ListFetchSiteConfigs()))
	seen := make(map[string]struct{}, len(rowsByCode))
	for _, cfg := range h.deps.ListFetchSiteConfigs() {
		siteCode := strings.TrimSpace(cfg.SiteCode)
		if siteCode == "" {
			continue
		}
		if row := rowsByCode[siteCode]; row != nil {
			out = append(out, fetchSiteConfigResponseFromRow(row, true))
			seen[siteCode] = struct{}{}
			continue
		}
		out = append(out, fetchSiteConfigResponseFromConfig(cfg, false))
		seen[siteCode] = struct{}{}
	}
	for siteCode, row := range rowsByCode {
		if _, ok := seen[siteCode]; ok || row == nil {
			continue
		}
		out = append(out, fetchSiteConfigResponseFromRow(row, true))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].SiteCode < out[j].SiteCode
	})

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

	fallback, _ := h.deps.GetFetchSiteConfig(siteCode)
	cfg, err := fetchSiteConfigFromUpdateRequest(siteCode, req, fallback)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	row, err := upsertFetchSiteConfig(c.Request.Context(), h.deps, cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "item": fetchSiteConfigResponseFromRow(row, true)})
}

func fetchSiteConfigFromUpdateRequest(siteCode string, req updateFetchSiteConfigRequest, fallback svc.FetchSiteConfig) (svc.FetchSiteConfig, error) {
	siteCode = strings.TrimSpace(siteCode)
	siteName := strings.TrimSpace(req.SiteName)
	if siteName == "" {
		siteName = strings.TrimSpace(fallback.SiteName)
	}
	baseURL := svc.NormalizeFetchSiteBaseURL(req.BaseURL)
	if siteName == "" {
		return svc.FetchSiteConfig{}, errors.New("site_name 不能为空")
	}
	if baseURL == "" {
		return svc.FetchSiteConfig{}, errors.New("base_url 不能为空")
	}
	if err := validateFetchSiteBaseURL(baseURL); err != nil {
		return svc.FetchSiteConfig{}, err
	}
	if req.TimeoutSeconds <= 0 {
		return svc.FetchSiteConfig{}, errors.New("timeout_seconds 必须大于 0")
	}
	if req.RequestSleepMs < 0 {
		return svc.FetchSiteConfig{}, errors.New("request_sleep_ms 不能小于 0")
	}
	if req.RetrySleepMs < 0 {
		return svc.FetchSiteConfig{}, errors.New("retry_sleep_ms 不能小于 0")
	}
	if req.MaxRetryTimes <= 0 {
		return svc.FetchSiteConfig{}, errors.New("max_retry_times 必须大于 0")
	}
	return svc.FetchSiteConfig{
		SiteCode:      siteCode,
		SiteName:      siteName,
		BaseURL:       baseURL,
		UserAgent:     strings.TrimSpace(req.UserAgent),
		Cookie:        strings.TrimSpace(req.Cookie),
		Proxy:         strings.TrimSpace(req.Proxy),
		Timeout:       time.Duration(req.TimeoutSeconds) * time.Second,
		RequestSleep:  time.Duration(req.RequestSleepMs) * time.Millisecond,
		RetrySleep:    time.Duration(req.RetrySleepMs) * time.Millisecond,
		MaxRetryTimes: req.MaxRetryTimes,
		Status:        req.Status,
	}, nil
}

func validateFetchSiteBaseURL(baseURL string) error {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("base_url 必须是完整 URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("base_url 只支持 http 或 https")
	}
	return nil
}

func upsertFetchSiteConfig(ctx context.Context, deps *dep.Dep, cfg svc.FetchSiteConfig) (*moviex.TFetchSite, error) {
	if deps == nil || deps.FetchSiteModel == nil {
		return nil, errors.New("fetch site model 未初始化")
	}
	now := time.Now().Unix()
	row, err := deps.FetchSiteModel.FindOneBySiteCode(ctx, cfg.SiteCode)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, err
	}
	if err == nil && row != nil {
		applyFetchSiteConfigToRow(row, cfg)
		row.UpdatedOn = now
		if err := deps.FetchSiteModel.Update(ctx, row); err != nil {
			return nil, err
		}
		deps.SetFetchSiteConfig(cfg)
		return row, nil
	}

	row = &moviex.TFetchSite{
		SiteCode:  cfg.SiteCode,
		CreatedOn: now,
		UpdatedOn: now,
	}
	applyFetchSiteConfigToRow(row, cfg)
	if _, err := deps.FetchSiteModel.Insert(ctx, row); err != nil {
		return nil, err
	}
	created, err := deps.FetchSiteModel.FindOneBySiteCode(ctx, cfg.SiteCode)
	if err != nil {
		return nil, err
	}
	deps.SetFetchSiteConfig(cfg)
	return created, nil
}

func applyFetchSiteConfigToRow(row *moviex.TFetchSite, cfg svc.FetchSiteConfig) {
	row.SiteCode = cfg.SiteCode
	row.SiteName = cfg.SiteName
	row.BaseUrl = cfg.BaseURL
	row.UserAgent = cfg.UserAgent
	row.Cookie = cfg.Cookie
	row.Proxy = cfg.Proxy
	row.TimeoutSeconds = int64(cfg.Timeout / time.Second)
	row.RequestSleepMs = int64(cfg.RequestSleep / time.Millisecond)
	row.RetrySleepMs = int64(cfg.RetrySleep / time.Millisecond)
	row.MaxRetryTimes = cfg.MaxRetryTimes
	row.Status = cfg.Status
}

func fetchSiteConfigResponseFromRow(row *moviex.TFetchSite, persisted bool) fetchSiteConfigResponse {
	if row == nil {
		return fetchSiteConfigResponse{}
	}
	return fetchSiteConfigResponse{
		ID:             row.Id,
		SiteCode:       row.SiteCode,
		SiteName:       row.SiteName,
		BaseURL:        row.BaseUrl,
		UserAgent:      row.UserAgent,
		Cookie:         row.Cookie,
		Proxy:          row.Proxy,
		TimeoutSeconds: row.TimeoutSeconds,
		RequestSleepMs: row.RequestSleepMs,
		RetrySleepMs:   row.RetrySleepMs,
		MaxRetryTimes:  row.MaxRetryTimes,
		Status:         row.Status,
		CreatedOn:      row.CreatedOn,
		UpdatedOn:      row.UpdatedOn,
		Persisted:      persisted,
	}
}

func fetchSiteConfigResponseFromConfig(cfg svc.FetchSiteConfig, persisted bool) fetchSiteConfigResponse {
	return fetchSiteConfigResponse{
		SiteCode:       cfg.SiteCode,
		SiteName:       cfg.SiteName,
		BaseURL:        cfg.BaseURL,
		UserAgent:      cfg.UserAgent,
		Cookie:         cfg.Cookie,
		Proxy:          cfg.Proxy,
		TimeoutSeconds: int64(cfg.Timeout / time.Second),
		RequestSleepMs: int64(cfg.RequestSleep / time.Millisecond),
		RetrySleepMs:   int64(cfg.RetrySleep / time.Millisecond),
		MaxRetryTimes:  cfg.MaxRetryTimes,
		Status:         cfg.Status,
		Persisted:      persisted,
	}
}
