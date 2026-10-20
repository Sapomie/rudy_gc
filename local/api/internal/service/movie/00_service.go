package movie

import (
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/fetchsite"
	"rudy_gc/local/api/internal/service/personstat"
)

type Service struct {
	deps          *dep.Dep
	fetchSiteSvc  *fetchsite.Service
	personStatSvc *personstat.Service
}

func NewService(d *dep.Dep) *Service {
	return &Service{
		deps:          d,
		fetchSiteSvc:  fetchsite.NewService(d),
		personStatSvc: personstat.NewService(d),
	}
}
