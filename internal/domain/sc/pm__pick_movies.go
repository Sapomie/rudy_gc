package sc

import (
	"context"
	"fmt"
	"path/filepath"
	"rudy_gc/internal/types"
	"rudy_gc/pkg/filetool"
	"rudy_gc/pkg/ptr"
	"sort"
	"strings"
	"time"
)

func (l *ScService) PickProcession() error {
	ctx := context.Background()

	reqs := []*requestWithWeight{
		//{
		//	req: &types.ListMovieFullRequest{Page: 1, PageSize: 10000, Owned: 3,
		//		ScTimesMin: 1,
		//		ScTimesMax: ptr.Int64(1),
		//		//ReleasingDateEnd: "2025-01-01",
		//		//FilmBirthTimeEnd: "2025-10-01",
		//	},
		//	w: 2,
		//},
		{
			req: &types.ListMovieFullRequest{Page: 1, PageSize: 10000, Owned: 3,
				ScTimesMax: ptr.Int64(0),
				//ReleasingDateStart: "2025-06-01",
				FilmBirthTimeEnd: "2025-10-01",
			},
			w: 10,
		},
		{
			req: &types.ListMovieFullRequest{Page: 1, PageSize: 10000, Owned: 3,
				ScTimesMax: ptr.Int64(0),
				//ReleasingDateStart: "2025-06-01",
				FilmBirthTimeStart: "2025-10-01",
				FilmBirthTimeEnd:   "2025-11-01",
				//ReleasingDateStart: "2025-10-01",
				//ViewWatchedMin:     100,
			},
			w: 12,
		},
		{
			req: &types.ListMovieFullRequest{Page: 1, PageSize: 10000, Owned: 3,
				ScTimesMax: ptr.Int64(0),
				//ReleasingDateStart: "2025-06-01",
				FilmBirthTimeStart: "2025-11-01",
				//ReleasingDateStart: "2025-10-01",
				//ViewWatchedMin:     100,
			},
			w: 12,
		},
	}

	// 例如抽取 20 个
	movieTypes, err := l.PickFromSources(ctx, reqs, 2)
	if err != nil {
		return err
	}

	l.LogPicks(movieTypes)

	err = l.copyMovieRank(movieTypes)
	if err != nil {
		return err
	}

	return nil
}

func (l *ScService) copyMovieRank(mfs []*types.MovieType) error {
	var count int
	for _, mf := range mfs {
		if err := l.copyFileToDestination(mf.VideoUrl); err != nil {
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
	if len(mts) == 0 {
		l.deps.Log.Info("没有可打印的 picks")
		return
	}

	// ---------- 排序 ----------
	sort.Slice(mts, func(i, j int) bool {
		ti := parseDate(mts[i].FilmBirthDate)
		tj := parseDate(mts[j].FilmBirthDate)
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

		filmDate := mt.FilmBirthDate
		releasingDate := mt.ReleasingDate
		scTimes := mt.ScTimes

		// 累加 Size
		if mt.VFilm != nil {
			totalSize += mt.VFilm.Size
		}

		l.deps.Log.Infof("%-25s | %-15s | %-12s | %-8d | %-12s | %-19s",
			name, cast0, filmDate, scTimes, releasingDate, castLastSc)
	}

	// ---------- 打印总大小 ----------
	gb := float64(totalSize) / (1024 * 1024 * 1024)
	l.deps.Log.Infof("总文件大小: %.2f GB", gb)
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
