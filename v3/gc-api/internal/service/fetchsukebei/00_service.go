package fetchsukebei

import (
	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/fetchsite"
)

type Service struct {
	deps    *dep.Dep
	siteSvc *fetchsite.Service
}

func NewService(d *dep.Dep) *Service {
	return &Service{
		deps:    d,
		siteSvc: fetchsite.NewService(d),
	}
}
