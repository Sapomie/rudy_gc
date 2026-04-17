package spider

import (
	"context"
	"fmt"
	"strings"
)

const sehuatangAlbumSourceType = "sehuatang_magnet"

func (l *CrawlLogic) syncSehuatangMovieRefsAfterMovieSaved(ctx context.Context, movieName string, movieJavID string, now int64) error {
	movieName = strings.TrimSpace(movieName)
	movieJavID = strings.TrimSpace(movieJavID)
	if movieName == "" || movieJavID == "" {
		return nil
	}

	missingRows, err := l.deps.SehuatangMagnetModel.ListMissingMovieJavIDByMovieName(ctx, movieName)
	if err != nil {
		return fmt.Errorf("list missing sehuatang movie_jav_id failed: %w", err)
	}
	for _, row := range missingRows {
		if row == nil || strings.TrimSpace(row.MovieJavId) != "" {
			continue
		}
		row.MovieJavId = movieJavID
		row.UpdatedOn = now
		if err := l.deps.SehuatangMagnetModel.Update(ctx, row); err != nil {
			return fmt.Errorf("update sehuatang movie_jav_id failed, id=%d: %w", row.Id, err)
		}
	}

	sourceRows, err := l.deps.SehuatangMagnetModel.ListByMovieKey(ctx, movieJavID, movieName)
	if err != nil {
		return fmt.Errorf("list sehuatang rows for album repair failed: %w", err)
	}
	sourceRowIDs := make([]int64, 0, len(sourceRows))
	seen := make(map[int64]struct{}, len(sourceRows))
	for _, row := range sourceRows {
		if row == nil || row.Id <= 0 {
			continue
		}
		if _, ok := seen[row.Id]; ok {
			continue
		}
		seen[row.Id] = struct{}{}
		sourceRowIDs = append(sourceRowIDs, row.Id)
	}
	if len(sourceRowIDs) == 0 {
		return nil
	}

	albumItems, err := l.deps.AlbumItemModel.ListMissingMovieJavIDBySourceRows(ctx, sehuatangAlbumSourceType, sourceRowIDs)
	if err != nil {
		return fmt.Errorf("list missing album item movie_jav_id failed: %w", err)
	}
	for _, item := range albumItems {
		if item == nil || strings.TrimSpace(item.MovieJavId) != "" {
			continue
		}
		item.MovieJavId = movieJavID
		if strings.TrimSpace(item.MovieName) == "" {
			item.MovieName = movieName
		}
		item.UpdatedOn = now
		if err := l.deps.AlbumItemModel.Update(ctx, item); err != nil {
			return fmt.Errorf("update album item movie_jav_id failed, id=%d: %w", item.Id, err)
		}
	}

	return nil
}
