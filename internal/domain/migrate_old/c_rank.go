package migrate

import (
	"context"
	"rudy_gc/internal/types"
	"rudy_gc/oldmodel/modelx"
)

func (s *Service) MigrateRank() error {
	ctx := context.Background()
	ranksOld, err := s.xModel.RankModel.All(ctx)
	if err != nil {
		return err
	}
	ranksNew, err := s.deps.RankRepo.All(ctx)
	if err != nil {
		return err
	}

	// --- 求 ranksNeedUpsert（RankKey 差集） ---
	newSet := make(map[string]struct{}, len(ranksNew))
	for _, v := range ranksNew {
		if v.RankKey == "" {
			continue
		}
		newSet[v.RankKey] = struct{}{}
	}

	seen := make(map[string]struct{}, len(ranksOld))
	ranksNeedUpsert := make([]*modelx.CRank, 0, len(ranksOld))
	for _, v := range ranksOld {
		if v.Name == "" {
			continue
		}
		// 去重 + 差集
		if _, dup := seen[v.Name]; dup {
			continue
		}
		seen[v.Name] = struct{}{}
		if _, exists := newSet[v.Name]; !exists {
			ranksNeedUpsert = append(ranksNeedUpsert, v)
		}
	}

	s.deps.Log.Infof("共 %d 条 Rank 需要 Upsert", len(ranksNeedUpsert))

	var count int
	javIdMap := make(map[string]struct{})
	for _, r := range ranksNeedUpsert {

		rankNew := types.Rank{
			RankKey:    r.Name,
			MovieJavId: r.MovieJavId,
			DayNumber:  r.DayNumber,
			RankPos:    r.Number,
			Category:   r.Category,
		}

		err = s.deps.RankRepo.Upsert(ctx, &rankNew)
		if err != nil {
			return err
		}
		javIdMap[r.MovieJavId] = struct{}{}
		count++
		s.deps.Log.Infof("已完成%v/%v", count, len(ranksNeedUpsert))
	}

	err = s.crawlLogic.UpdateMovieRankInfo(ctx, javIdMap)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpDateAllRankInfo() error {
	ctx := context.Background()
	javIds, err := s.xModel.RankModel.DistinctJavId(ctx)
	if err != nil {
		return err
	}

	var count int
	for _, javId := range javIds {
		err = s.crawlLogic.AddRankInfo(ctx, javId)
		if err != nil {
			return err
		}

		count++
		s.deps.Log.Infof("已完成%v/%v", count, len(javIds))
	}

	return nil
}
