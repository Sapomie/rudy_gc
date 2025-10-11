package sc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"strings"
	"time"
)

func (l *ScService) AddSc(ctx context.Context, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	scName := filepath.Base(dir)
	if len(scName) != 16 {
		l.deps.Log.Error("Invalid directory name length")
		return fmt.Errorf("directory name %s does not have the expected length of 16 characters", scName)
	}

	scTime, err := getScTime(scName)
	if err != nil {
		return fmt.Errorf("failed to parse SC time from %s: %w", scName, err)
	}

	var (
		count         int64
		comeMovie     string
		movieJavIdMap = make(map[string]struct{})
	)

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed to get info for entry %s: %w", entry.Name(), err)
		}

		if !isValidFile(info) {
			continue // 过滤不符合条件的文件
		}

		movieName := extractMovieName(info.Name())
		vf, err := l.deps.FilmRepo.FindOneByMovieName(ctx, movieName)
		if err != nil {
			return fmt.Errorf("failed to find film by name %s: %w", movieName, err)
		}
		movieJavIdMap[vf.MovieJavId] = struct{}{}

		gl := createGList(scName, movieName, vf.MovieJavId, info.Name())
		if gl.IsCome == consts.GListIsCome {
			comeMovie = vf.MovieName
		}
		_, err = l.deps.GListRepo.Upsert(ctx, gl)
		if err != nil {
			return fmt.Errorf("failed to upsert glist: %w", err)
		}
		count++
	}

	sc := &types.GSc{
		Name:          scName,
		ScTime:        scTime,
		ComeMovieName: comeMovie,
		MovieNumber:   count,
	}

	_, err = l.deps.ScRepo.Upsert(ctx, sc)
	if err != nil {
		return fmt.Errorf("failed to upsert sc: %w", err)
	}

	err = l.AddMovieAndCastScInfo(ctx, movieJavIdMap)
	if err != nil {
		return fmt.Errorf("failed to add movie and cast sc: %w", err)
	}

	return nil
}

func createGList(scName, movieName, movieJavId, fileName string) *types.GList {
	isCome := consts.GListIsNotCome
	if hasComeMark(fileName) {
		isCome = consts.GListIsCome
	}
	return &types.GList{
		Name:       fmt.Sprintf("%s__%s", scName, movieName),
		ScName:     scName,
		MovieJavId: movieJavId,
		IsCome:     isCome,
	}
}

func getScTime(scName string) (int64, error) {
	// 定义时间布局，Go 的时间布局以示例时间为基准
	const layout = "2006-01-02-15-04"

	t, err := time.ParseInLocation(layout, scName, time.Local)
	if err != nil {
		return 0, fmt.Errorf("error parsing time from %s: %w", scName, err)
	}

	return t.Unix(), nil
}

func isValidFile(info os.FileInfo) bool {
	return strings.HasSuffix(info.Name(), ".mp4") && info.Size() >= 20000
}

func hasComeMark(fullName string) bool {
	if strings.Contains(fullName, "__come__") {
		return true
	} else {
		return false
	}
}

func extractMovieName(fileName string) string {
	strs := strings.Split(fileName, "_")
	return strs[0]
}
