package migrate

import (
	"context"
	"rudy_gc/internal/types"
	"rudy_gc/pkg/ptr"
	"time"
)

func (s *Service) MigrateTranslationAndNeedDown() error {
	ctx := context.Background()
	all, err := s.xModel.MovieModel.FindMoviesAll(ctx, 1000000)
	if err != nil {
		return err
	}

	var count int
	for _, m := range all {
		now := time.Now().Unix()
		p := types.MinfoPatch{
			Chinese:      ptr.String(m.Chinese),
			NeedDownload: ptr.Int64(m.NeedDownload),
			UpdatedOn:    &now,
		}

		err := s.deps.MinfoRepo.UpdatePartialByJavId(ctx, m.JavId, p)
		if err != nil {
			s.deps.Log.Errorf("不存在%v___%v", m.Name, m.JavId)
			continue
		}
		s.movieSvc.InvalidateMovieType(ctx, m.JavId)

		itp := types.ItemPatch{
			HasChinese: ptr.Int64(m.HasChinese),
			UpdatedOn:  &now,
		}

		err = s.deps.ItemRepo.UpdatePartialByJavId(ctx, m.JavId, itp)
		if err != nil {
			s.deps.Log.Errorf("不存在%v___%v", m.Name, m.JavId)
			continue
		}

		count++
		s.deps.Log.Infof("已完成%v/%v", count, len(all))
	}

	return nil
}
