package media

import (
	"context"
	"errors"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) syncJAVScMediaSnapshot(ctx context.Context, movieJavID string) error {
	movieJavID = strings.TrimSpace(movieJavID)
	if movieJavID == "" {
		return nil
	}
	birthTime := int64(0)
	mediaRow, err := s.deps.WMediaModel.FindOneByMovieJavId(ctx, movieJavID)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return err
	}
	if mediaRow != nil && mediaRow.IsRemoved != int64(consts.FilmIsRemoved) {
		birthTime = mediaRow.BirthTime
	}
	statRow, err := s.deps.GScStatModel.FindOneByMovieJavId(ctx, movieJavID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil
		}
		return err
	}
	if statRow.MediaBirthTime == birthTime {
		return nil
	}
	statRow.MediaBirthTime = birthTime
	statRow.UpdatedOn = time.Now().Unix()
	return s.deps.GScStatModel.Update(ctx, statRow)
}

func (s *Service) syncJAVScMediaSnapshots(ctx context.Context, movieJavIDs ...string) error {
	seen := make(map[string]struct{}, len(movieJavIDs))
	for _, movieJavID := range movieJavIDs {
		movieJavID = strings.TrimSpace(movieJavID)
		if movieJavID == "" {
			continue
		}
		if _, ok := seen[movieJavID]; ok {
			continue
		}
		seen[movieJavID] = struct{}{}
		if err := s.syncJAVScMediaSnapshot(ctx, movieJavID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) syncFC2MediaSnapshots(ctx context.Context, movieID int64) error {
	if movieID <= 0 {
		return nil
	}
	present := int64(0)
	birthTime := int64(0)
	mediaRow, err := s.deps.FC2MediaModel.FindOneByMovieID(ctx, movieID)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return err
	}
	if mediaRow != nil && mediaRow.IsRemoved != int64(consts.FilmIsRemoved) {
		present = 1
		birthTime = mediaRow.BirthTime
	}
	now := time.Now().Unix()
	if err := s.deps.FC2MovieModel.UpdateMediaSnapshot(ctx, movieID, present, birthTime, now); err != nil {
		return err
	}
	statRow, err := s.deps.FC2ScStatModel.FindOneByMovieID(ctx, movieID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil
		}
		return err
	}
	if statRow.MediaBirthTime == birthTime {
		return nil
	}
	statRow.MediaBirthTime = birthTime
	statRow.UpdatedOn = now
	return s.deps.FC2ScStatModel.Update(ctx, statRow)
}
