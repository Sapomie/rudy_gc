package fetchsite

import (
	"context"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	fetchsitetypes "rudy_gc/local/api/internal/types/fetchsite"
)

const (
	statusEnabled    int64 = 1
	legacyRetryTimes int64 = 45
)

var (
	legacyRequestSleep = 3 * time.Second
	legacyRetrySleep   = 3 * time.Second
)

func Load(ctx context.Context, cfg config.Config, model moviex.TFetchSiteModel, logger *logrus.Logger) map[string]fetchsitetypes.Config {
	configs := defaultConfigs(cfg)
	if model == nil {
		return configs
	}

	for siteCode, fallback := range configs {
		row, err := model.FindOneBySiteCode(ctx, siteCode)
		if err == nil {
			configs[siteCode] = mergeConfig(fallback, FromRow(row))
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

func FromRow(row *moviex.TFetchSite) fetchsitetypes.Config {
	if row == nil {
		return fetchsitetypes.Config{}
	}
	return fetchsitetypes.Config{
		SiteCode:      row.SiteCode,
		SiteName:      row.SiteName,
		BaseURL:       NormalizeBaseURL(row.BaseUrl),
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

func NormalizeBaseURL(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}

func defaultConfigs(cfg config.Config) map[string]fetchsitetypes.Config {
	timeout := time.Duration(cfg.Fetcher.Timeout) * time.Second
	return map[string]fetchsitetypes.Config{
		fetchsitetypes.CodeJavlib: {
			SiteCode: fetchsitetypes.CodeJavlib, SiteName: "Javlib",
			BaseURL:   NormalizeBaseURL("https://" + strings.TrimSpace(cfg.Fetcher.JavAddress)),
			UserAgent: cfg.Fetcher.UserAgent, Cookie: cfg.Fetcher.Cookie, Timeout: timeout,
			RequestSleep: legacyRequestSleep, RetrySleep: 5 * time.Second,
			MaxRetryTimes: 20, Status: statusEnabled, // 沿用 Javlib 现有重试次数和基础间隔
		},
		fetchsitetypes.CodeJavbus: {
			SiteCode: fetchsitetypes.CodeJavbus, SiteName: "JavBus",
			BaseURL:   NormalizeBaseURL("https://" + strings.TrimSpace(cfg.Fetcher.BusAddress)),
			UserAgent: cfg.Fetcher.UserAgent, Timeout: timeout,
			RequestSleep: legacyRequestSleep, RetrySleep: legacyRetrySleep,
			MaxRetryTimes: legacyRetryTimes, Status: statusEnabled,
		},
		fetchsitetypes.CodeSukebei: {
			SiteCode: fetchsitetypes.CodeSukebei, SiteName: "Sukebei",
			BaseURL:   NormalizeBaseURL("https://sukebei.nyaa.si"),
			UserAgent: cfg.Fetcher.UserAgent, Timeout: timeout,
			RequestSleep: legacyRequestSleep, RetrySleep: legacyRetrySleep,
			MaxRetryTimes: legacyRetryTimes, Status: statusEnabled,
		},
		fetchsitetypes.CodeSehuatang: {
			SiteCode: fetchsitetypes.CodeSehuatang, SiteName: "Sehuatang",
			BaseURL:   NormalizeBaseURL("https://vzzr.qnc8.net"),
			UserAgent: cfg.Fetcher.UserAgent, Timeout: timeout,
			RequestSleep: legacyRequestSleep, RetrySleep: legacyRetrySleep,
			MaxRetryTimes: legacyRetryTimes, Status: statusEnabled,
		},
		fetchsitetypes.CodeBaiduFanyi: {
			SiteCode: fetchsitetypes.CodeBaiduFanyi, SiteName: "BaiduFanyi",
			BaseURL:   NormalizeBaseURL("https://api.fanyi.baidu.com"),
			UserAgent: cfg.Fetcher.UserAgent, Timeout: timeout,
			RequestSleep: legacyRequestSleep, RetrySleep: legacyRetrySleep,
			MaxRetryTimes: legacyRetryTimes, Status: statusEnabled,
		},
	}
}

func mergeConfig(fallback, override fetchsitetypes.Config) fetchsitetypes.Config {
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
