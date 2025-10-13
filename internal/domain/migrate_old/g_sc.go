package migrate

import (
	"context"
	"errors"
	"rudy_gc/internal/domain/sc"
	"rudy_gc/internal/types"
	"rudy_gc/pkg/convert"
	"time"
)

func (s *Service) MigrateSc() error {
	ctx := context.Background()
	allScs, err := s.xModel.GScModel.FindAll(ctx)
	if err != nil {
		return err
	}

	last := time.Date(2025, 1, 19, 22, 30, 0, 0, time.Local).Unix()
	for _, old := range allScs {
		now := time.Now().Unix()
		var coolDown int64
		coolDown = old.ScTime - last
		last = old.ScTime
		scNew := types.GSc{
			Name:          old.Name,
			MovieNumber:   old.MovieNumber,
			ScTime:        old.ScTime,
			ComeMovieName: old.ComeMovieName,
			Cooldown:      coolDown,
			CreatedOn:     now,
			UpdatedOn:     now,
		}
		_, err = s.deps.ScRepo.Upsert(ctx, &scNew)
		if err != nil {
			return errors.New("sc upsert err:" + err.Error())
		}

		tt := convert.FloatTo(float64(coolDown) / 60 / 60 / 24).Decimal(2)
		s.deps.Log.Info(old.Name, "----", tt)
	}

	return nil
}

func (s *Service) MigrateGlist() error {
	ctx := context.Background()
	allGlist, err := s.xModel.GListModel.FindAll(ctx)
	if err != nil {
		return err
	}

	var count int
	for _, old := range allGlist {
		scNew := types.GList{
			Name:       old.Name,
			ScName:     old.ScName,
			MovieJavId: old.MovieJavId,
			IsCome:     old.IsCome,
			CreatedOn:  old.CreatedOn,
			UpdatedOn:  old.UpdatedOn,
		}
		_, err = s.deps.GListRepo.Upsert(ctx, &scNew)
		if err != nil {
			return errors.New("sc upsert err:" + err.Error())
		}
		count++
		s.deps.Log.Infof("%v/%v", count, len(allGlist))
	}

	return nil
}

func (s *Service) AddScInfoToMinfo() error {
	ctx := context.Background()
	gLists, err := s.deps.GListRepo.FindAll(ctx)
	if err != nil {
		return err
	}
	javIdMap := make(map[string]struct{})
	for _, gList := range gLists {
		javIdMap[gList.MovieJavId] = struct{}{}
	}
	err = sc.NewScService(s.deps).AddMovieAndCastScInfo(ctx, javIdMap)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) FindDiff() error {
	ctx := context.Background()

	// 旧数据（来自 xModel）
	oldFilms, err := s.xModel.FilmModel.FindAll(ctx, 0)
	if err != nil {
		return err
	}

	// 新数据（来自 FilmRepo）
	newFilms, err := s.deps.FilmRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	// 构建两个集合（去重）
	oldSet := make(map[string]struct{}, len(oldFilms))
	for _, f := range oldFilms {
		if f.Name != "" {
			oldSet[f.Name] = struct{}{}
		}
	}

	newSet := make(map[string]struct{}, len(newFilms))
	for _, f := range newFilms {
		if f.MovieName != "" {
			newSet[f.MovieName] = struct{}{}
		}
	}

	// 计算差集
	var onlyInOld, onlyInNew []string

	// 旧库有，但新库没有
	for name := range oldSet {
		if _, ok := newSet[name]; !ok {
			onlyInOld = append(onlyInOld, name)
		}
	}

	// 新库有，但旧库没有
	for name := range newSet {
		if _, ok := oldSet[name]; !ok {
			onlyInNew = append(onlyInNew, name)
		}
	}

	// 输出结果（或进一步处理）
	s.deps.Log.Infof("旧库独有 %d 条", len(onlyInOld))
	s.deps.Log.Infof("新库独有 %d 条", len(onlyInNew))

	// （示例）打印前10条差异
	s.deps.Log.Infof("旧库独有示例: %v", preview(onlyInOld, 10))
	s.deps.Log.Infof("新库独有示例: %v", preview(onlyInNew, 10))

	return nil
}

// 仅展示前 n 条，避免日志太长
func preview(ss []string, n int) []string {
	if len(ss) <= n {
		return ss
	}
	return ss[:n]
}
