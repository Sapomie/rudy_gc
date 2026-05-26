package sehuatang

import (
	"context"
	"errors"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
)

const (
	sehuatangSourceType = "sehuatang_magnet"
	downloadAlbumName   = "下载中"
	pendingDownloadName = "待下载"
)

func (s *Service) fillAlbumStatus(ctx context.Context, rows []*ListRow) error {
	if len(rows) == 0 {
		return nil
	}

	downloadAlbumID, err := s.findAlbumIDByName(ctx, downloadAlbumName)
	if err != nil {
		return err
	}
	pendingAlbumID, err := s.findAlbumIDByName(ctx, pendingDownloadName)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if row == nil || row.Id <= 0 {
			continue
		}
		row.CanFavorite = strings.TrimSpace(row.MovieRouteKey) != ""

		if downloadAlbumID > 0 {
			exists, err := s.existsInAlbum(ctx, downloadAlbumID, row.Id)
			if err != nil {
				return err
			}
			row.IsInDownload = exists
		}
		if pendingAlbumID > 0 {
			exists, err := s.existsInAlbum(ctx, pendingAlbumID, row.Id)
			if err != nil {
				return err
			}
			row.IsInPending = exists
		}
	}
	return nil
}

func (s *Service) findAlbumIDByName(ctx context.Context, albumName string) (int64, error) {
	row, err := s.deps.AlbumModel.FindOneByName(ctx, strings.TrimSpace(albumName))
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	if row == nil {
		return 0, nil
	}
	return row.Id, nil
}

func (s *Service) existsInAlbum(ctx context.Context, albumID int64, sourceRowID int64) (bool, error) {
	_, err := s.deps.AlbumItemModel.FindOneByAlbumIdSourceTypeSourceRowId(ctx, albumID, sehuatangSourceType, sourceRowID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, moviex.ErrNotFound) {
		return false, nil
	}
	return false, err
}
