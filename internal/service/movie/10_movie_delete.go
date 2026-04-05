package movie

import (
	"context"
	"strings"

	"rudy_gc/internal/service/moviereleaseagg"
)

func (s *Service) DeleteMovieByJavID(ctx context.Context, javID, fallbackName, deleteSource string) error {
	javID = strings.TrimSpace(javID)
	if javID == "" {
		return nil
	}

	deleteSource = strings.TrimSpace(deleteSource)
	if deleteSource == "" {
		deleteSource = "manual"
	}

	delCtx, err := s.loadMovieDeleteContext(ctx, javID, strings.TrimSpace(fallbackName))
	if err != nil {
		return err
	}

	if err := s.upsertDeletedMovie(ctx, delCtx, deleteSource); err != nil {
		return err
	}
	if err := s.deleteMovieRows(ctx, delCtx); err != nil {
		return err
	}
	if err := s.rebuildMovieDeleteAffectedStats(ctx, delCtx); err != nil {
		return err
	}
	if delCtx.Movie != nil && delCtx.Movie.ReleasingDate > 0 {
		if err := moviereleaseagg.NewService(s.deps).MarkReleaseTimesDirty(ctx, delCtx.Movie.ReleasingDate); err != nil {
			return err
		}
	}

	s.InvalidateMovieType(ctx, javID)
	return s.rebuildAggsAfterFlow(ctx, "movie_delete")
}
