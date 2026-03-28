package svc

import (
	"context"
	"errors"

	"rudy_gc/data/modelx/moviex"
)

func (d *Deps) SyncPersonStatsByIDs(ctx context.Context, ids []int64, now int64) error {
	if d == nil || d.CastModel == nil || d.PersonModel == nil {
		return nil
	}

	statsMap, err := d.CastModel.AggregatePersonStatsByIDs(ctx, ids)
	if err != nil {
		return err
	}

	for _, id := range uniquePositiveInt64s(ids) {
		row, err := d.PersonModel.FindOne(ctx, id)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				continue
			}
			return err
		}
		if row == nil {
			continue
		}

		stat := statsMap[id]
		var (
			movieNumber      int64
			ownedMovieNumber int64
			scTimes          int64
			comeTimes        int64
			lastScTime       int64
			highestRank      int64
			rankTimes        int64
		)
		if stat != nil {
			movieNumber = stat.MovieNumber
			ownedMovieNumber = stat.OwnedMovieNumber
			scTimes = stat.ScTimes
			comeTimes = stat.ComeTimes
			lastScTime = stat.LastScTime
			highestRank = stat.HighestRank
			rankTimes = stat.RankTimes
		}

		if row.MovieNumber == movieNumber &&
			row.OwnedMovieNumber == ownedMovieNumber &&
			row.ScTimes == scTimes &&
			row.ComeTimes == comeTimes &&
			row.LastScTime == lastScTime &&
			row.HighestRank == highestRank &&
			row.RankTimes == rankTimes {
			continue
		}

		row.MovieNumber = movieNumber
		row.OwnedMovieNumber = ownedMovieNumber
		row.ScTimes = scTimes
		row.ComeTimes = comeTimes
		row.LastScTime = lastScTime
		row.HighestRank = highestRank
		row.RankTimes = rankTimes
		row.UpdatedOn = now
		if err := d.PersonModel.Update(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func uniquePositiveInt64s(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
