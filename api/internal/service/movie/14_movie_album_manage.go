package movie

import (
	"context"
	"errors"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/repository/moviex"
)

var (
	ErrMovieAlbumNameEmpty     = errors.New("电影相册名称不能为空")
	ErrMovieAlbumNameTooLong   = errors.New("电影相册名称过长")
	ErrMovieAlbumNameExists    = errors.New("电影相册名称已存在")
	ErrMovieAlbumNotFound      = errors.New("电影相册不存在")
	ErrMovieAlbumUpdateBlocked = errors.New("业务电影相册不允许编辑")
	ErrMovieAlbumDeleteBlocked = errors.New("业务电影相册不允许删除")
	ErrMovieAlbumNotEmpty      = errors.New("电影相册中还有数据，不能删除")
)

type MovieAlbumOption struct {
	ID        int64
	Name      string
	Selected  bool
	Protected bool
}

type UpdateMovieAlbumResult struct {
	ID   int64
	Name string
}

type DeleteMovieAlbumResult struct {
	ID   int64
	Name string
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

func IsProtectedMovieAlbumName(albumName string) bool {
	switch normalizeMovieAlbumName(albumName) {
	case consts.MovieNeedDownloadAlbumName, consts.MovieDeleteAlbumName, "v3", "vx":
		return true
	default:
		return false
	}
}

func validateMovieAlbumDeletion(itemCount int64) error {
	if itemCount > 0 {
		return ErrMovieAlbumNotEmpty
	}
	return nil
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
		out = append(out, &MovieAlbumOption{ID: row.Id, Name: name, Protected: IsProtectedMovieAlbumName(name)})
	}
	return out, nil
}

func (s *Service) UpdateMovieAlbum(ctx context.Context, albumID int64, albumName string) (*UpdateMovieAlbumResult, error) {
	if albumID <= 0 {
		return nil, ErrMovieAlbumInvalidID
	}

	name := normalizeMovieAlbumName(albumName)
	if name == "" {
		return nil, ErrMovieAlbumNameEmpty
	}
	if len([]byte(name)) > 128 {
		return nil, ErrMovieAlbumNameTooLong
	}

	row, err := s.deps.MovieAlbumModel.FindOne(ctx, albumID)
	if errors.Is(err, moviex.ErrNotFound) || row == nil {
		return nil, ErrMovieAlbumNotFound
	}
	if err != nil {
		return nil, err
	}
	if IsProtectedMovieAlbumName(row.Name) {
		return nil, ErrMovieAlbumUpdateBlocked
	}

	allRows, err := s.deps.MovieAlbumModel.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, existing := range allRows {
		if existing != nil && existing.Id != albumID && strings.EqualFold(strings.TrimSpace(existing.Name), name) {
			return nil, ErrMovieAlbumNameExists
		}
	}

	row.Name = name
	row.UpdatedOn = time.Now().Unix()
	if err := s.deps.MovieAlbumModel.Update(ctx, row); err != nil {
		again, againErr := s.deps.MovieAlbumModel.FindOneByName(ctx, name)
		if againErr == nil && again != nil && again.Id != albumID {
			return nil, ErrMovieAlbumNameExists
		}
		return nil, err
	}
	return &UpdateMovieAlbumResult{ID: row.Id, Name: name}, nil
}

func (s *Service) DeleteMovieAlbum(ctx context.Context, albumID int64) (*DeleteMovieAlbumResult, error) {
	if albumID <= 0 {
		return nil, ErrMovieAlbumInvalidID
	}

	row, err := s.deps.MovieAlbumModel.FindOne(ctx, albumID)
	if errors.Is(err, moviex.ErrNotFound) || row == nil {
		return nil, ErrMovieAlbumNotFound
	}
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(row.Name)
	if IsProtectedMovieAlbumName(name) {
		return nil, ErrMovieAlbumDeleteBlocked
	}

	itemCount, err := s.deps.MovieAlbumItemModel.CountPageRows(ctx, albumID, moviex.CMovieAlbumItemPageFilter{})
	if err != nil {
		return nil, err
	}
	if err := validateMovieAlbumDeletion(itemCount); err != nil {
		return nil, err
	}

	if err := s.deps.MovieAlbumModel.DeleteIfEmpty(ctx, albumID); err != nil {
		switch {
		case errors.Is(err, moviex.ErrNotFound):
			return nil, ErrMovieAlbumNotFound
		case errors.Is(err, moviex.ErrCMovieAlbumNotEmpty):
			return nil, ErrMovieAlbumNotEmpty
		default:
			return nil, err
		}
	}
	return &DeleteMovieAlbumResult{ID: albumID, Name: name}, nil
}

func (s *Service) ensureMovieNeedDownloadAlbum(ctx context.Context) (int64, error) {
	return s.getMovieAlbumIDByName(ctx, normalizeMovieNeedDownloadAlbumName(), true)
}

func (s *Service) ensureMovieDeleteAlbum(ctx context.Context) (int64, error) {
	return s.getMovieAlbumIDByName(ctx, normalizeMovieDeleteAlbumName(), true)
}
