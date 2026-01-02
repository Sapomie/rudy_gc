package movie

import (
	"context"

	"rudy_gc/internal/types"
)

const (
	rankDayDefaultPageSize = 18
	rankDayMaxPageSize     = 20000
)

func (s *MovieService) FindLatestRankDayNumber(ctx context.Context) (int64, error) {
	return s.deps.RankRepo.FindLatestDayNumber(ctx)
}

func (s *MovieService) FindEarliestRankDayNumber(ctx context.Context) (int64, error) {
	return s.deps.RankRepo.FindEarliestDayNumber(ctx)
}

func (s *MovieService) ListMovieTypesByRankDay(ctx context.Context, dayNumber, page, pageSize int64) ([]*types.MovieType, int64, error) {
	ranks, err := s.deps.RankRepo.ListByDayNumber(ctx, dayNumber)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(ranks))
	if total == 0 {
		return []*types.MovieType{}, 0, nil
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > rankDayMaxPageSize {
		pageSize = rankDayDefaultPageSize
	}

	start := (page - 1) * pageSize
	if start >= total {
		return []*types.MovieType{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	rankSlice := ranks[start:end]
	out := make([]*types.MovieType, 0, len(rankSlice))
	for _, rk := range rankSlice {
		if rk == nil {
			continue
		}
		mt, err := s.GetMovieType(ctx, rk.MovieJavId)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, mt)
	}
	return out, total, nil
}
