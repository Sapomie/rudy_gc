package film

import (
	"rudy_gc/internal/svc"
)

type Service struct {
	deps *svc.Deps
}

func New(deps *svc.Deps) *Service {
	return &Service{deps}
}
