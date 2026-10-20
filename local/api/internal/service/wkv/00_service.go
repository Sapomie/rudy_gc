package wkv

import "rudy_gc/local/api/internal/dep"

const (
	ItemKeyHRKTime = "hrk_time"
)

type Service struct {
	deps *dep.Dep
}

func NewService(deps *dep.Dep) *Service {
	return &Service{deps: deps}
}
