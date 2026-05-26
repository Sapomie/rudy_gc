package page

import (
	"rudy-gc-api/internal/dep"
	"rudy-gc-api/internal/model/modelx"
)

type Service struct {
	repo *modelx.PageRepo
}

func New(d *dep.Dep) *Service {
	return &Service{
		repo: modelx.NewPageRepo(d.Conn),
	}
}
