package rank

import (
	"fmt"
	"net/url"

	"rudy-gc-api/internal/consts"
	"rudy-gc-api/internal/dep"
	"rudy-gc-api/internal/model/modelx"
	"rudy-gc-api/internal/types"
)

type Service struct {
	repo *modelx.MovieReadRepo
}

func New(d *dep.Dep) *Service {
	return &Service{
		repo: modelx.NewMovieReadRepo(d.Conn, d.Config),
	}
}

func buildPeriodHref(typeName string, category int64, key string) string {
	values := url.Values{}
	values.Set("type", typeName)
	values.Set("category", fmt.Sprintf("%d", category))
	if key != "" {
		values.Set("key", key)
	}
	return "/moviecardperiodrank?" + values.Encode()
}

func buildTypeLinks(currentType string, category int64) []*types.RankSwitchLink {
	periodTypes := []int64{
		consts.RankPeriodTypeWeek,
		consts.RankPeriodTypeMonth,
		consts.RankPeriodTypeQuarter,
		consts.RankPeriodTypeYear,
	}
	out := make([]*types.RankSwitchLink, 0, len(periodTypes))
	for _, periodType := range periodTypes {
		typeName := consts.RankPeriodTypeName(periodType)
		out = append(out, &types.RankSwitchLink{
			Label:  consts.RankPeriodTypeLabel(periodType),
			Href:   buildPeriodHref(typeName, category, ""),
			Active: typeName == currentType,
		})
	}
	return out
}

func buildCategoryLinks(currentType string, currentCategory int64) []*types.RankSwitchLink {
	categories := []int64{consts.BestCategoryMonth, consts.BestCategoryAllTime}
	out := make([]*types.RankSwitchLink, 0, len(categories))
	for _, category := range categories {
		out = append(out, &types.RankSwitchLink{
			Label:  consts.BestCategoryLabel(category),
			Href:   buildPeriodHref(currentType, category, ""),
			Active: category == currentCategory,
		})
	}
	return out
}
