package sc

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"rudy_gc/internal/types"
	"strings"
	"syscall"
)

func (l *ScService) MoveScFilm(ctx context.Context, scName string) error {
	gLists, err := l.deps.GListRepo.FindGList(ctx, scName, nil, 1, 100)
	if err != nil {
		return err
	}
	if len(gLists) == 100 {
		l.deps.Log.Warn("result hits pageSize=100; there may be more, consider paging")
	}

	dir, err := l.deps.DirectoryRepo.FindOneByName(ctx, "v3")
	if err != nil {
		return err
	}

	mfs := make([]*types.Film, 0, len(gLists))
	for _, m := range gLists {
		f, err := l.deps.FilmRepo.FindOneByMovieJavId(ctx, m.MovieJavId)
		if err != nil {
			return err
		}
		if f == nil {
			l.deps.Log.Warn("film not found by javId, skip", m.MovieJavId)
			continue
		}
		if f.Dir4Id == dir.Id || strings.Contains(f.FullDir, "MARKV3") {
			l.deps.Log.Info("跳过：", f.MovieName, f.FullDir)
			continue
		}
		mfs = append(mfs, f)
	}

	for _, f := range mfs {
		var newDir string
		for _, p := range l.deps.Config.Film.Pairs {
			if f.RootDir == filepath.Clean(p.RootDir) {
				newDir = p.MoveFilmDestination
				break
			}
		}
		if newDir == "" {
			l.deps.Log.Error("no target dir mapping for root", f.RootDir, "movie", f.MovieName)
			continue
		}

		oldPath := filepath.Join(f.FullDir, f.FileName)
		newPath := filepath.Join(newDir, f.FileName)

		// 目标目录
		if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
			return err
		}

		// 只允许同卷；跨卷(EXDEV)直接报错返回
		if err := os.Rename(oldPath, newPath); err != nil {
			if errors.Is(err, syscall.EXDEV) {
				return errors.New("跨卷移动被禁止: " + oldPath + " => " + newPath)
			}
			l.deps.Log.Errorf("移动Movie错误:%v", err.Error())
			continue
		}

		l.deps.Log.Info("moved", oldPath, "=>", newPath)
	}

	return nil
}
