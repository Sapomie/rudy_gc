package spider

import (
	"context"
	"fmt"
	"strings"

	moviedomain "rudy_gc/local/api/internal/types/movie"
)

const sehuatangAlbumSourceType = "sh_magnet"

func (l *Service) syncSehuatangMovieRefsAfterMovieSaved(ctx context.Context, movieName string, movieJavID string, now int64) error {
	movieName = strings.TrimSpace(movieName)
	movieJavID = strings.TrimSpace(movieJavID)
	if movieName == "" || movieJavID == "" {
		return nil
	}
	matches, err := l.deps.MovieModel.FindMoviesByName(ctx, movieName)
	if err != nil {
		return fmt.Errorf("find sehuatang movie references failed: %w", err)
	}
	if len(matches) != 1 || matches[0] == nil {
		return nil
	}
	movieJavID = strings.TrimSpace(matches[0].JavId)
	if movieJavID == "" {
		return nil
	}

	missingRows, err := l.deps.ShMagnetModel.ListMissingMovieJavIDByMovieName(ctx, movieName)
	if err != nil {
		return fmt.Errorf("list missing sehuatang movie_jav_id failed: %w", err)
	}
	for _, row := range missingRows {
		if row == nil || strings.TrimSpace(row.MovieJavId) != "" {
			continue
		}
		if err := l.deps.ShMagnetModel.UpdateMovieFields(ctx, row.Id, movieJavID, movieName, now); err != nil {
			return fmt.Errorf("update sehuatang movie_jav_id failed, id=%d: %w", row.Id, err)
		}
	}
	if _, err := l.deps.ShSehuatangThreadModel.UpdateJAVMovieFieldsByMovieName(ctx, movieName, movieJavID, now); err != nil {
		return fmt.Errorf("update sehuatang thread movie fields failed: %w", err)
	}

	sourceRows, err := l.deps.ShMagnetModel.ListByMovieKey(ctx, movieJavID, movieName)
	if err != nil {
		return fmt.Errorf("list sehuatang rows for album repair failed: %w", err)
	}
	sourceRowIDs := make([]int64, 0, len(sourceRows))
	seen := make(map[int64]struct{}, len(sourceRows))
	for _, row := range sourceRows {
		if row == nil || row.Id <= 0 {
			continue
		}
		if strings.TrimSpace(row.MovieType) != moviedomain.SehuatangThreadMovieTypeJAV {
			if err := l.deps.ShMagnetModel.UpdateMovieFields(ctx, row.Id, movieJavID, movieName, now); err != nil {
				return fmt.Errorf("update sehuatang movie type failed, id=%d: %w", row.Id, err)
			}
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
