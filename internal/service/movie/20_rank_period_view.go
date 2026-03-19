package movie

import (
	"fmt"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/types"
)

func buildRankPeriodMovieCard(movie *types.MovieType, item *moviex.CRankPeriodItem) *RankPeriodMovieCard {
	if movie == nil || item == nil {
		return nil
	}

	card := &RankPeriodMovieCard{
		Movie:        movie,
		Item:         item,
		PrevRankText: "--",
	}

	if item.PrevRank > 0 {
		card.PrevRankText = fmt.Sprintf("#%d", item.PrevRank)
	}

	switch {
	case item.PrevRank <= 0:
		card.RankChangeText = "新上榜"
		card.RankChangeClass = "bg-primary"
	case item.RankChange > 0:
		card.RankChangeText = fmt.Sprintf("↑ %d", item.RankChange)
		card.RankChangeClass = "bg-success"
	case item.RankChange < 0:
		card.RankChangeText = fmt.Sprintf("↓ %d", -item.RankChange)
		card.RankChangeClass = "bg-danger"
	default:
		card.RankChangeText = "持平"
		card.RankChangeClass = "bg-secondary"
	}

	return card
}
