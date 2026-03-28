package fetchsite

import (
	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/spider/fetcher"
	"rudy_gc/internal/svc"
)

type Service struct {
	deps *dep.Dep
}

func NewService(d *dep.Dep) *Service {
	return &Service{deps: d}
}

const (
	FetchSiteCodeJavbus  = svc.FetchSiteCodeJavbus
	FetchSiteCodeSukebei = svc.FetchSiteCodeSukebei

	FetchStatusPending = int64(1)
	FetchStatusRunning = int64(2)
	FetchStatusSuccess = int64(3)
	FetchStatusFailed  = int64(4)
)

type SiteConfig = svc.FetchSiteConfig
type RequestOptions = fetcher.RequestOptions
type Response = fetcher.Response
