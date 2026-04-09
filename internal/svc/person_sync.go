package svc

import (
	"context"
	"errors"

	"rudy_gc/internal/model/modelx/moviex"
)

func (d *Deps) SyncPersonStatsByIDs(ctx context.Context, ids []int64, now int64) error {
	if d == nil || d.CastModel == nil || d.PersonModel == nil {
		return nil
	}
	filteredIDs := uniquePositiveInt64s(ids)
	if len(filteredIDs) == 0 {
		return nil
	}

	statsMap, err := d.CastModel.AggregatePersonStatsByIDs(ctx, filteredIDs)
	if err != nil {
		return err
	}

	for _, id := range filteredIDs {
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
			movieNumber       int64
			ownedMovieNumber  int64
			ownedWMediaNumber int64
			scTimes           int64
			comeTimes         int64
			lastScTime        int64
			highestRank       int64
			rankTimes         int64
		)
		if stat != nil {
			movieNumber = stat.MovieNumber
			ownedMovieNumber = stat.OwnedMovieNumber
			ownedWMediaNumber = stat.OwnedWMediaNumber
			scTimes = stat.ScTimes
			comeTimes = stat.ComeTimes
			lastScTime = stat.LastScTime
			highestRank = stat.HighestRank
			rankTimes = stat.RankTimes
		}

		if row.MovieNumber == movieNumber &&
			row.OwnedMovieNumber == ownedMovieNumber &&
			row.OwnedWMediaNumber == ownedWMediaNumber &&
			row.ScTimes == scTimes &&
			row.ComeTimes == comeTimes &&
			row.LastScTime == lastScTime &&
			row.HighestRank == highestRank &&
			row.RankTimes == rankTimes {
			continue
		}

		row.MovieNumber = movieNumber
		row.OwnedMovieNumber = ownedMovieNumber
		row.OwnedWMediaNumber = ownedWMediaNumber
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
	if d.CPersonScModel != nil {
		if err := d.CPersonScModel.RebuildByPersonIDs(ctx, filteredIDs, now); err != nil {
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
