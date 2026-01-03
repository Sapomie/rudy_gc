package logic

import (
	"context"
	"fmt"
	"rudy_gc/internal/consts"
	"time"
)

type affectedMovieNumbers struct {
	castIDs     map[int64]struct{}
	genreIDs    map[int64]struct{}
	directorIDs map[int64]struct{}
	labelIDs    map[int64]struct{}
	makerIDs    map[int64]struct{}
	prefixIDs   map[int64]struct{}
}

func newAffectedMovieNumbers() *affectedMovieNumbers {
	return &affectedMovieNumbers{
		castIDs:     map[int64]struct{}{},
		genreIDs:    map[int64]struct{}{},
		directorIDs: map[int64]struct{}{},
		labelIDs:    map[int64]struct{}{},
		makerIDs:    map[int64]struct{}{},
		prefixIDs:   map[int64]struct{}{},
	}
}

func (a *affectedMovieNumbers) addFromResponse(resp *saveParsedMovieResponse) {
	if resp == nil {
		return
	}
	for _, id := range resp.castIDs {
		if id > 0 {
			a.castIDs[id] = struct{}{}
		}
	}
	for _, id := range resp.genreIDs {
		if id > 0 {
			a.genreIDs[id] = struct{}{}
		}
	}
	if resp.directorID > 0 {
		a.directorIDs[resp.directorID] = struct{}{}
	}
	if resp.labelID > 0 {
		a.labelIDs[resp.labelID] = struct{}{}
	}
	if resp.makerID > 0 {
		a.makerIDs[resp.makerID] = struct{}{}
	}
	if resp.prefixID > 0 {
		a.prefixIDs[resp.prefixID] = struct{}{}
	}
}

func (a *affectedMovieNumbers) addIDs(
	castIDs []int64,
	genreIDs []int64,
	directorIDs []int64,
	labelIDs []int64,
	makerIDs []int64,
	prefixIDs []int64,
) {
	for _, id := range castIDs {
		if id > 0 {
			a.castIDs[id] = struct{}{}
		}
	}
	for _, id := range genreIDs {
		if id > 0 {
			a.genreIDs[id] = struct{}{}
		}
	}
	for _, id := range directorIDs {
		if id > 0 {
			a.directorIDs[id] = struct{}{}
		}
	}
	for _, id := range labelIDs {
		if id > 0 {
			a.labelIDs[id] = struct{}{}
		}
	}
	for _, id := range makerIDs {
		if id > 0 {
			a.makerIDs[id] = struct{}{}
		}
	}
	for _, id := range prefixIDs {
		if id > 0 {
			a.prefixIDs[id] = struct{}{}
		}
	}
}

func (l *CrawlLogic) updateMovieNumbers(ctx context.Context, affected *affectedMovieNumbers) error {
	if affected == nil {
		return nil
	}
	now := time.Now().Unix()

	return l.runParallel(
		ctx,
		func(ctx context.Context) error { return l.updateCastMovieNumbers(ctx, affected.castIDs, now) },
		func(ctx context.Context) error { return l.updateGenreMovieNumbers(ctx, affected.genreIDs, now) },
		func(ctx context.Context) error { return l.updateDirectorMovieNumbers(ctx, affected.directorIDs, now) },
		func(ctx context.Context) error { return l.updateLabelMovieNumbers(ctx, affected.labelIDs, now) },
		func(ctx context.Context) error { return l.updateMakerMovieNumbers(ctx, affected.makerIDs, now) },
		func(ctx context.Context) error { return l.updatePrefixMovieNumbers(ctx, affected.prefixIDs, now) },
	)
}

func (l *CrawlLogic) UpdateAllMovieNumbers(ctx context.Context) error {
	castIDs, err := l.deps.CastRepo.ListAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("list am_cast ids: %w", err)
	}
	genreIDs, err := l.deps.GenreRepo.ListAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("list am_genre ids: %w", err)
	}
	directorIDs, err := l.deps.DirectorRepo.ListAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("list am_director ids: %w", err)
	}
	labelIDs, err := l.deps.LabelRepo.ListAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("list am_label ids: %w", err)
	}
	makerIDs, err := l.deps.MakerRepo.ListAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("list am_maker ids: %w", err)
	}
	prefixIDs, err := l.deps.PrefixRepo.ListAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("list am_prefix ids: %w", err)
	}

	affected := newAffectedMovieNumbers()
	affected.addIDs(castIDs, genreIDs, directorIDs, labelIDs, makerIDs, prefixIDs)

	return l.updateMovieNumbers(ctx, affected)
}

func (l *CrawlLogic) updateCastMovieNumbers(ctx context.Context, ids map[int64]struct{}, now int64) error {
	for id := range ids {
		if err := l.deps.CastRepo.UpdateMovieNumbersByID(ctx, id, consts.FilmIsNotRemoved, now); err != nil {
			return fmt.Errorf("update am_cast movie numbers (id=%d): %w", id, err)
		}
	}
	return nil
}

func (l *CrawlLogic) updateGenreMovieNumbers(ctx context.Context, ids map[int64]struct{}, now int64) error {
	for id := range ids {
		if err := l.deps.GenreRepo.UpdateMovieNumbersByID(ctx, id, consts.FilmIsNotRemoved, now); err != nil {
			return fmt.Errorf("update am_genre movie numbers (id=%d): %w", id, err)
		}
	}
	return nil
}

func (l *CrawlLogic) updateDirectorMovieNumbers(ctx context.Context, ids map[int64]struct{}, now int64) error {
	for id := range ids {
		if err := l.deps.DirectorRepo.UpdateMovieNumbersByID(ctx, id, consts.FilmIsNotRemoved, now); err != nil {
			return fmt.Errorf("update am_director movie numbers (id=%d): %w", id, err)
		}
	}
	return nil
}

func (l *CrawlLogic) updateLabelMovieNumbers(ctx context.Context, ids map[int64]struct{}, now int64) error {
	for id := range ids {
		if err := l.deps.LabelRepo.UpdateMovieNumbersByID(ctx, id, consts.FilmIsNotRemoved, now); err != nil {
			return fmt.Errorf("update am_label movie numbers (id=%d): %w", id, err)
		}
	}
	return nil
}

func (l *CrawlLogic) updateMakerMovieNumbers(ctx context.Context, ids map[int64]struct{}, now int64) error {
	for id := range ids {
		if err := l.deps.MakerRepo.UpdateMovieNumbersByID(ctx, id, consts.FilmIsNotRemoved, now); err != nil {
			return fmt.Errorf("update am_maker movie numbers (id=%d): %w", id, err)
		}
	}
	return nil
}

func (l *CrawlLogic) updatePrefixMovieNumbers(ctx context.Context, ids map[int64]struct{}, now int64) error {
	for id := range ids {
		if err := l.deps.PrefixRepo.UpdateMovieNumbersByID(ctx, id, consts.FilmIsNotRemoved, now); err != nil {
			return fmt.Errorf("update am_prefix movie numbers (id=%d): %w", id, err)
		}
	}
	return nil
}
