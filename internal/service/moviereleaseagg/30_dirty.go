package moviereleaseagg

import (
	"context"
	"errors"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) MarkMovieDirty(ctx context.Context, javID string) error {
	row, err := s.deps.MovieModel.FindOneByJavId(ctx, javID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil
		}
		return err
	}
	return s.MarkMovieRowsDirty(ctx, row)
}

func (s *Service) MarkMoviesDirty(ctx context.Context, javIDs ...string) error {
	for _, javID := range javIDs {
		if err := s.MarkMovieDirty(ctx, javID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) MarkReleaseTimesDirty(ctx context.Context, releaseTimes ...int64) error {
	now := time.Now().Unix()
	for _, releaseTime := range releaseTimes {
		bucketMonth := bucketMonthFromReleaseTime(releaseTime)
		if bucketMonth <= 0 {
			continue
		}
		sc := scopeFromBucketMonth(bucketMonth)
		if err := s.deps.MovieReleaseAggDirtyModel.TouchMonth(ctx, bucketMonth, sc.Key, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) MarkMovieRowsDirty(ctx context.Context, rows ...*moviex.AMovie) error {
	releaseTimes := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.ReleasingDate <= 0 {
			continue
		}
		releaseTimes = append(releaseTimes, row.ReleasingDate)
	}
	return s.MarkReleaseTimesDirty(ctx, releaseTimes...)
}
