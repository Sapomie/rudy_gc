package migrate

import (
	"context"
	"errors"
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
