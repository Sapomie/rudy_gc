package sc

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"rudy_gc/internal/types"
	"rudy_gc/pkg/filetool"
	"rudy_gc/pkg/ptr"
	"sort"
	"strings"
	"time"
)

type PickRequestWithWeight struct {
	Req    types.ListMovieFullRequest `json:"req"`
	Weight int64                      `json:"weight"`
}

func (l *ScService) PickProcession() error {
	ctx := context.Background()

	reqs := []*requestWithWeight{

		{
			req: &types.ListMovieFullRequest{Page: 1, PageSize: 10000, MediaOwned: 3,
				ScTimesMax:          ptr.Int64(0),
				MediaBirthTimeStart: "2025-10-01",
				MediaBirthTimeEnd:   "2025-11-11",
				MediaDir4:           "vx",
			},
			w: 12,
		},
		{
			req: &types.ListMovieFullRequest{Page: 1, PageSize: 10000, MediaOwned: 3,
				ScTimesMax: ptr.Int64(0),

				MediaBirthTimeStart: "2025-12-01",
				MediaDir4:           "vx",
			},
			w: 12,
		},
	}

	// 例如抽取 20 个
	movieTypes, err := l.PickFromSources(ctx, reqs, 25)
	if err != nil {
		return err
	}

	l.LogPicksBySource(movieTypes, SmartPickSourceWMedia)

	err = l.copyMovieRank(movieTypes, SmartPickSourceWMedia)
	if err != nil {
		return err
	}

	return nil
}

func (l *ScService) PickFromRequests(ctx context.Context, reqs []PickRequestWithWeight, n int) ([]*types.MovieType, error) {
	if len(reqs) == 0 {
		return nil, errors.New("reqs is empty")
	}
	if n <= 0 {
		n = 25
	}

	converted := make([]*requestWithWeight, 0, len(reqs))
	for i := range reqs {
		req := reqs[i].Req
		converted = append(converted, &requestWithWeight{
			req: &req,
			w:   reqs[i].Weight,
		})
	}

	movieTypes, err := l.PickFromSources(ctx, converted, n)
	if err != nil {
		return nil, err
	}

	l.LogPicksBySource(movieTypes, SmartPickSourceWMedia)
	return movieTypes, nil
}

func (l *ScService) copyMovieRank(mfs []*types.MovieType, source string) error {
	var count int
	for _, mf := range mfs {
		videoURL := SmartPickMovieVideoURL(mf, source)
		if videoURL == "" {
			continue
		}
		if err := l.copyFileToDestination(videoURL); err != nil {
			l.deps.Log.Error("copy err: ", err)
		}
		count++
		l.deps.Log.Infof("%v/%v", count, len(mfs))
	}
	return nil
}

func (l *ScService) copyFileToDestination(srcFilePath string) error {
	destFilePath := filepath.Join(l.deps.Config.Film.CopyDestinationPath, filepath.Base(srcFilePath))
	if err := filetool.CopyFileWithProgress(srcFilePath, destFilePath); err != nil {
		return fmt.Errorf("failed to copy from %s to %s: %w", srcFilePath, destFilePath, err)
	}
	return nil
}
func (l *ScService) LogPicks(mts []*types.MovieType) {
	l.LogPicksBySource(mts, SmartPickSourceWMedia)
}

func (l *ScService) LogPicksBySource(mts []*types.MovieType, source string) {
	if len(mts) == 0 {
		l.deps.Log.Info("没有可打印的 picks")
		return
	}

	// ---------- 排序 ----------
	sort.Slice(mts, func(i, j int) bool {
		ti := parseDate(SmartPickMovieBirthDate(mts[i], source))
		tj := parseDate(SmartPickMovieBirthDate(mts[j], source))
		return ti.Before(tj)
	})

	// ---------- 打印表头 ----------
	l.deps.Log.Infof("%-25s | %-15s | %-12s | %-8s | %-12s | %-19s",
		"Name", "Cast[0]", "FilmDate", "ScTimes", "RDate", "CastLastSc")
	l.deps.Log.Info(strings.Repeat("-", 100))

	// ---------- 打印每行 ----------
	var totalSize int64
	for _, mt := range mts {
		if mt == nil {
			continue
		}

		name := mt.Name
		if len(name) > 25 {
			name = name[:22] + "..."
		}

		cast0 := "-"
		castLastSc := "-"
		if len(mt.Cast) > 0 && mt.Cast[0] != nil {
			if mt.Cast[0].Name != "" {
				cast0 = mt.Cast[0].Name
			}
			if mt.Cast[0].LastScTime > 0 {
				castLastSc = time.Unix(mt.Cast[0].LastScTime, 0).Format("2006-01-02 15:04:05")
			}
		}

		filmDate := SmartPickMovieBirthDate(mt, source)
		releasingDate := mt.ReleasingDate
		scTimes := mt.ScTimes

		// 累加 Size
		totalSize += SmartPickMovieSize(mt, source)

		l.deps.Log.Infof("%-25s | %-15s | %-12s | %-8d | %-12s | %-19s",
			name, cast0, filmDate, scTimes, releasingDate, castLastSc)
	}

	// ---------- 打印总大小 ----------
	gb := float64(totalSize) / (1024 * 1024 * 1024)
	l.deps.Log.Infof("总文件大小: %.2f GB", gb)
}

func SmartPickMovieBirthDate(mt *types.MovieType, source string) string {
	if mt == nil {
		return ""
	}
	return mt.FilmBirthDateWMedia
}

func SmartPickMovieVideoURL(mt *types.MovieType, source string) string {
	if mt == nil {
		return ""
	}
	return mt.VideoUrlWMedia
}

func SmartPickMovieOwned(mt *types.MovieType, source string) int64 {
	if mt == nil {
		return 0
	}
	return mt.OwnedWMedia
}

func SmartPickMovieSize(mt *types.MovieType, source string) int64 {
	if mt == nil {
		return 0
	}
	if mt.WMedia != nil {
		return mt.WMedia.Size
	}
	return 0
}

// 辅助：解析日期字符串（支持 "2006-01-02" / "20060102"）
func parseDate(s string) time.Time {
	layouts := []string{"2006-01-02", "20060102"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	// 无法解析时返回“最大时间”，确保排在最后
	return time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
}
