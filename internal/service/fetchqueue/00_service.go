package fetchqueue

import (
	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/fetchjavbus"
	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/service/fetchsukebei"
)

type Service struct {
	deps       *dep.Dep
	siteSvc    *fetchsite.Service
	javbusSvc  *fetchjavbus.Service
	sukebeiSvc *fetchsukebei.Service
}

func NewService(d *dep.Dep) *Service {
	return &Service{
		deps:       d,
		siteSvc:    fetchsite.NewService(d),
		javbusSvc:  fetchjavbus.NewService(d),
		sukebeiSvc: fetchsukebei.NewService(d),
	}
}
