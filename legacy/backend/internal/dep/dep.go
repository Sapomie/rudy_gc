package dep

import (
	"rudy_gc/internal/config"
	"rudy_gc/internal/svc"
)

type Dep = svc.Deps

func New(c config.Config) (*Dep, error) {
	return svc.NewDeps(c)
}
