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
	ErrMovieAlbumInvalidID                   = errors.New("无效电影相册")
	ErrMovieAlbumItemInvalid                 = errors.New("无效电影相册条目")
	ErrMovieAlbumTargetInvalid               = errors.New("无效目标电影相册")
	ErrMovieAlbumTargetSame                  = errors.New("目标电影相册不能与当前相册相同")
	ErrMovieAlbumUnsupportedRemoveDownloaded = errors.New("当前电影相册不支持移除已下载条目")
)

type MovieAlbumItemRow struct {
	ID            int64
	AlbumID       int64
	MovieJavID    string
	MovieName     string
	ReleasingDate int64
	SortNo        int64
	CreatedOn     int64
}

type RemoveDownloadedMovieAlbumItemsResult struct {
	RemovedCount int64
	SkippedCount int64
}

type RemoveDownloadedMovieAlbumPreviewItem struct {
	ItemID     int64
	MovieJavID string
	MovieName  string
}

type RemoveDownloadedMovieAlbumPreviewResult struct {
	Items []*RemoveDownloadedMovieAlbumPreviewItem
	Count int64
}

type TransferMovieAlbumItemsResult struct {
	ProcessedCount int64
	SkippedCount   int64
	FailedIDs      []int64
}

func (s *Service) AddMovieToAlbum(ctx context.Context, albumName string, movieJavID string) (bool, string, error) {
	normalizedAlbumName := normalizeMovieAlbumName(albumName)
	normalizedMovieJavID := strings.TrimSpace(movieJavID)
	if normalizedAlbumName == "" {
		return false, normalizedAlbumName, ErrMovieAlbumNameEmpty
	}
	if normalizedMovieJavID == "" {
		return false, normalizedAlbumName, ErrInvalidSourceRow
	}

	movieName, releasingDate, err := s.loadMovieNameAndReleasingDateByJavID(ctx, normalizedMovieJavID)
	if err != nil {
		return false, normalizedAlbumName, err
	}

	albumID, err := s.getMovieAlbumIDByName(ctx, normalizedAlbumName, true)
	if err != nil {
		return false, normalizedAlbumName, err
	}

	_, err = s.deps.MovieAlbumItemModel.FindOneByAlbumIdMovieJavId(ctx, albumID, normalizedMovieJavID)
	if err == nil {
		return false, normalizedAlbumName, nil
	}
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return false, normalizedAlbumName, err
	}

	sortNo, err := s.nextMovieAlbumSortNo(ctx, albumID)
	if err != nil {
		return false, normalizedAlbumName, err
	}
	now := time.Now().Unix()
	_, err = s.deps.MovieAlbumItemModel.Insert(ctx, &moviex.CMovieAlbumItem{
		AlbumId:       albumID,
		MovieJavId:    normalizedMovieJavID,
		MovieName:     movieName,
		ReleasingDate: releasingDate,
		SortNo:        sortNo,
		CreatedOn:     now,
		UpdatedOn:     now,
	})
	if err != nil {
		return false, normalizedAlbumName, err
	}
	return true, normalizedAlbumName, nil
}

func (s *Service) RemoveMovieAlbumItemByID(ctx context.Context, albumID int64, itemID int64) (bool, error) {
	if albumID <= 0 {
		return false, ErrMovieAlbumInvalidID
	}
	if itemID <= 0 {
		return false, ErrMovieAlbumItemInvalid
	}
	row, err := s.deps.MovieAlbumItemModel.FindOne(ctx, itemID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	if row == nil || row.AlbumId != albumID {
		return false, nil
	}
	if err := s.deps.MovieAlbumItemModel.Delete(ctx, row.Id); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) MoveMovieAlbumItemsByIDs(ctx context.Context, sourceAlbumID int64, targetAlbumID int64, itemIDs []int64) (*TransferMovieAlbumItemsResult, error) {
	return s.transferMovieAlbumItemsByIDs(ctx, sourceAlbumID, targetAlbumID, itemIDs, true)
}

func (s *Service) CopyMovieAlbumItemsByIDs(ctx context.Context, sourceAlbumID int64, targetAlbumID int64, itemIDs []int64) (*TransferMovieAlbumItemsResult, error) {
	return s.transferMovieAlbumItemsByIDs(ctx, sourceAlbumID, targetAlbumID, itemIDs, false)
}

func (s *Service) transferMovieAlbumItemsByIDs(ctx context.Context, sourceAlbumID int64, targetAlbumID int64, itemIDs []int64, move bool) (*TransferMovieAlbumItemsResult, error) {
	if sourceAlbumID <= 0 {
		return nil, ErrMovieAlbumInvalidID
	}
	if targetAlbumID <= 0 {
		return nil, ErrMovieAlbumTargetInvalid
	}
	if sourceAlbumID == targetAlbumID {
		return nil, ErrMovieAlbumTargetSame
	}
	if _, err := s.deps.MovieAlbumModel.FindOne(ctx, sourceAlbumID); err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, ErrMovieAlbumInvalidID
		}
		return nil, err
	}
	if _, err := s.deps.MovieAlbumModel.FindOne(ctx, targetAlbumID); err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, ErrMovieAlbumTargetInvalid
		}
		return nil, err
	}

	result := &TransferMovieAlbumItemsResult{FailedIDs: []int64{}}
	ids := uniquePositiveIDs(itemIDs)
	if len(ids) == 0 {
		return result, nil
	}
	nextSortNo, err := s.nextMovieAlbumSortNo(ctx, targetAlbumID)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()

	for _, itemID := range ids {
		row, err := s.deps.MovieAlbumItemModel.FindOne(ctx, itemID)
		if err != nil || row == nil || row.AlbumId != sourceAlbumID {
			result.FailedIDs = append(result.FailedIDs, itemID)
			continue
		}

		existing, findErr := s.deps.MovieAlbumItemModel.FindOneByAlbumIdMovieJavId(ctx, targetAlbumID, row.MovieJavId)
		if findErr == nil && existing != nil {
			if !move {
				result.SkippedCount++
				continue
			}
			if err := s.deps.MovieAlbumItemModel.Delete(ctx, row.Id); err != nil {
				result.FailedIDs = append(result.FailedIDs, itemID)
				continue
			}
			result.ProcessedCount++
			continue
		}
		if findErr != nil && !errors.Is(findErr, moviex.ErrNotFound) {
			result.FailedIDs = append(result.FailedIDs, itemID)
			continue
		}

		if move {
			row.AlbumId = targetAlbumID
			row.SortNo = nextSortNo
			row.UpdatedOn = now
			if err := s.deps.MovieAlbumItemModel.Update(ctx, row); err != nil {
				result.FailedIDs = append(result.FailedIDs, itemID)
				continue
			}
		} else {
			if _, err := s.deps.MovieAlbumItemModel.Insert(ctx, &moviex.CMovieAlbumItem{
				AlbumId:       targetAlbumID,
				MovieJavId:    row.MovieJavId,
				MovieName:     row.MovieName,
				ReleasingDate: row.ReleasingDate,
				SortNo:        nextSortNo,
				CreatedOn:     now,
				UpdatedOn:     now,
			}); err != nil {
				result.FailedIDs = append(result.FailedIDs, itemID)
				continue
			}
		}

		nextSortNo++
		result.ProcessedCount++
	}

	return result, nil
}

func (s *Service) RemoveMovieFromAlbum(ctx context.Context, albumName string, movieJavID string) (bool, string, error) {
	normalizedAlbumName := normalizeMovieAlbumName(albumName)
	normalizedMovieJavID := strings.TrimSpace(movieJavID)
	if normalizedAlbumName == "" {
		return false, normalizedAlbumName, ErrMovieAlbumNameEmpty
	}
	if normalizedMovieJavID == "" {
		return false, normalizedAlbumName, ErrMovieAlbumItemInvalid
	}

	albumID, err := s.getMovieAlbumIDByName(ctx, normalizedAlbumName, false)
	if err != nil {
		return false, normalizedAlbumName, err
	}
	if albumID <= 0 {
		return false, normalizedAlbumName, nil
	}

	removed, err := s.deps.MovieAlbumItemModel.DeleteByAlbumIdMovieJavId(ctx, albumID, normalizedMovieJavID)
	return removed, normalizedAlbumName, err
}

func (s *Service) RemoveDownloadedMovieAlbumItems(ctx context.Context, albumID int64) (*RemoveDownloadedMovieAlbumItemsResult, error) {
	preview, err := s.PreviewRemoveDownloadedMovieAlbumItems(ctx, albumID)
	if err != nil {
		return nil, err
	}

	result := &RemoveDownloadedMovieAlbumItemsResult{}
	if preview == nil || len(preview.Items) == 0 {
		return result, nil
	}
	for _, item := range preview.Items {
		if item == nil || item.ItemID <= 0 {
			result.SkippedCount++
			continue
		}
		removed, err := s.RemoveMovieAlbumItemByID(ctx, albumID, item.ItemID)
		if err != nil {
			return nil, err
		}
		if removed {
			result.RemovedCount++
			continue
		}
		result.SkippedCount++
	}

	return result, nil
}

func (s *Service) PreviewRemoveDownloadedMovieAlbumItems(ctx context.Context, albumID int64) (*RemoveDownloadedMovieAlbumPreviewResult, error) {
	items, err := s.listDownloadedMovieAlbumItemsForRemoval(ctx, albumID)
	if err != nil {
		return nil, err
	}
	return &RemoveDownloadedMovieAlbumPreviewResult{
		Items: items,
		Count: int64(len(items)),
	}, nil
}

func (s *Service) ListMovieAlbumsByMovieJavID(ctx context.Context, movieJavID string) ([]*MovieAlbumOption, error) {
	options, err := s.ListMovieAlbumOptions(ctx)
	if err != nil {
		return nil, err
	}
	normalizedMovieJavID := strings.TrimSpace(movieJavID)
	if normalizedMovieJavID == "" || len(options) == 0 {
		return options, nil
	}
	rows, err := s.deps.MovieAlbumItemModel.ListByMovieJavId(ctx, normalizedMovieJavID)
	if err != nil {
		return nil, err
	}
	selected := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		selected[row.AlbumId] = struct{}{}
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		_, option.Selected = selected[option.ID]
	}
	return options, nil
}

func (s *Service) GetMovieNeedDownloadStatus(ctx context.Context, movieJavID string) (int64, error) {
	inAlbum, err := s.IsMovieInMovieAlbum(ctx, normalizeMovieNeedDownloadAlbumName(), movieJavID)
	if err != nil {
		return 0, err
	}
	if inAlbum {
		return consts.MovieNeedDownLoadOK, nil
	}
	return consts.MovieNeedDownLoadNone, nil
}

func (s *Service) IsMovieMarkedDelete(ctx context.Context, movieJavID string) (bool, error) {
	return s.IsMovieInMovieAlbum(ctx, normalizeMovieDeleteAlbumName(), movieJavID)
}

func (s *Service) IsMovieInMovieAlbum(ctx context.Context, albumName string, movieJavID string) (bool, error) {
	normalizedAlbumName := normalizeMovieAlbumName(albumName)
	normalizedMovieJavID := strings.TrimSpace(movieJavID)
	if normalizedAlbumName == "" || normalizedMovieJavID == "" {
		return false, nil
	}

	albumID, err := s.getMovieAlbumIDByName(ctx, normalizedAlbumName, false)
	if err != nil {
		return false, err
	}
	if albumID <= 0 {
		return false, nil
	}

	_, err = s.deps.MovieAlbumItemModel.FindOneByAlbumIdMovieJavId(ctx, albumID, normalizedMovieJavID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, moviex.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (s *Service) getMovieAlbumIDByName(ctx context.Context, albumName string, createWhenMissing bool) (int64, error) {
	normalizedAlbumName := normalizeMovieAlbumName(albumName)
	if normalizedAlbumName == "" {
		return 0, ErrMovieAlbumNameEmpty
	}
	album, err := s.deps.MovieAlbumModel.FindOneByName(ctx, normalizedAlbumName)
	if err == nil && album != nil {
		return album.Id, nil
	}
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return 0, err
	}
	if !createWhenMissing {
		return 0, nil
	}

	now := time.Now().Unix()
	result, err := s.deps.MovieAlbumModel.Insert(ctx, &moviex.CMovieAlbum{
		Name:      normalizedAlbumName,
		Remark:    "默认电影相册",
		CreatedOn: now,
		UpdatedOn: now,
	})
	if err != nil {
		albumAgain, againErr := s.deps.MovieAlbumModel.FindOneByName(ctx, normalizedAlbumName)
		if againErr == nil && albumAgain != nil {
			return albumAgain.Id, nil
		}
		return 0, err
	}
	insertID, idErr := result.LastInsertId()
	if idErr == nil && insertID > 0 {
		return insertID, nil
	}
	albumAgain, againErr := s.deps.MovieAlbumModel.FindOneByName(ctx, normalizedAlbumName)
	if againErr != nil {
		return 0, againErr
	}
	return albumAgain.Id, nil
}

func (s *Service) nextMovieAlbumSortNo(ctx context.Context, albumID int64) (int64, error) {
	rows, err := s.deps.MovieAlbumItemModel.ListPageRows(ctx, albumID, 0, 1, "`sort_no` DESC, `id` DESC", moviex.CMovieAlbumItemPageFilter{})
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 || rows[0] == nil {
		return 1, nil
	}
	return rows[0].SortNo + 1, nil
}

func canRemoveDownloadedMovieAlbum(albumName string) bool {
	switch strings.TrimSpace(albumName) {
	case consts.MovieNeedDownloadAlbumName, "v3", "vx":
		return true
	default:
		return false
	}
}

func (s *Service) listDownloadedMovieAlbumItemsForRemoval(ctx context.Context, albumID int64) ([]*RemoveDownloadedMovieAlbumPreviewItem, error) {
	if albumID <= 0 {
		return nil, ErrMovieAlbumInvalidID
	}

	albumRow, err := s.deps.MovieAlbumModel.FindOne(ctx, albumID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, ErrMovieAlbumInvalidID
		}
		return nil, err
	}
	if albumRow == nil || !canRemoveDownloadedMovieAlbum(strings.TrimSpace(albumRow.Name)) {
		return nil, ErrMovieAlbumUnsupportedRemoveDownloaded
	}

	rows, err := s.deps.MovieAlbumItemModel.ListByAlbumId(ctx, albumID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []*RemoveDownloadedMovieAlbumPreviewItem{}, nil
	}

	uniqJavIDs := make([]string, 0, len(rows))
	seenJavIDs := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		javID := strings.TrimSpace(row.MovieJavId)
		if javID == "" {
			continue
		}
		if _, ok := seenJavIDs[javID]; ok {
			continue
		}
		seenJavIDs[javID] = struct{}{}
		uniqJavIDs = append(uniqJavIDs, javID)
	}

	downloadedByJavID := make(map[string]struct{}, len(uniqJavIDs))
	if len(uniqJavIDs) > 0 {
		mediaRows, err := s.deps.WMediaModel.ListByMovieJavIds(ctx, uniqJavIDs)
		if err != nil {
			return nil, err
		}
		for _, row := range mediaRows {
			if row == nil || row.IsRemoved != consts.FilmIsNotRemoved {
				continue
			}
			javID := strings.TrimSpace(row.MovieJavId)
			if javID == "" {
				continue
			}
			downloadedByJavID[javID] = struct{}{}
		}
	}

	out := make([]*RemoveDownloadedMovieAlbumPreviewItem, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		javID := strings.TrimSpace(row.MovieJavId)
		if _, ok := downloadedByJavID[javID]; !ok {
			continue
		}
		out = append(out, &RemoveDownloadedMovieAlbumPreviewItem{
			ItemID:     row.Id,
			MovieJavID: javID,
			MovieName:  strings.TrimSpace(row.MovieName),
		})
	}
	return out, nil
}
