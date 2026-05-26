package media

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/movie"
)

type Service struct {
	deps     *dep.Dep
	movieSvc *movie.Service
}

func NewService(d *dep.Dep) *Service {
	return &Service{
		deps:     d,
		movieSvc: movie.NewService(d),
	}
}

func (s *Service) mediaRoots() []string {
	raw := s.deps.Config.Media.RootDirs
	if len(raw) == 0 {
		return nil
	}

	out := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, root := range raw {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		cleaned := filepath.Clean(root)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		out = append(out, cleaned)
	}
	return out
}

func (s *Service) rebuildMediaAggs(ctx context.Context, flowKey string, rows ...*moviex.WMedia) {
	_ = ctx
	if len(rows) == 0 {
		return
	}
	javIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		javID := strings.TrimSpace(row.MovieJavId)
		if javID == "" {
			continue
		}
		javIDs = append(javIDs, javID)
	}
	s.movieSvc.EnqueueAggRebuildByMovieJavIDs(flowKey, javIDs...)
}

func (s *Service) syncPersonStatsByMovieJavIDs(ctx context.Context, now int64, javIDs ...string) error {
	if s == nil || s.deps == nil || s.deps.MovieCastModel == nil || s.deps.CastModel == nil {
		return nil
	}
	if now <= 0 {
		now = time.Now().Unix()
	}

	seenMovieIDs := make(map[string]struct{}, len(javIDs))
	castIDs := make([]int64, 0, len(javIDs)*2)
	seenCastIDs := make(map[int64]struct{}, len(javIDs)*2)
	for _, javID := range javIDs {
		javID = strings.TrimSpace(javID)
		if javID == "" {
			continue
		}
		if _, ok := seenMovieIDs[javID]; ok {
			continue
		}
		seenMovieIDs[javID] = struct{}{}

		ids, err := s.deps.MovieCastModel.ListCastIDsByMovieJavId(ctx, javID)
		if err != nil {
			return err
		}
		for _, castID := range ids {
			if castID <= 0 {
				continue
			}
			if _, ok := seenCastIDs[castID]; ok {
				continue
			}
			seenCastIDs[castID] = struct{}{}
			castIDs = append(castIDs, castID)
		}
	}
	if len(castIDs) == 0 {
		return nil
	}

	personIDs := make([]int64, 0, len(castIDs))
	seenPersonIDs := make(map[int64]struct{}, len(castIDs))
	for _, castID := range castIDs {
		castRow, err := s.deps.CastModel.FindOne(ctx, castID)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				continue
			}
			return err
		}
		if castRow == nil || castRow.PersonId <= 0 {
			continue
		}
		if _, ok := seenPersonIDs[castRow.PersonId]; ok {
			continue
		}
		seenPersonIDs[castRow.PersonId] = struct{}{}
		personIDs = append(personIDs, castRow.PersonId)
	}
	if len(personIDs) == 0 {
		return nil
	}
	return s.deps.SyncPersonStatsByIDs(ctx, personIDs, now)
}
