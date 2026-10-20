package fetchsite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	fetchsitetypes "rudy_gc/local/api/internal/types/fetchsite"
)

type javlibConfigModel interface {
	FindOneBySiteCode(context.Context, string) (*moviex.TFetchSite, error)
	InsertIfMissing(context.Context, *moviex.TFetchSite) (sql.Result, error)
}

// EnsureJavlibConfig 将旧 YAML 配置首次写入站点表，已有数据库配置优先。
func EnsureJavlibConfig(ctx context.Context, cfg config.Config, model javlibConfigModel) error {
	if model == nil {
		return errors.New("Javlib 站点配置模型未初始化")
	}
	if _, err := model.FindOneBySiteCode(ctx, fetchsitetypes.CodeJavlib); err == nil {
		return nil
	} else if !errors.Is(err, moviex.ErrNotFound) {
		return fmt.Errorf("读取 Javlib 站点配置失败：%w", err)
	}

	site := defaultConfigs(cfg)[fetchsitetypes.CodeJavlib]
	parsed, err := url.Parse(site.BaseURL)
	if err != nil || parsed.Host == "" || parsed.Scheme != "https" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("初始化 Javlib 配置需要有效的 Fetcher.JavAddress 域名")
	}
	if site.Timeout <= 0 {
		return errors.New("初始化 Javlib 配置需要有效的 Fetcher.Timeout")
	}
	now := time.Now().Unix()
	_, err = model.InsertIfMissing(ctx, &moviex.TFetchSite{
		SiteCode: site.SiteCode, SiteName: site.SiteName, BaseUrl: site.BaseURL,
		UserAgent: site.UserAgent, Cookie: site.Cookie, Proxy: site.Proxy,
		TimeoutSeconds: int64(site.Timeout / time.Second),
		RequestSleepMs: int64(site.RequestSleep / time.Millisecond),
		RetrySleepMs:   int64(site.RetrySleep / time.Millisecond),
		MaxRetryTimes:  site.MaxRetryTimes, Status: site.Status,
		CreatedOn: now, UpdatedOn: now,
	})
	if err != nil {
		return fmt.Errorf("初始化 Javlib 站点配置失败：%w", err)
	}
	return nil
}
