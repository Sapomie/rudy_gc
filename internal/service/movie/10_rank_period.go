package movie

import (
	"context"
	"errors"
	"fmt"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/taskctx"
	"rudy_gc/internal/types"
)

func (s *Service) RebuildCurrentRankPeriods(ctx context.Context) error {
	latestDay, err := s.deps.RankModel.FindLatestDayNumber(ctx)
	if err != nil {
		return err
	}
	if latestDay <= 0 {
		return nil
	}

	specs := buildCurrentRankPeriodSpecs(latestDay)
	categories := []int64{
		consts.BestCategoryMonth,
		consts.BestCategoryAllTime,
	}

	for _, spec := range specs {
		for _, category := range categories {
			if err := s.rebuildRankPeriod(ctx, spec, category); err != nil {
				return fmt.Errorf("rebuild rank period type=%s category=%d failed: %w", consts.RankPeriodTypeName(spec.PeriodType), category, err)
			}
		}
	}
	return nil
}

func (s *Service) RebuildAllRankPeriods(ctx context.Context) error {
	earliestDay, err := s.deps.RankModel.FindEarliestDayNumber(ctx)
	if err != nil {
		return err
	}
	latestDay, err := s.deps.RankModel.FindLatestDayNumber(ctx)
	if err != nil {
		return err
	}
	if earliestDay <= 0 || latestDay <= 0 || earliestDay > latestDay {
		return nil
	}

	specs := buildHistoricalRankPeriodSpecs(earliestDay, latestDay)
	categories := []int64{
		consts.BestCategoryMonth,
		consts.BestCategoryAllTime,
	}
	total := len(specs) * len(categories)
	if total == 0 {
		return nil
	}

	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:       "rank_period_prepare",
		Message:     fmt.Sprintf("开始回填周期排行，共 %d 个周期", len(specs)),
		QueuedCount: total,
	})

	handled := 0
	success := 0
	failed := 0

	for _, spec := range specs {
		for _, category := range categories {
			if err := taskctx.WaitIfPaused(ctx); err != nil {
				return err
			}

			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:             "rank_period_backfill",
				Message:           fmt.Sprintf("回填 %s %s category=%d", consts.RankPeriodTypeLabel(spec.PeriodType), spec.PeriodKey, category),
				HandledCount:      handled,
				SuccessCount:      success,
				FailedCount:       failed,
				QueuedCount:       total - handled,
				CurrentPhaseKey:   consts.RankPeriodTypeName(spec.PeriodType),
				PhaseKey:          spec.PeriodKey,
				PhaseHandledCount: handled,
				PhaseTotalCount:   total,
				PhaseSuccessCount: success,
				PhaseFailedCount:  failed,
			})

			if err := s.rebuildRankPeriod(ctx, spec, category); err != nil {
				failed++
				handled++
				taskctx.ReportProgress(ctx, taskctx.Progress{
					Stage:             "rank_period_backfill",
					Message:           fmt.Sprintf("回填失败 %s %s category=%d", consts.RankPeriodTypeLabel(spec.PeriodType), spec.PeriodKey, category),
					HandledCount:      handled,
					SuccessCount:      success,
					FailedCount:       failed,
					QueuedCount:       total - handled,
					CurrentPhaseKey:   consts.RankPeriodTypeName(spec.PeriodType),
					PhaseKey:          spec.PeriodKey,
					PhaseHandledCount: handled,
					PhaseTotalCount:   total,
					PhaseSuccessCount: success,
					PhaseFailedCount:  failed,
				})
				return fmt.Errorf("rebuild all rank periods type=%s key=%s category=%d failed: %w", consts.RankPeriodTypeName(spec.PeriodType), spec.PeriodKey, category, err)
			}

			handled++
			success++
		}
	}

	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:             "rank_period_done",
		Message:           fmt.Sprintf("周期排行回填完成，共 %d 个周期", len(specs)),
		HandledCount:      handled,
		SuccessCount:      success,
		FailedCount:       failed,
		QueuedCount:       total - handled,
		PhaseHandledCount: handled,
		PhaseTotalCount:   total,
		PhaseSuccessCount: success,
		PhaseFailedCount:  failed,
	})
	return nil
}

func (s *Service) BuildRankPeriodPage(ctx context.Context, req RankPeriodPageRequest) (*RankPeriodPage, error) {
	if req.PeriodType == 0 {
		req.PeriodType = consts.RankPeriodTypeMonth
	}
	if req.Category == 0 {
		req.Category = consts.BestCategoryMonth
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = defaultPeriodPageSize
	}
	if req.PageSize > maxPeriodPageSize {
		req.PageSize = maxPeriodPageSize
	}

	var (
		period *moviex.CRankPeriod
		err    error
	)
	if req.PeriodKey == "" {
		period, err = s.deps.RankPeriodModel.FindLatestByPeriodTypeCategory(ctx, req.PeriodType, req.Category)
	} else {
		period, err = s.deps.RankPeriodModel.FindOneByPeriodTypePeriodKeyCategory(ctx, req.PeriodType, req.PeriodKey, req.Category)
	}
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}

	total, err := s.deps.RankPeriodItemModel.CountByPeriodId(ctx, period.Id)
	if err != nil {
		return nil, err
	}

	rows, err := s.deps.RankPeriodItemModel.ListByPeriodIdPage(ctx, period.Id, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	cards := make([]*RankPeriodMovieCard, 0, len(rows))
	movies := make([]*types.MovieType, 0, len(rows))
	for _, row := range rows {
		mt, err := s.GetMovieType(ctx, row.MovieJavId)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				continue
			}
			return nil, err
		}
		card := buildRankPeriodMovieCard(mt, row)
		if card == nil {
			continue
		}
		cards = append(cards, card)
		movies = append(movies, mt)
	}

	prevPeriod, err := s.deps.RankPeriodModel.FindPrevByPeriodTypeCategory(ctx, period.PeriodType, period.Category, period.StartDayNumber)
	if err != nil {
		return nil, err
	}
	nextPeriod, err := s.deps.RankPeriodModel.FindNextByPeriodTypeCategory(ctx, period.PeriodType, period.Category, period.StartDayNumber)
	if err != nil {
		return nil, err
	}

	return &RankPeriodPage{
		Title:           fmt.Sprintf("MovieCard %s %s", consts.RankPeriodTypeLabel(period.PeriodType), period.PeriodKey),
		Period:          period,
		PrevPeriod:      prevPeriod,
		NextPeriod:      nextPeriod,
		Cards:           cards,
		Movies:          movies,
		Total:           total,
		PeriodTypeLabel: consts.RankPeriodTypeLabel(period.PeriodType),
		CategoryLabel:   consts.BestCategoryLabel(period.Category),
		RangeStart:      consts.GetDateStringByRankDayNumber(period.StartDayNumber),
		RangeEnd:        consts.GetDateStringByRankDayNumber(period.EndDayNumber),
	}, nil
}

func (s *Service) rebuildRankPeriod(ctx context.Context, spec rankPeriodSpec, category int64) error {
	rows, err := s.deps.RankModel.ListByDayRangeAndCategory(ctx, spec.StartDayNumber, spec.EndDayNumber, category)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	prevRankMap := map[string]int64{}
	prevPeriod, err := s.deps.RankPeriodModel.FindPrevByPeriodTypeCategory(ctx, spec.PeriodType, category, spec.StartDayNumber)
	if err != nil {
		return err
	}
	if prevPeriod != nil {
		prevRows, err := s.deps.RankPeriodItemModel.ListByPeriodId(ctx, prevPeriod.Id)
		if err != nil {
			return err
		}
		for _, row := range prevRows {
			prevRankMap[row.MovieJavId] = row.RankPos
		}
	}

	peakRankMap, err := s.deps.RankPeriodItemModel.FindPeakRankMapByPeriodTypeCategoryBeforeDay(ctx, spec.PeriodType, category, spec.StartDayNumber)
	if err != nil {
		return err
	}

	resolvedSpec, items := buildRankPeriodItems(spec, rows, prevRankMap, peakRankMap)
	period, err := s.upsertRankPeriod(ctx, resolvedSpec, category, consts.RankPeriodStatusProcessing)
	if err != nil {
		return err
	}

	if err := s.syncRankPeriodItems(ctx, period, items); err != nil {
		return err
	}

	period.Status = consts.RankPeriodStatusReady
	period.UpdatedOn = time.Now().Unix()
	return s.deps.RankPeriodModel.Update(ctx, period)
}

func (s *Service) upsertRankPeriod(ctx context.Context, spec rankPeriodSpec, category, status int64) (*moviex.CRankPeriod, error) {
	now := time.Now().Unix()

	row, err := s.deps.RankPeriodModel.FindOneByPeriodTypePeriodKeyCategory(ctx, spec.PeriodType, spec.PeriodKey, category)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, err
	}

	if errors.Is(err, moviex.ErrNotFound) {
		row = &moviex.CRankPeriod{
			PeriodType:     spec.PeriodType,
			PeriodKey:      spec.PeriodKey,
			Category:       category,
			StartDayNumber: spec.StartDayNumber,
			EndDayNumber:   spec.EndDayNumber,
			PickDays:       spec.PickDays,
			TopN:           spec.TopN,
			Status:         status,
			CreatedOn:      now,
			UpdatedOn:      now,
		}
		ret, err := s.deps.RankPeriodModel.Insert(ctx, row)
		if err != nil {
			return nil, err
		}
		id, err := ret.LastInsertId()
		if err != nil {
			return nil, err
		}
		row.Id = id
		return row, nil
	}

	row.StartDayNumber = spec.StartDayNumber
	row.EndDayNumber = spec.EndDayNumber
	row.PickDays = spec.PickDays
	row.TopN = spec.TopN
	row.Status = status
	row.UpdatedOn = now
	if err := s.deps.RankPeriodModel.Update(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *Service) syncRankPeriodItems(ctx context.Context, period *moviex.CRankPeriod, items []*moviex.CRankPeriodItem) error {
	existingRows, err := s.deps.RankPeriodItemModel.ListByPeriodId(ctx, period.Id)
	if err != nil {
		return err
	}

	existingMap := make(map[string]*moviex.CRankPeriodItem, len(existingRows))
	for _, row := range existingRows {
		existingMap[row.MovieJavId] = row
	}

	now := time.Now().Unix()
	for _, row := range items {
		row.PeriodId = period.Id
		row.UpdatedOn = now
		if old, ok := existingMap[row.MovieJavId]; ok {
			row.Id = old.Id
			row.CreatedOn = old.CreatedOn
			if err := s.deps.RankPeriodItemModel.Update(ctx, row); err != nil {
				return err
			}
			delete(existingMap, row.MovieJavId)
			continue
		}

		row.CreatedOn = now
		if _, err := s.deps.RankPeriodItemModel.Insert(ctx, row); err != nil {
			return err
		}
	}

	for _, stale := range existingMap {
		if err := s.deps.RankPeriodItemModel.Delete(ctx, stale.Id); err != nil {
			return err
		}
	}

	return nil
}
