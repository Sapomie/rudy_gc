package fetchsite

import (
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/spider/fetcher"
	fetchsitetypes "rudy_gc/local/api/internal/types/fetchsite"
)

type Service struct {
	deps *dep.Dep
}

func NewService(d *dep.Dep) *Service {
	return &Service{deps: d}
}

const (
	FetchSiteCodeJavlib     = fetchsitetypes.CodeJavlib
	FetchSiteCodeJavbus     = fetchsitetypes.CodeJavbus
	FetchSiteCodeSukebei    = fetchsitetypes.CodeSukebei
	FetchSiteCodeSehuatang  = fetchsitetypes.CodeSehuatang
	FetchSiteCodeBaiduFanyi = fetchsitetypes.CodeBaiduFanyi

	FetchStatusPending = int64(1)
	FetchStatusRunning = int64(2)
	FetchStatusSuccess = int64(3)
	FetchStatusFailed  = int64(4)
)

type SiteConfig = fetchsitetypes.Config
type RequestOptions = fetcher.RequestOptions
type Response = fetcher.Response
