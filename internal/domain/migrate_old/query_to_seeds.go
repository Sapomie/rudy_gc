package migrate

import (
	"context"
	"rudy_gc/internal/types"

	"github.com/pkg/errors"
)

func (s *Service) MigrateSeeds() error {
	ctx := context.Background()
	queries, err := s.xModel.QueryModel.FindAll(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, q := range queries {
		seed := types.Seed{
			Name:          q.Name,
			Active:        q.Active,
			SearchType:    q.SearchType,
			NameType:      q.NameType,
			PageNow:       q.PageNow,
			Offset:        q.Offset,
			StartPage:     q.StartPage,
			EndPage:       q.EndPage,
			LastQueryTime: q.LastQueryTime,
		}
		_, err := s.deps.SeedRepo.Upsert(ctx, &seed)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
