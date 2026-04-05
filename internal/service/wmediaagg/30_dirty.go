package wmediaagg

import (
	"context"
	"errors"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) MarkMovieDirty(ctx context.Context, javID string) error {
	row, err := s.deps.WMediaModel.FindOneByMovieJavId(ctx, javID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil
		}
		return err
	}
	return s.MarkMediaRowsDirty(ctx, row)
}

func (s *Service) MarkMoviesDirty(ctx context.Context, javIDs ...string) error {
	for _, javID := range javIDs {
		if err := s.MarkMovieDirty(ctx, javID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) MarkMediaRowsDirty(ctx context.Context, rows ...*moviex.WMedia) error {
	now := time.Now().Unix()
	for _, row := range rows {
		if row == nil || row.BirthTime <= 0 {
			continue
		}
		bucketDay := bucketDayFromBirthTime(row.BirthTime)
		if bucketDay <= 0 {
			continue
		}
		dayScope := scopeFromBucketDay(bucketDay)
		if err := s.deps.WMediaAggDirtyModel.TouchDay(ctx, bucketDay, dayScope.Key, now); err != nil {
			return err
		}
	}
	return nil
}
