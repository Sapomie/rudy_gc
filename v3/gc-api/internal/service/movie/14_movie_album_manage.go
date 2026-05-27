package movie

import (
	"context"
	"errors"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
)

var (
	ErrMovieAlbumNameEmpty   = errors.New("电影相册名称不能为空")
	ErrMovieAlbumNameTooLong = errors.New("电影相册名称过长")
	ErrMovieAlbumNameExists  = errors.New("电影相册名称已存在")
)

type MovieAlbumOption struct {
	ID       int64
	Name     string
	Selected bool
}

func normalizeMovieAlbumName(raw string) string {
	return strings.TrimSpace(raw)
}

func normalizeMovieNeedDownloadAlbumName() string {
	return consts.MovieNeedDownloadAlbumName
}

func normalizeMovieDeleteAlbumName() string {
	return consts.MovieDeleteAlbumName
}

func (s *Service) CreateMovieAlbum(ctx context.Context, albumName string) (*MovieAlbumOption, error) {
	name := normalizeMovieAlbumName(albumName)
	if name == "" {
		return nil, ErrMovieAlbumNameEmpty
	}
	if len([]byte(name)) > 128 {
		return nil, ErrMovieAlbumNameTooLong
	}

	row, err := s.deps.MovieAlbumModel.FindOneByName(ctx, name)
	if err == nil && row != nil {
		return nil, ErrMovieAlbumNameExists
	}
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, err
	}

	now := time.Now().Unix()
	result, err := s.deps.MovieAlbumModel.Insert(ctx, &moviex.CMovieAlbum{
		Name:      name,
		Remark:    "页面新增电影相册",
		CreatedOn: now,
		UpdatedOn: now,
	})
	if err != nil {
		again, againErr := s.deps.MovieAlbumModel.FindOneByName(ctx, name)
		if againErr == nil && again != nil {
			return nil, ErrMovieAlbumNameExists
		}
		return nil, err
	}

	insertID, idErr := result.LastInsertId()
	if idErr == nil && insertID > 0 {
		return &MovieAlbumOption{ID: insertID, Name: name}, nil
	}

	again, againErr := s.deps.MovieAlbumModel.FindOneByName(ctx, name)
	if againErr != nil {
		return nil, againErr
	}
	return &MovieAlbumOption{ID: again.Id, Name: strings.TrimSpace(again.Name)}, nil
}

func (s *Service) ListMovieAlbumOptions(ctx context.Context) ([]*MovieAlbumOption, error) {
	rows, err := s.deps.MovieAlbumModel.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*MovieAlbumOption, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		name := strings.TrimSpace(row.Name)
		if name == "" {
			continue
		}
		out = append(out, &MovieAlbumOption{ID: row.Id, Name: name})
	}
	return out, nil
}

func (s *Service) ensureMovieNeedDownloadAlbum(ctx context.Context) (int64, error) {
	return s.getMovieAlbumIDByName(ctx, normalizeMovieNeedDownloadAlbumName(), true)
}

func (s *Service) ensureMovieDeleteAlbum(ctx context.Context) (int64, error) {
	return s.getMovieAlbumIDByName(ctx, normalizeMovieDeleteAlbumName(), true)
}
