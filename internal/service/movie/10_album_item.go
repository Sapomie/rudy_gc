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
	sourceTypeJavbusMagnet = "javbus_magnet"
	sourceTypeSukebei      = "sukebei_torrent"
)

var (
	ErrInvalidSourceType = errors.New("无效来源类型")
	ErrInvalidSourceRow  = errors.New("无效来源记录")
	ErrSourceNotFound    = errors.New("来源记录不存在")
	ErrSourceMovieMiss   = errors.New("来源记录与影片不匹配")
	ErrInvalidAlbumID    = errors.New("无效相册")
	ErrInvalidAlbumItem  = errors.New("无效相册条目")
	ErrAlbumMismatch     = errors.New("相册条目不属于当前相册")
)

type albumSourcePayload struct {
	MovieName   string
	InfoHash    string
	Size        string
	PublishTime int64
}

func (s *Service) AddFetchResourceToDefaultAlbum(ctx context.Context, movieJavID string, sourceType string, sourceRowID int64) (bool, error) {
	normalizedMovieJavID := strings.TrimSpace(movieJavID)
	if normalizedMovieJavID == "" || sourceRowID <= 0 {
		return false, ErrInvalidSourceRow
	}

	normalizedSourceType := strings.TrimSpace(sourceType)
	sourcePayload, err := s.readSourcePayload(ctx, normalizedMovieJavID, normalizedSourceType, sourceRowID)
	if err != nil {
		return false, err
	}

	albumID, err := s.getDefaultAlbumID(ctx, true)
	if err != nil {
		return false, err
	}

	_, err = s.deps.AlbumItemModel.FindOneByAlbumIdSourceTypeSourceRowId(ctx, albumID, normalizedSourceType, sourceRowID)
	if err == nil {
		return false, nil
	}
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return false, err
	}

	now := time.Now().Unix()
	_, err = s.deps.AlbumItemModel.Insert(ctx, &moviex.TmAlbumItem{
		AlbumId:     albumID,
		SourceType:  normalizedSourceType,
		SourceRowId: sourceRowID,
		MovieJavId:  normalizedMovieJavID,
		MovieName:   sourcePayload.MovieName,
		InfoHash:    sourcePayload.InfoHash,
		Size:        sourcePayload.Size,
		PublishTime: sourcePayload.PublishTime,
		CreatedOn:   now,
		UpdatedOn:   now,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) RemoveFetchResourceFromDefaultAlbum(ctx context.Context, movieJavID string, sourceType string, sourceRowID int64) (bool, error) {
	normalizedMovieJavID := strings.TrimSpace(movieJavID)
	if normalizedMovieJavID == "" || sourceRowID <= 0 {
		return false, ErrInvalidSourceRow
	}
	normalizedSourceType := strings.TrimSpace(sourceType)
	if _, err := s.readSourcePayload(ctx, normalizedMovieJavID, normalizedSourceType, sourceRowID); err != nil {
		return false, err
	}

	albumID, err := s.getDefaultAlbumID(ctx, false)
	if err != nil {
		return false, err
	}
	if albumID <= 0 {
		return false, nil
	}

	return s.deps.AlbumItemModel.DeleteByAlbumIdSourceTypeSourceRowId(ctx, albumID, normalizedSourceType, sourceRowID)
}

func (s *Service) ListDefaultAlbumItemsByMovieJavID(ctx context.Context, movieJavID string) ([]*moviex.TmAlbumItem, error) {
	normalizedMovieJavID := strings.TrimSpace(movieJavID)
	if normalizedMovieJavID == "" {
		return []*moviex.TmAlbumItem{}, nil
	}

	albumID, err := s.getDefaultAlbumID(ctx, false)
	if err != nil {
		return nil, err
	}
	if albumID <= 0 {
		return []*moviex.TmAlbumItem{}, nil
	}
	return s.deps.AlbumItemModel.ListByAlbumIdMovieJavId(ctx, albumID, normalizedMovieJavID)
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
			InfoHash:    strings.TrimSpace(row.InfoHash),
			Size:        strings.TrimSpace(row.SizeText),
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
			InfoHash:    strings.TrimSpace(row.InfoHash),
			Size:        strings.TrimSpace(row.SizeText),
			PublishTime: row.PublishTime,
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
	album, err := s.deps.AlbumModel.FindOneByName(ctx, defaultAlbumName)
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
		Name:      defaultAlbumName,
		Remark:    "默认下载中相册",
		CreatedOn: now,
		UpdatedOn: now,
	})
	if err != nil {
		albumAgain, againErr := s.deps.AlbumModel.FindOneByName(ctx, defaultAlbumName)
		if againErr == nil && albumAgain != nil {
			return albumAgain.Id, nil
		}
		return 0, fmt.Errorf("创建默认下载中相册失败: %w", err)
	}

	insertID, err := result.LastInsertId()
	if err == nil && insertID > 0 {
		return insertID, nil
	}

	albumAgain, againErr := s.deps.AlbumModel.FindOneByName(ctx, defaultAlbumName)
	if againErr != nil {
		return 0, againErr
	}
	return albumAgain.Id, nil
}
