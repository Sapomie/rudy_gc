package personstat

import "rudy_gc/local/api/internal/dep"

type Service struct {
	deps *dep.Dep
}

func NewService(deps *dep.Dep) *Service {
	return &Service{deps: deps}
}
