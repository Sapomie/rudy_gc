package media

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
)

type ScNativeBackfillResult struct {
	MissingCount  int `json:"missing_count"`
	InsertedCount int `json:"inserted_count"`
}

func (s *Service) BackfillMissingScNativeMedia(ctx context.Context) (*ScNativeBackfillResult, error) {
	rows, err := s.deps.WMediaModel.ListScMissingNativeMedia(ctx)
	if err != nil {
		return nil, err
	}

	result := &ScNativeBackfillResult{
		MissingCount: len(rows),
	}
	if len(rows) == 0 {
		return result, nil
	}

	now := time.Now().Unix()
	dirtyJavIDs := make([]string, 0, len(rows))

	for _, item := range rows {
		if item == nil {
			continue
		}

		row := buildMissingScNativeMediaPlaceholder(item, now)
		if _, err := s.deps.WMediaModel.Insert(ctx, row); err != nil {
			if isDuplicateEntryErr(err) {
				existing, findErr := s.deps.WMediaModel.FindOneByMovieJavIdSourceType(ctx, row.MovieJavId, consts.WMediaSourceNative)
				if findErr == nil && existing != nil {
					continue
				}
			}
			return result, fmt.Errorf("insert native w_media failed, movie_jav_id=%s: %w", row.MovieJavId, err)
		}

		dirtyJavIDs = append(dirtyJavIDs, row.MovieJavId)
		result.InsertedCount++
	}

	if result.InsertedCount == 0 {
		return result, nil
	}
	if err := s.syncPersonStatsByMovieJavIDs(ctx, now, dirtyJavIDs...); err != nil {
		return result, err
	}

	s.movieSvc.EnqueueAggRebuildByMovieJavIDs("sc_native_backfill", dirtyJavIDs...)
	s.invalidateMovieTypeCaches(ctx, dirtyJavIDs...)
	return result, nil
}

func buildMissingScNativeMediaPlaceholder(item *moviex.MissingScNativeMediaRow, now int64) *moviex.WMedia {
	movieJavID := strings.TrimSpace(item.MovieJavId)
	movieName := strings.TrimSpace(item.MovieName)
	if movieName == "" {
		movieName = movieJavID
	}

	birthTime := item.MediaBirthTime
	if birthTime < 0 {
		birthTime = 0
	}
	releasingDate := item.ReleasingDate
	if releasingDate < 0 {
		releasingDate = 0
	}

	return &moviex.WMedia{
		MovieJavId:        movieJavID,
		MovieName:         movieName,
		FileName:          buildScNativeBackfillFileName(movieName, movieJavID),
		SourceType:        consts.WMediaSourceNative,
		SourceTorrentHash: buildScNativeBackfillSourceHash(movieJavID),
		DirectoryId:       0,
		RootDir:           "",
		FullDir:           "",
		Alias:             "sc_native_backfill_" + movieJavID,
		Size:              0,
		Width:             0,
		Height:            0,
		BitRate:           0,
		Duration:          0,
		FrameAverage:      0,
		HasSub:            consts.FilmNoSub,
		SelfMake:          consts.FilmNoSelfMake,
		HasMask:           consts.FilmNotErased,
		NeedScanMeta:      consts.FilmMetaDataNoNeedScan,
		IsRemoved:         consts.FilmIsRemoved,
		RemoveTime:        now,
		BirthTime:         birthTime,
		ReleasingDate:     releasingDate,
		CreatedOn:         now,
		UpdatedOn:         now,
	}
}

func buildScNativeBackfillFileName(movieName, movieJavID string) string {
	safeName := strings.TrimSpace(movieName)
	if safeName == "" {
		safeName = strings.TrimSpace(movieJavID)
	}
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	safeName = strings.TrimSpace(replacer.Replace(safeName))
	if safeName == "" {
		safeName = "sc_removed"
	}
	if len(safeName) > 180 {
		safeName = safeName[:180]
	}
	return safeName + "__sc_removed__" + strings.TrimSpace(movieJavID) + ".mp4"
}

func buildScNativeBackfillSourceHash(movieJavID string) string {
	sum := sha1.Sum([]byte("sc-native-backfill:" + strings.TrimSpace(movieJavID)))
	return hex.EncodeToString(sum[:])
}
