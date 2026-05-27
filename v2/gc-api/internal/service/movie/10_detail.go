package movie

import (
	"context"
	"time"

	"rudy-gc-api/internal/model/modelx"
	"rudy-gc-api/internal/types"
)

const movieCardCacheTTL = 7 * 24 * time.Hour

func (s *Service) Detail(ctx context.Context, movieName string) (*types.MovieDetailResponse, error) {
	card, err := s.loadMovieCardByName(ctx, movieName)
	if err != nil {
		return nil, err
	}

	media, err := s.repo.LoadMovieMedia(ctx, card.MovieJavID)
	if err != nil && !isNotFound(err) {
		return nil, err
	}
	ranks, err := s.repo.LoadMovieRanks(ctx, card.MovieJavID)
	if err != nil {
		return nil, err
	}
	scEvents, err := s.repo.LoadMovieScEvents(ctx, card.MovieName)
	if err != nil {
		return nil, err
	}
	javbusFetch, err := s.repo.LoadJavbusFetch(ctx, card.MovieJavID)
	if err != nil {
		return nil, err
	}
	javbusMagnets, err := s.repo.TJavbusMagnetRowList(ctx, card.MovieJavID)
	if err != nil {
		return nil, err
	}
	sukebeiFetch, err := s.repo.LoadSukebeiFetch(ctx, card.MovieJavID)
	if err != nil {
		return nil, err
	}
	sukebeiRows, err := s.repo.TSukebeiTorrentRowList(ctx, card.MovieJavID)
	if err != nil {
		return nil, err
	}
	sehuatangRows, err := s.repo.TSehuatangMagnetRowList(ctx, card.MovieJavID, card.MovieName)
	if err != nil {
		return nil, err
	}
	albums, err := s.repo.LoadMovieAlbums(ctx, card.MovieJavID)
	if err != nil {
		return nil, err
	}
	markedDelete, err := s.repo.IsMovieMarkedDelete(ctx, card.MovieJavID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.ApplyFetchResourceStatus(ctx, card.MovieJavID, javbusMagnets, sukebeiRows, sehuatangRows); err != nil {
		return nil, err
	}

	return &types.MovieDetailResponse{
		Movie:         card,
		Media:         media,
		MovieAlbums:   albums,
		MarkedDelete:  markedDelete,
		RankInfos:     ranks,
		ScEvents:      scEvents,
		JavbusFetch:   javbusFetch,
		JavbusMagnets: javbusMagnets,
		SukebeiFetch:  sukebeiFetch,
		SukebeiRows:   sukebeiRows,
		SehuatangRows: sehuatangRows,
	}, nil
}

func (s *Service) loadMovieCardByName(ctx context.Context, movieName string) (*types.MovieCard, error) {
	javID, err := s.repo.FindMovieJavIDByName(ctx, movieName)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		card, err := s.cache.GetMovieCard(ctx, javID)
		if err != nil {
			return nil, err
		}
		if card != nil {
			return card, nil
		}
	}
	cards, err := s.repo.LoadMovieCardsByIDs(ctx, []string{javID})
	if err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return nil, modelx.ErrNotFound
	}
	card := cards[0]
	if s.cache != nil {
		if err := s.cache.SetMovieCard(ctx, card.MovieJavID, card, movieCardCacheTTL); err != nil {
			return nil, err
		}
	}
	return card, nil
}
