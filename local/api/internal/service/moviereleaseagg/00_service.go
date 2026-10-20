package moviereleaseagg

import "rudy_gc/local/api/internal/dep"

type Service struct {
	deps *dep.Dep
}

func NewService(d *dep.Dep) *Service {
	return &Service{deps: d}
}
