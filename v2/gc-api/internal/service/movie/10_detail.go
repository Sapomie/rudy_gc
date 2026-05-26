package movie

import (
	"context"

	"rudy-gc-api/internal/types"
)

func (s *Service) Detail(ctx context.Context, movieName string) (*types.MovieDetailResponse, error) {
	card, err := s.repo.FindMovieCardByName(ctx, movieName)
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

	return &types.MovieDetailResponse{
		Movie:         card,
		Media:         media,
		RankInfos:     ranks,
		ScEvents:      scEvents,
		JavbusFetch:   javbusFetch,
		JavbusMagnets: javbusMagnets,
		SukebeiFetch:  sukebeiFetch,
		SukebeiRows:   sukebeiRows,
		SehuatangRows: sehuatangRows,
	}, nil
}
