package handler

import (
	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/svc"
)

func newFetchSitePageService(deps *svc.Deps) *fetchsite.Service {
	return fetchsite.NewService(deps)
}
