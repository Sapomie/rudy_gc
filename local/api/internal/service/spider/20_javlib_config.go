package spider

import (
	"context"
	"errors"

	"rudy_gc/local/api/internal/service/fetchsite"
	"rudy_gc/local/api/internal/service/spider/fetcher"
)

func (l *Service) javlibURL(parts ...string) (string, error) {
	site := fetchsite.NewService(l.deps.Dep)
	cfg, err := site.GetConfig(fetchsite.FetchSiteCodeJavlib)
	if err != nil {
		return "", err
	}
	if cfg.Status != 1 {
		return "", errors.New("Javlib 站点已停用")
	}
	return site.BuildURL(fetchsite.FetchSiteCodeJavlib, append([]string{"cn"}, parts...)...)
}

func (l *Service) fetchJavlib(ctx context.Context, rawURL string) (*fetcher.Response, error) {
	site := fetchsite.NewService(l.deps.Dep)
	cfg, err := site.GetConfig(fetchsite.FetchSiteCodeJavlib)
	if err != nil {
		return nil, err
	}
	if cfg.Status != 1 {
		return nil, errors.New("Javlib 站点已停用")
	}
	return site.Get(ctx, fetchsite.FetchSiteCodeJavlib, rawURL, fetchsite.RequestOptions{})
}
