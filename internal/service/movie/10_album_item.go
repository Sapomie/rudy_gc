package movie

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

const (
	defaultAlbumName       = "下载中"
	pendingAlbumName       = "待下载"
	sourceTypeJavbusMagnet = "javbus_magnet"
	sourceTypeSukebei      = "sukebei_torrent"
	sourceTypeSehuatang    = "sehuatang_magnet"
)

var (
	ErrInvalidSourceType = errors.New("无效来源类型")
	ErrInvalidSourceRow  = errors.New("无效来源记录")
	ErrSourceNotFound    = errors.New("来源记录不存在")
	ErrSourceMovieMiss   = errors.New("来源记录与影片不匹配")
	ErrInvalidAlbumID    = errors.New("无效相册")
	ErrInvalidAlbumItem  = errors.New("无效相册条目")
	ErrAlbumMismatch     = errors.New("相册条目不属于当前相册")
	ErrTargetAlbumID     = errors.New("无效目标相册")
	ErrTargetAlbumSame   = errors.New("目标相册不能与当前相册相同")
)

type albumSourcePayload struct {
	MovieName   string
	MovieJavID  string
	InfoHash    string
	Size        string
	PublishTime int64
}

func normalizeAlbumName(raw string) string {
	name := strings.TrimSpace(raw)
	if name == "" {
		return defaultAlbumName
	}
	return name
}

func (s *Service) AddFetchResourceToAlbum(ctx context.Context, albumName, movieJavID string, sourceType string, sourceRowID int64) (bool, string, error) {
	normalizedAlbumName := normalizeAlbumName(albumName)
	normalizedMovieJavID := strings.TrimSpace(movieJavID)
	if sourceRowID <= 0 {
		return false, normalizedAlbumName, ErrInvalidSourceRow
	}

	normalizedSourceType := strings.TrimSpace(sourceType)
	if normalizedMovieJavID == "" && normalizedSourceType != sourceTypeSehuatang {
		return false, normalizedAlbumName, ErrInvalidSourceRow
	}
	sourcePayload, err := s.readSourcePayload(ctx, normalizedMovieJavID, normalizedSourceType, sourceRowID)
	if err != nil {
		return false, normalizedAlbumName, err
	}

	albumID, err := s.getAlbumIDByName(ctx, normalizedAlbumName, true)
	if err != nil {
		return false, normalizedAlbumName, err
	}

	_, err = s.deps.AlbumItemModel.FindOneByAlbumIdSourceTypeSourceRowId(ctx, albumID, normalizedSourceType, sourceRowID)
	if err == nil {
		return false, normalizedAlbumName, nil
	}
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return false, normalizedAlbumName, err
	}

	now := time.Now().Unix()
	_, err = s.deps.AlbumItemModel.Insert(ctx, &moviex.TmAlbumItem{
		AlbumId:     albumID,
		SourceType:  normalizedSourceType,
		SourceRowId: sourceRowID,
		MovieJavId:  strings.TrimSpace(sourcePayload.MovieJavID),
		MovieName:   sourcePayload.MovieName,
		InfoHash:    sourcePayload.InfoHash,
		Size:        sourcePayload.Size,
		PublishTime: sourcePayload.PublishTime,
		CreatedOn:   now,
		UpdatedOn:   now,
	})
	if err != nil {
		return false, normalizedAlbumName, err
	}
	return true, normalizedAlbumName, nil
}

func (s *Service) AddFetchResourceToDefaultAlbum(ctx context.Context, movieJavID string, sourceType string, sourceRowID int64) (bool, error) {
	created, _, err := s.AddFetchResourceToAlbum(ctx, defaultAlbumName, movieJavID, sourceType, sourceRowID)
	return created, err
}

func (s *Service) RemoveFetchResourceFromAlbum(ctx context.Context, albumName, movieJavID string, sourceType string, sourceRowID int64) (bool, string, error) {
	normalizedAlbumName := normalizeAlbumName(albumName)
	normalizedMovieJavID := strings.TrimSpace(movieJavID)
	if sourceRowID <= 0 {
		return false, normalizedAlbumName, ErrInvalidSourceRow
	}
	normalizedSourceType := strings.TrimSpace(sourceType)
	if normalizedMovieJavID == "" && normalizedSourceType != sourceTypeSehuatang {
		return false, normalizedAlbumName, ErrInvalidSourceRow
	}
	if _, err := s.readSourcePayload(ctx, normalizedMovieJavID, normalizedSourceType, sourceRowID); err != nil {
		return false, normalizedAlbumName, err
	}

	albumID, err := s.getAlbumIDByName(ctx, normalizedAlbumName, false)
	if err != nil {
		return false, normalizedAlbumName, err
	}
	if albumID <= 0 {
		return false, normalizedAlbumName, nil
	}

	removed, err := s.deps.AlbumItemModel.DeleteByAlbumIdSourceTypeSourceRowId(ctx, albumID, normalizedSourceType, sourceRowID)
	return removed, normalizedAlbumName, err
}

func (s *Service) RemoveFetchResourceFromDefaultAlbum(ctx context.Context, movieJavID string, sourceType string, sourceRowID int64) (bool, error) {
	removed, _, err := s.RemoveFetchResourceFromAlbum(ctx, defaultAlbumName, movieJavID, sourceType, sourceRowID)
	return removed, err
}

func (s *Service) ListAlbumItemsByMovieJavID(ctx context.Context, albumName, movieJavID string) ([]*moviex.TmAlbumItem, string, error) {
	normalizedAlbumName := normalizeAlbumName(albumName)
	normalizedMovieJavID := strings.TrimSpace(movieJavID)
	if normalizedMovieJavID == "" {
		return []*moviex.TmAlbumItem{}, normalizedAlbumName, nil
	}

	albumID, err := s.getAlbumIDByName(ctx, normalizedAlbumName, false)
	if err != nil {
		return nil, normalizedAlbumName, err
	}
	if albumID <= 0 {
		return []*moviex.TmAlbumItem{}, normalizedAlbumName, nil
	}
	items, err := s.deps.AlbumItemModel.ListByAlbumIdMovieJavId(ctx, albumID, normalizedMovieJavID)
	return items, normalizedAlbumName, err
}

func (s *Service) ListDefaultAlbumItemsByMovieJavID(ctx context.Context, movieJavID string) ([]*moviex.TmAlbumItem, error) {
	items, _, err := s.ListAlbumItemsByMovieJavID(ctx, defaultAlbumName, movieJavID)
	return items, err
}

func (s *Service) RemoveAlbumItemByID(ctx context.Context, albumID int64, itemID int64) (bool, error) {
	if albumID <= 0 {
		return false, ErrInvalidAlbumID
	}
	if itemID <= 0 {
		return false, ErrInvalidAlbumItem
	}

	row, err := s.deps.AlbumItemModel.FindOne(ctx, itemID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	if row == nil {
		return false, nil
	}
	if row.AlbumId != albumID {
		return false, ErrAlbumMismatch
	}

	if err := s.deps.AlbumItemModel.Delete(ctx, row.Id); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) RemoveAlbumItemsByIDs(ctx context.Context, albumID int64, itemIDs []int64) (int64, []int64, error) {
	if albumID <= 0 {
		return 0, nil, ErrInvalidAlbumID
	}
	uniqIDs := uniquePositiveIDs(itemIDs)
	if len(uniqIDs) == 0 {
		return 0, []int64{}, nil
	}

	removedCount := int64(0)
	failedIDs := make([]int64, 0)
	for _, itemID := range uniqIDs {
		removed, err := s.RemoveAlbumItemByID(ctx, albumID, itemID)
		if err != nil {
			failedIDs = append(failedIDs, itemID)
			continue
		}
		if removed {
			removedCount++
		}
	}
	return removedCount, failedIDs, nil
}

func (s *Service) MoveAlbumItemsByIDs(ctx context.Context, sourceAlbumID int64, targetAlbumID int64, itemIDs []int64) (int64, []int64, error) {
	if sourceAlbumID <= 0 {
		return 0, nil, ErrInvalidAlbumID
	}
	if targetAlbumID <= 0 {
		return 0, nil, ErrTargetAlbumID
	}
	if sourceAlbumID == targetAlbumID {
		return 0, nil, ErrTargetAlbumSame
	}
	if _, err := s.deps.AlbumModel.FindOne(ctx, targetAlbumID); err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return 0, nil, ErrTargetAlbumID
		}
		return 0, nil, err
	}

	uniqIDs := uniquePositiveIDs(itemIDs)
	if len(uniqIDs) == 0 {
		return 0, []int64{}, nil
	}

	now := time.Now().Unix()
	movedCount := int64(0)
	failedIDs := make([]int64, 0)
	for _, itemID := range uniqIDs {
		row, err := s.deps.AlbumItemModel.FindOne(ctx, itemID)
		if err != nil {
			failedIDs = append(failedIDs, itemID)
			continue
		}
		if row == nil || row.AlbumId != sourceAlbumID {
			failedIDs = append(failedIDs, itemID)
			continue
		}

		existing, err := s.deps.AlbumItemModel.FindOneByAlbumIdSourceTypeSourceRowId(ctx, targetAlbumID, row.SourceType, row.SourceRowId)
		if err == nil && existing != nil {
			if err := s.deps.AlbumItemModel.Delete(ctx, row.Id); err != nil {
				failedIDs = append(failedIDs, itemID)
				continue
			}
			movedCount++
			continue
		}
		if err != nil && !errors.Is(err, moviex.ErrNotFound) {
			failedIDs = append(failedIDs, itemID)
			continue
		}

		row.AlbumId = targetAlbumID
		row.UpdatedOn = now
		if err := s.deps.AlbumItemModel.Update(ctx, row); err != nil {
			failedIDs = append(failedIDs, itemID)
			continue
		}
		movedCount++
	}
	return movedCount, failedIDs, nil
}

func uniquePositiveIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return []int64{}
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (s *Service) readSourcePayload(ctx context.Context, movieJavID string, sourceType string, sourceRowID int64) (*albumSourcePayload, error) {
	switch sourceType {
	case sourceTypeJavbusMagnet:
		row, err := s.deps.JavbusMagnetModel.FindOne(ctx, sourceRowID)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				return nil, ErrSourceNotFound
			}
			return nil, err
		}
		if !strings.EqualFold(strings.TrimSpace(row.MovieJavId), movieJavID) {
			return nil, ErrSourceMovieMiss
		}
		movieName, err := s.loadMovieNameByJavID(ctx, row.MovieJavId)
		if err != nil {
			return nil, err
		}
		return &albumSourcePayload{
			MovieName:   movieName,
			MovieJavID:  strings.TrimSpace(row.MovieJavId),
			InfoHash:    strings.TrimSpace(row.InfoHash),
			Size:        formatResourceSizeBytes(row.SizeBytes),
			PublishTime: row.ShareDate,
		}, nil
	case sourceTypeSukebei:
		row, err := s.deps.SukebeiTorrentModel.FindOne(ctx, sourceRowID)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				return nil, ErrSourceNotFound
			}
			return nil, err
		}
		if !strings.EqualFold(strings.TrimSpace(row.MovieJavId), movieJavID) {
			return nil, ErrSourceMovieMiss
		}
		movieName, err := s.loadMovieNameByJavID(ctx, row.MovieJavId)
		if err != nil {
			return nil, err
		}
		return &albumSourcePayload{
			MovieName:   movieName,
			MovieJavID:  strings.TrimSpace(row.MovieJavId),
			InfoHash:    strings.TrimSpace(row.InfoHash),
			Size:        formatResourceSizeBytes(row.SizeBytes),
			PublishTime: row.PublishTime,
		}, nil
	case sourceTypeSehuatang:
		row, err := s.deps.SehuatangMagnetModel.FindOne(ctx, sourceRowID)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				return nil, ErrSourceNotFound
			}
			return nil, err
		}
		actualMovieJavID := strings.TrimSpace(row.MovieJavId)
		if movieJavID != "" && actualMovieJavID != "" && !strings.EqualFold(actualMovieJavID, movieJavID) {
			return nil, ErrSourceMovieMiss
		}
		movieName := strings.TrimSpace(row.MovieName)
		if actualMovieJavID != "" {
			movieNameByJavID, err := s.loadMovieNameByJavID(ctx, actualMovieJavID)
			if err != nil {
				return nil, err
			}
			movieName = strings.TrimSpace(movieNameByJavID)
		}
		if movieName == "" {
			movieName = strings.TrimSpace(row.ThreadTitle)
		}
		return &albumSourcePayload{
			MovieName:   movieName,
			MovieJavID:  actualMovieJavID,
			InfoHash:    strings.TrimSpace(row.InfoHash),
			Size:        "",
			PublishTime: row.PostTime,
		}, nil
	default:
		return nil, ErrInvalidSourceType
	}
}

func (s *Service) loadMovieNameByJavID(ctx context.Context, movieJavID string) (string, error) {
	row, err := s.deps.MovieModel.FindOneByJavId(ctx, strings.TrimSpace(movieJavID))
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return strings.TrimSpace(movieJavID), nil
		}
		return "", err
	}
	movieName := strings.TrimSpace(row.Name)
	if movieName == "" {
		return strings.TrimSpace(movieJavID), nil
	}
	return movieName, nil
}

func (s *Service) getDefaultAlbumID(ctx context.Context, createWhenMissing bool) (int64, error) {
	return s.getAlbumIDByName(ctx, defaultAlbumName, createWhenMissing)
}

func (s *Service) getAlbumIDByName(ctx context.Context, albumName string, createWhenMissing bool) (int64, error) {
	normalizedAlbumName := normalizeAlbumName(albumName)

	album, err := s.deps.AlbumModel.FindOneByName(ctx, normalizedAlbumName)
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
	result, err := s.deps.AlbumModel.Insert(ctx, &moviex.TAlbum{
		Name:      normalizedAlbumName,
		Remark:    fmt.Sprintf("默认%s相册", normalizedAlbumName),
		CreatedOn: now,
		UpdatedOn: now,
	})
	if err != nil {
		albumAgain, againErr := s.deps.AlbumModel.FindOneByName(ctx, normalizedAlbumName)
		if againErr == nil && albumAgain != nil {
			return albumAgain.Id, nil
		}
		return 0, fmt.Errorf("创建默认%s相册失败: %w", normalizedAlbumName, err)
	}

	insertID, err := result.LastInsertId()
	if err == nil && insertID > 0 {
		return insertID, nil
	}

	albumAgain, againErr := s.deps.AlbumModel.FindOneByName(ctx, normalizedAlbumName)
	if againErr != nil {
		return 0, againErr
	}
	return albumAgain.Id, nil
}
