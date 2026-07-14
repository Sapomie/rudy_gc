package wkv

import "rudy_gc/internal/svc"

const (
	ItemKeyHRKTime = "hrk_time"
)

type Service struct {
	deps *svc.Deps
}

func NewService(deps *svc.Deps) *Service {
	return &Service{deps: deps}
}
