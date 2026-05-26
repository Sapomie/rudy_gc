package rank

import (
	"context"
	"fmt"

	"rudy-gc-api/internal/consts"
	"rudy-gc-api/internal/types"
)

func (s *Service) Period(ctx context.Context, typeName string, category, page, pageSize int64, key string) (*types.RankPeriodResponse, error) {
	if typeName == "" {
		typeName = consts.RankPeriodTypeName(consts.RankPeriodTypeMonth)
	}
	if category == 0 {
		category = consts.BestCategoryMonth
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 18
	}

	periodType := consts.RankPeriodTypeFromName(typeName)
	period, err := s.repo.LoadRankPeriod(ctx, periodType, category, key)
	if err != nil {
		return nil, err
	}

	rows, total, err := s.repo.LoadRankPeriodItems(ctx, period.Id, page, pageSize)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		ids = append(ids, row.MovieJavId)
	}
	cards, err := s.repo.LoadMovieCardsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	cardMap := make(map[string]*types.MovieCard, len(cards))
	for _, card := range cards {
		if card == nil {
			continue
		}
		cardMap[card.MovieJavID] = card
	}

	items := make([]*types.RankPeriodCard, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		card := cardMap[row.MovieJavId]
		if card == nil {
			continue
		}
		item := &types.RankPeriodCard{
			Movie:           card,
			RankPos:         row.RankPos,
			Score:           row.Score,
			DaysInRank:      row.DaysInRank,
			UsedPickDays:    row.UsedPickDays,
			BestRank:        row.BestRank,
			WorstPickedRank: row.WorstPickedRank,
			PrevRank:        row.PrevRank,
			PeakRank:        row.PeakRank,
			PrevRankText:    "--",
		}
		switch {
		case row.PrevRank <= 0:
			item.RankChangeText = "新上榜"
			item.RankChangeClass = "primary"
		case row.RankChange > 0:
			item.RankChange = row.RankChange
			item.RankChangeText = fmt.Sprintf("↑ %d", row.RankChange)
			item.RankChangeClass = "success"
			item.PrevRankText = fmt.Sprintf("#%d", row.PrevRank)
		case row.RankChange < 0:
			item.RankChange = row.RankChange
			item.RankChangeText = fmt.Sprintf("↓ %d", -row.RankChange)
			item.RankChangeClass = "danger"
			item.PrevRankText = fmt.Sprintf("#%d", row.PrevRank)
		default:
			item.RankChangeText = "持平"
			item.RankChangeClass = "muted"
			item.PrevRankText = fmt.Sprintf("#%d", row.PrevRank)
		}
		items = append(items, item)
	}

	prevPeriod, err := s.repo.LoadPrevRankPeriod(ctx, period.PeriodType, period.Category, period.StartDayNumber)
	if err != nil {
		return nil, err
	}
	nextPeriod, err := s.repo.LoadNextRankPeriod(ctx, period.PeriodType, period.Category, period.StartDayNumber)
	if err != nil {
		return nil, err
	}

	return &types.RankPeriodResponse{
		Title:           fmt.Sprintf("MovieCard %s %s", consts.RankPeriodTypeLabel(period.PeriodType), period.PeriodKey),
		PeriodKey:       period.PeriodKey,
		PeriodType:      typeName,
		PeriodTypeLabel: consts.RankPeriodTypeLabel(period.PeriodType),
		Category:        period.Category,
		CategoryLabel:   consts.BestCategoryLabel(period.Category),
		RangeStart:      consts.GetDateStringByRankDayNumber(period.StartDayNumber),
		RangeEnd:        consts.GetDateStringByRankDayNumber(period.EndDayNumber),
		Items:           items,
		Total:           total,
		Page:            page,
		PageSize:        pageSize,
		PrevDisabled:    prevPeriod == nil,
		NextDisabled:    nextPeriod == nil,
		TypeLinks:       buildTypeLinks(typeName, period.Category),
		CategoryLinks:   buildCategoryLinks(typeName, period.Category),
		LatestHref:      buildPeriodHref(typeName, period.Category, ""),
	}, nil
}
