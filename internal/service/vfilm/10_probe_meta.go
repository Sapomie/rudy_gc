package vfilm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

func (s *Service) ProbeFilmMetaByID(ctx context.Context, id int64) (*types.FilmProbeMetaResult, error) {
	film, err := s.filmFindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if film == nil {
		return nil, types.ErrNotFound
	}

	fullPath := buildFilmAbsolutePath(film)
	if strings.TrimSpace(fullPath) == "" {
		return nil, fmt.Errorf("影片路径为空")
	}
	if _, err := os.Stat(fullPath); err != nil {
		return nil, fmt.Errorf("影片文件不存在: %w", err)
	}

	if err := s.scanAndAttachMetadata(film, fullPath); err != nil {
		return nil, err
	}
	film.NeedScanMeta = consts.FilmMetaDataNoNeedScan

	if _, _, err := s.filmUpsert(ctx, film); err != nil {
		return nil, err
	}
	s.movieSvc.InvalidateMovieType(ctx, film.MovieJavId)

	return &types.FilmProbeMetaResult{
		Id:              film.Id,
		Height:          film.Height,
		DurationMinutes: formatFilmDurationMinutes(film.Duration),
		BitRate:         film.BitRate,
		FrameAverage:    formatFilmFrameAverage(film.FrameAverage),
	}, nil
}

func buildFilmAbsolutePath(film *types.Film) string {
	if film == nil {
		return ""
	}
	fileName := strings.TrimSpace(film.FileName)
	if fileName == "" {
		return ""
	}

	fullDir := strings.TrimSpace(film.FullDir)
	if fullDir != "" {
		return filepath.Join(fullDir, fileName)
	}

	rootDir := strings.TrimSpace(film.RootDir)
	if rootDir != "" {
		return filepath.Join(rootDir, fileName)
	}

	return fileName
}
