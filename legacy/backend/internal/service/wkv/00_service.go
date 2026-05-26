package wkv

import "rudy_gc/internal/svc"

const (
	ItemKeySHTTime = "sht_time"
	ItemKeyHRKTime = "hrk_time"
)

type Service struct {
	deps *svc.Deps
}

func NewService(deps *svc.Deps) *Service {
	return &Service{deps: deps}
}
