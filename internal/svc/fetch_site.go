package svc

import (
	"context"
	"strings"
	"time"

	"rudy_gc/internal/config"
	"rudy_gc/internal/model/modelx/moviex"

	"github.com/sirupsen/logrus"
)

const (
	FetchSiteCodeJavbus    = "javbus"
	FetchSiteCodeSukebei   = "sukebei"
	FetchSiteCodeSehuatang = "sehuatang"

	fetchSiteStatusEnabled    int64 = 1
	legacyFetchSiteRetryTimes int64 = 45
)

var (
	legacyFetchSiteRequestSleep = 3 * time.Second
	legacyFetchSiteRetrySleep   = 3 * time.Second
)

type FetchSiteConfig struct {
	SiteCode string
	SiteName string
	BaseURL  string

	UserAgent string
	Cookie    string
	Proxy     string

	Timeout       time.Duration
	RequestSleep  time.Duration
	RetrySleep    time.Duration
	MaxRetryTimes int64
	Status        int64
}

func defaultFetchSiteConfigs(cfg config.Config) map[string]FetchSiteConfig {
	timeout := time.Duration(cfg.Fetcher.Timeout) * time.Second
	return map[string]FetchSiteConfig{
		FetchSiteCodeJavbus: {
			SiteCode:      FetchSiteCodeJavbus,
			SiteName:      "JavBus",
			BaseURL:       normalizeFetchSiteBaseURL("https://" + strings.TrimSpace(cfg.Fetcher.BusAddress)),
			UserAgent:     cfg.Fetcher.UserAgent,
			Cookie:        cfg.Fetcher.Cookie,
			Proxy:         cfg.Fetcher.Proxy,
			Timeout:       timeout,
			RequestSleep:  legacyFetchSiteRequestSleep,
			RetrySleep:    legacyFetchSiteRetrySleep,
			MaxRetryTimes: legacyFetchSiteRetryTimes,
			Status:        fetchSiteStatusEnabled,
		},
		FetchSiteCodeSukebei: {
			SiteCode:      FetchSiteCodeSukebei,
			SiteName:      "Sukebei",
			BaseURL:       normalizeFetchSiteBaseURL("https://sukebei.nyaa.si"),
			UserAgent:     cfg.Fetcher.UserAgent,
			Cookie:        "",
			Proxy:         cfg.Fetcher.Proxy,
			Timeout:       timeout,
			RequestSleep:  legacyFetchSiteRequestSleep,
			RetrySleep:    legacyFetchSiteRetrySleep,
			MaxRetryTimes: legacyFetchSiteRetryTimes,
			Status:        fetchSiteStatusEnabled,
		},
		FetchSiteCodeSehuatang: {
			SiteCode:      FetchSiteCodeSehuatang,
			SiteName:      "Sehuatang",
			BaseURL:       normalizeFetchSiteBaseURL("https://vzzr.qnc8.net"),
			UserAgent:     cfg.Fetcher.UserAgent,
			Cookie:        "",
			Proxy:         "",
			Timeout:       timeout,
			RequestSleep:  legacyFetchSiteRequestSleep,
			RetrySleep:    legacyFetchSiteRetrySleep,
			MaxRetryTimes: legacyFetchSiteRetryTimes,
			Status:        fetchSiteStatusEnabled,
		},
	}
}

func loadFetchSiteConfigs(ctx context.Context, cfg config.Config, model moviex.TFetchSiteModel, logger *logrus.Logger) map[string]FetchSiteConfig {
	configs := defaultFetchSiteConfigs(cfg)
	if model == nil {
		return configs
	}

	for siteCode, fallback := range configs {
		row, err := model.FindOneBySiteCode(ctx, siteCode)
		if err == nil {
			configs[siteCode] = mergeFetchSiteConfig(fallback, mapFetchSiteRow(row))
			continue
		}
		if logger != nil && err != moviex.ErrNotFound {
			logger.WithError(err).Warnf("load fetch site config failed: %s", siteCode)
			continue
		}
		if logger != nil {
			logger.Warnf("fetch site config missing, fallback to legacy config: %s", siteCode)
		}
	}

	return configs
}

func mapFetchSiteRow(row *moviex.TFetchSite) FetchSiteConfig {
	if row == nil {
		return FetchSiteConfig{}
	}
	return FetchSiteConfig{
		SiteCode:      row.SiteCode,
		SiteName:      row.SiteName,
		BaseURL:       normalizeFetchSiteBaseURL(row.BaseUrl),
		UserAgent:     row.UserAgent,
		Cookie:        row.Cookie,
		Proxy:         row.Proxy,
		Timeout:       time.Duration(row.TimeoutSeconds) * time.Second,
		RequestSleep:  time.Duration(row.RequestSleepMs) * time.Millisecond,
		RetrySleep:    time.Duration(row.RetrySleepMs) * time.Millisecond,
		MaxRetryTimes: row.MaxRetryTimes,
		Status:        row.Status,
	}
}

func mergeFetchSiteConfig(fallback, override FetchSiteConfig) FetchSiteConfig {
	if override.SiteCode == "" {
		override.SiteCode = fallback.SiteCode
	}
	if override.SiteName == "" {
		override.SiteName = fallback.SiteName
	}
	if override.BaseURL == "" {
		override.BaseURL = fallback.BaseURL
	}
	return override
}

func normalizeFetchSiteBaseURL(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}

func (d *Deps) GetFetchSiteConfig(siteCode string) (FetchSiteConfig, bool) {
	if d == nil || d.FetchSites == nil {
		return FetchSiteConfig{}, false
	}
	cfg, ok := d.FetchSites[siteCode]
	return cfg, ok
}

func (d *Deps) GetFetchSiteBaseURL(siteCode string) string {
	cfg, ok := d.GetFetchSiteConfig(siteCode)
	if !ok {
		return ""
	}
	return cfg.BaseURL
}
