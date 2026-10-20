package fetchsite

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	fetchsiterepo "rudy_gc/local/api/internal/repo/fetchsite"
	"rudy_gc/local/api/internal/service/spider/fetcher"
)

var ErrInvalidConfig = errors.New("invalid fetch site config")

type ConfigView struct {
	Config    SiteConfig
	ID        int64
	CreatedOn int64
	UpdatedOn int64
	Persisted bool
}

type UpdateConfigInput struct {
	SiteCode       string
	SiteName       string
	BaseURL        string
	UserAgent      string
	Cookie         string
	Proxy          string
	TimeoutSeconds int64
	RequestSleepMs int64
	RetrySleepMs   int64
	MaxRetryTimes  int64
	Status         int64
}

func (s *Service) ListConfigs(ctx context.Context) ([]ConfigView, error) {
	if s == nil || s.deps == nil || s.deps.FetchSiteRegistry == nil {
		return nil, errors.New("fetch site registry 未初始化")
	}

	rowsByCode := make(map[string]*moviex.TFetchSite)
	if s.deps.FetchSiteModel != nil {
		rows, err := s.deps.FetchSiteModel.ListAll(ctx)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			if row == nil {
				continue
			}
			rowsByCode[strings.TrimSpace(row.SiteCode)] = row
		}
	}

	configs := s.deps.FetchSiteRegistry.List()
	out := make([]ConfigView, 0, len(rowsByCode)+len(configs))
	seen := make(map[string]struct{}, len(configs))
	for _, cfg := range configs {
		siteCode := strings.TrimSpace(cfg.SiteCode)
		if siteCode == "" {
			continue
		}
		if row := rowsByCode[siteCode]; row != nil {
			out = append(out, configViewFromRow(row))
		} else {
			out = append(out, ConfigView{Config: cfg})
		}
		seen[siteCode] = struct{}{}
	}
	for siteCode, row := range rowsByCode {
		if _, ok := seen[siteCode]; ok || row == nil {
			continue
		}
		out = append(out, configViewFromRow(row))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Config.SiteCode < out[j].Config.SiteCode
	})
	return out, nil
}

func (s *Service) UpdateConfig(ctx context.Context, input UpdateConfigInput) (ConfigView, error) {
	if s == nil || s.deps == nil || s.deps.FetchSiteModel == nil {
		return ConfigView{}, errors.New("fetch site model 未初始化")
	}

	fallback, _ := s.deps.FetchSiteRegistry.Get(input.SiteCode)
	cfg, err := configFromUpdateInput(input, fallback)
	if err != nil {
		return ConfigView{}, err
	}

	now := time.Now().Unix()
	row, err := s.deps.FetchSiteModel.FindOneBySiteCode(ctx, cfg.SiteCode)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return ConfigView{}, err
	}
	if err == nil && row != nil {
		applyConfigToRow(row, cfg)
		row.UpdatedOn = now
		if err := s.deps.FetchSiteModel.Update(ctx, row); err != nil {
			return ConfigView{}, err
		}
	} else {
		row = &moviex.TFetchSite{
			SiteCode:  cfg.SiteCode,
			CreatedOn: now,
			UpdatedOn: now,
		}
		applyConfigToRow(row, cfg)
		if _, err := s.deps.FetchSiteModel.Insert(ctx, row); err != nil {
			return ConfigView{}, err
		}
		row, err = s.deps.FetchSiteModel.FindOneBySiteCode(ctx, cfg.SiteCode)
		if err != nil {
			return ConfigView{}, err
		}
	}

	s.applyRuntimeConfig(cfg)
	return configViewFromRow(row), nil
}

func configFromUpdateInput(input UpdateConfigInput, fallback SiteConfig) (SiteConfig, error) {
	siteCode := strings.TrimSpace(input.SiteCode)
	siteName := strings.TrimSpace(input.SiteName)
	if siteName == "" {
		siteName = strings.TrimSpace(fallback.SiteName)
	}
	baseURL := fetchsiterepo.NormalizeBaseURL(input.BaseURL)
	if siteCode == "" {
		return SiteConfig{}, invalidConfigError("site_code 不能为空")
	}
	if siteName == "" {
		return SiteConfig{}, invalidConfigError("site_name 不能为空")
	}
	if baseURL == "" {
		return SiteConfig{}, invalidConfigError("base_url 不能为空")
	}
	if err := validateBaseURL(baseURL); err != nil {
		return SiteConfig{}, err
	}
	if input.TimeoutSeconds <= 0 {
		return SiteConfig{}, invalidConfigError("timeout_seconds 必须大于 0")
	}
	if input.RequestSleepMs < 0 {
		return SiteConfig{}, invalidConfigError("request_sleep_ms 不能小于 0")
	}
	if input.RetrySleepMs < 0 {
		return SiteConfig{}, invalidConfigError("retry_sleep_ms 不能小于 0")
	}
	if input.MaxRetryTimes <= 0 {
		return SiteConfig{}, invalidConfigError("max_retry_times 必须大于 0")
	}
	return SiteConfig{
		SiteCode:      siteCode,
		SiteName:      siteName,
		BaseURL:       baseURL,
		UserAgent:     strings.TrimSpace(input.UserAgent),
		Cookie:        strings.TrimSpace(input.Cookie),
		Proxy:         strings.TrimSpace(input.Proxy),
		Timeout:       time.Duration(input.TimeoutSeconds) * time.Second,
		RequestSleep:  time.Duration(input.RequestSleepMs) * time.Millisecond,
		RetrySleep:    time.Duration(input.RetrySleepMs) * time.Millisecond,
		MaxRetryTimes: input.MaxRetryTimes,
		Status:        input.Status,
	}, nil
}

func validateBaseURL(baseURL string) error {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return invalidConfigError("base_url 必须是完整 URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return invalidConfigError("base_url 只支持 http 或 https")
	}
	return nil
}

func invalidConfigError(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidConfig, message)
}

func applyConfigToRow(row *moviex.TFetchSite, cfg SiteConfig) {
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

func configViewFromRow(row *moviex.TFetchSite) ConfigView {
	if row == nil {
		return ConfigView{}
	}
	return ConfigView{
		Config:    fetchsiterepo.FromRow(row),
		ID:        row.Id,
		CreatedOn: row.CreatedOn,
		UpdatedOn: row.UpdatedOn,
		Persisted: true,
	}
}

func (s *Service) applyRuntimeConfig(cfg SiteConfig) {
	if s.deps.FetchSiteRegistry != nil {
		s.deps.FetchSiteRegistry.Set(cfg)
	}
	if s.deps.Fetcher != nil {
		s.deps.Fetcher.SetSiteConfig(cfg.SiteCode, fetcher.Config{
			UserAgent: cfg.UserAgent,
			Cookie:    cfg.Cookie,
			Proxy:     cfg.Proxy,
			Timeout:   cfg.Timeout,
		})
	}
}
