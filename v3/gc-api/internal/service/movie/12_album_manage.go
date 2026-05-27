package movie

import (
	"context"
	"errors"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

var (
	ErrAlbumNameEmpty   = errors.New("相册名称不能为空")
	ErrAlbumNameTooLong = errors.New("相册名称过长")
	ErrAlbumNameExists  = errors.New("相册名称已存在")
)

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
