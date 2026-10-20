package movie

import (
	"context"
	"errors"
	"strings"
	"time"

	"rudy_gc/local/api/internal/model/modelx/moviex"
)

var (
	ErrAlbumNameEmpty     = errors.New("相册名称不能为空")
	ErrAlbumNameTooLong   = errors.New("相册名称过长")
	ErrAlbumNameExists    = errors.New("相册名称已存在")
	ErrAlbumInvalidID     = errors.New("相册 ID 无效")
	ErrAlbumNotFound      = errors.New("相册不存在")
	ErrAlbumDeleteBlocked = errors.New("系统相册不允许删除")
	ErrAlbumUpdateBlocked = errors.New("系统相册不允许编辑")
	ErrAlbumNotEmpty      = errors.New("相册中还有数据，不能删除")
)

type DeleteAlbumResult struct {
	AlbumID   int64
	AlbumName string
}

type UpdateAlbumResult struct {
	AlbumID   int64
	AlbumName string
}

func IsProtectedTorrentAlbumName(albumName string) bool {
	switch strings.TrimSpace(albumName) {
	case "下载中", "待下载", "Media", "FcMedia":
		return true
	default:
		return false
	}
}

func validateAlbumDeletion(itemCount int64) error {
	if itemCount > 0 {
		return ErrAlbumNotEmpty
	}
	return nil
}

func (s *Service) CreateAlbum(ctx context.Context, albumName string) (*TorrentAlbumOption, error) {
	name := strings.TrimSpace(albumName)
	if name == "" {
		return nil, ErrAlbumNameEmpty
	}
	if len([]byte(name)) > 128 {
		return nil, ErrAlbumNameTooLong
	}

	row, err := s.deps.AlbumModel.FindOneByName(ctx, name)
	if err == nil && row != nil {
		return nil, ErrAlbumNameExists
	}
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, err
	}

	now := time.Now().Unix()
	result, err := s.deps.AlbumModel.Insert(ctx, &moviex.TAlbum{
		Name:      name,
		Remark:    "页面新增相册",
		CreatedOn: now,
		UpdatedOn: now,
	})
	if err != nil {
		again, againErr := s.deps.AlbumModel.FindOneByName(ctx, name)
		if againErr == nil && again != nil {
			return nil, ErrAlbumNameExists
		}
		return nil, err
	}

	insertID, idErr := result.LastInsertId()
	if idErr == nil && insertID > 0 {
		return &TorrentAlbumOption{
			ID:   insertID,
			Name: name,
		}, nil
	}

	again, againErr := s.deps.AlbumModel.FindOneByName(ctx, name)
	if againErr != nil {
		return nil, againErr
	}
	return &TorrentAlbumOption{
		ID:   again.Id,
		Name: strings.TrimSpace(again.Name),
	}, nil
}

func (s *Service) DeleteAlbum(ctx context.Context, albumID int64) (*DeleteAlbumResult, error) {
	if albumID <= 0 {
		return nil, ErrAlbumInvalidID
	}

	row, err := s.deps.AlbumModel.FindOne(ctx, albumID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, ErrAlbumNotFound
		}
		return nil, err
	}
	if row == nil {
		return nil, ErrAlbumNotFound
	}
	if IsProtectedTorrentAlbumName(row.Name) {
		return nil, ErrAlbumDeleteBlocked
	}

	itemCount, err := s.deps.AlbumItemModel.CountPageRows(ctx, albumID, moviex.AlbumItemPageFilter{})
	if err != nil {
		return nil, err
	}
	if err := validateAlbumDeletion(itemCount); err != nil {
		return nil, err
	}

	err = s.deps.AlbumModel.DeleteIfEmpty(ctx, albumID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, ErrAlbumNotFound
		}
		if errors.Is(err, moviex.ErrTAlbumNotEmpty) {
			return nil, ErrAlbumNotEmpty
		}
		return nil, err
	}
	return &DeleteAlbumResult{
		AlbumID:   albumID,
		AlbumName: strings.TrimSpace(row.Name),
	}, nil
}

func (s *Service) UpdateAlbum(ctx context.Context, albumID int64, albumName string) (*UpdateAlbumResult, error) {
	if albumID <= 0 {
		return nil, ErrAlbumInvalidID
	}
	name := strings.TrimSpace(albumName)
	if name == "" {
		return nil, ErrAlbumNameEmpty
	}
	if len([]byte(name)) > 128 {
		return nil, ErrAlbumNameTooLong
	}

	row, err := s.deps.AlbumModel.FindOne(ctx, albumID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, ErrAlbumNotFound
		}
		return nil, err
	}
	if row == nil {
		return nil, ErrAlbumNotFound
	}
	if IsProtectedTorrentAlbumName(row.Name) {
		return nil, ErrAlbumUpdateBlocked
	}

	albums, err := s.deps.AlbumModel.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, existing := range albums {
		if existing != nil && existing.Id != albumID && strings.EqualFold(strings.TrimSpace(existing.Name), name) {
			return nil, ErrAlbumNameExists
		}
	}

	row.Name = name
	row.UpdatedOn = time.Now().Unix()
	if err := s.deps.AlbumModel.Update(ctx, row); err != nil {
		again, againErr := s.deps.AlbumModel.FindOneByName(ctx, name)
		if againErr == nil && again != nil && again.Id != albumID {
			return nil, ErrAlbumNameExists
		}
		return nil, err
	}
	return &UpdateAlbumResult{AlbumID: albumID, AlbumName: name}, nil
}
