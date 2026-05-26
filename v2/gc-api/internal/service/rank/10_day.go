package rank

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"rudy-gc-api/internal/consts"
	"rudy-gc-api/internal/types"
)

func (s *Service) Day(ctx context.Context, date string, page, pageSize int64) (*types.RankDayResponse, error) {
	minDay, err := s.repo.FindEarliestRankDayNumber(ctx)
	if err != nil {
		return nil, err
	}
	maxDay, err := s.repo.FindLatestRankDayNumber(ctx)
	if err != nil {
		return nil, err
	}
	if maxDay <= 0 {
		return &types.RankDayResponse{
			Title:    "MovieCard Rank",
			RankDate: "",
			Items:    []*types.MovieCard{},
		}, nil
	}
	if minDay <= 0 {
		minDay = 1
	}

	dayNumber := maxDay
	if date != "" {
		dayNumber = consts.GetRankDayNumber(date)
		if dayNumber < minDay {
			dayNumber = minDay
		}
		if dayNumber > maxDay {
			dayNumber = maxDay
		}
	}
	rankDate := consts.GetDateStringByRankDayNumber(dayNumber)

	ids, total, err := s.repo.ListRankDayMovieIDs(ctx, dayNumber, page, pageSize)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.LoadMovieCardsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	prevDay := max(dayNumber-1, minDay)
	nextDay := min(dayNumber+1, maxDay)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomDay := minDay
	if maxDay > minDay {
		randomDay = rng.Int63n(maxDay-minDay+1) + minDay
	}

	return &types.RankDayResponse{
		Title:        fmt.Sprintf("MovieCard Rank %s", rankDate),
		RankDate:     rankDate,
		Items:        items,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
		PrevDate:     consts.GetDateStringByRankDayNumber(prevDay),
		NextDate:     consts.GetDateStringByRankDayNumber(nextDay),
		RandomDate:   consts.GetDateStringByRankDayNumber(randomDay),
		PrevDisabled: dayNumber <= minDay,
		NextDisabled: dayNumber >= maxDay,
	}, nil
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
