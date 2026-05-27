package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/movie"
	"rudy_gc/internal/types"
)

type rankPeriodSwitch struct {
	Label  string
	Href   string
	Active bool
}

// /moviecardperiodrank：按周/月/季/年周期排名显示
func (h *MovieHTMLHandler) ListMovieCardPeriodRank(c *gin.Context) {
	typeName := strings.TrimSpace(c.DefaultQuery("type", consts.RankPeriodTypeName(consts.RankPeriodTypeMonth)))
	periodType := consts.RankPeriodTypeFromName(typeName)
	category := parseInt64OrDefault(c.DefaultQuery("category", strconv.FormatInt(consts.BestCategoryMonth, 10)), consts.BestCategoryMonth)
	periodKey := strings.TrimSpace(c.Query("key"))
	page := parseInt64OrDefault(c.DefaultQuery("p", "1"), 1)
	pageSize := parseInt64OrDefault(c.DefaultQuery("ps", strconv.FormatInt(defaultPageSize, 10)), defaultPageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = defaultPageSize
	}

	view, err := h.movieSvc.BuildRankPeriodPage(c.Request.Context(), movie.RankPeriodPageRequest{
		PeriodType: periodType,
		PeriodKey:  periodKey,
		Category:   category,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			c.String(http.StatusOK, "暂无周期排行数据")
			return
		}
		c.String(http.StatusInternalServerError, "查询周期排行失败: %v", err)
		return
	}

	h.enqueueJavIDsNonBlocking(rankPeriodCardJavIDs(view.Cards))

	pi := BuildPageInfo(c, view.Total, page, pageSize, pageWindow)
	currentTypeName := consts.RankPeriodTypeName(view.Period.PeriodType)
	currentCategory := view.Period.Category

	c.HTML(http.StatusOK, "page.list_movie_card_rank_period", gin.H{
		"Title":           view.Title,
		"cards":           view.Cards,
		"total":           view.Total,
		"PageInfo":        pi,
		"pageInfo":        pi,
		"periodKey":       view.Period.PeriodKey,
		"periodTypeLabel": view.PeriodTypeLabel,
		"categoryLabel":   view.CategoryLabel,
		"rangeStart":      view.RangeStart,
		"rangeEnd":        view.RangeEnd,
		"typeLinks":       buildRankPeriodTypeLinks(c, currentTypeName, currentCategory),
		"categoryLinks":   buildRankPeriodCategoryLinks(c, currentTypeName, currentCategory),
		"latestHref":      buildRankPeriodHref(c.Request.URL.Path, currentTypeName, currentCategory, ""),
		"prevHref":        rankPeriodNavHref(c.Request.URL.Path, currentTypeName, currentCategory, view.PrevPeriod),
		"nextHref":        rankPeriodNavHref(c.Request.URL.Path, currentTypeName, currentCategory, view.NextPeriod),
		"prevDisabled":    view.PrevPeriod == nil,
		"nextDisabled":    view.NextPeriod == nil,
	})
}

func buildRankPeriodTypeLinks(c *gin.Context, currentType string, category int64) []rankPeriodSwitch {
	periodTypes := []int64{
		consts.RankPeriodTypeWeek,
		consts.RankPeriodTypeMonth,
		consts.RankPeriodTypeQuarter,
		consts.RankPeriodTypeYear,
	}
	out := make([]rankPeriodSwitch, 0, len(periodTypes))
	for _, periodType := range periodTypes {
		typeName := consts.RankPeriodTypeName(periodType)
		out = append(out, rankPeriodSwitch{
			Label:  consts.RankPeriodTypeLabel(periodType),
			Href:   buildRankPeriodHref(c.Request.URL.Path, typeName, category, ""),
			Active: typeName == currentType,
		})
	}
	return out
}

func buildRankPeriodCategoryLinks(c *gin.Context, currentType string, currentCategory int64) []rankPeriodSwitch {
	categories := []int64{
		consts.BestCategoryMonth,
		consts.BestCategoryAllTime,
	}
	out := make([]rankPeriodSwitch, 0, len(categories))
	for _, category := range categories {
		out = append(out, rankPeriodSwitch{
			Label:  consts.BestCategoryLabel(category),
			Href:   buildRankPeriodHref(c.Request.URL.Path, currentType, category, ""),
			Active: category == currentCategory,
		})
	}
	return out
}

func rankPeriodNavHref(path, typeName string, category int64, period *moviex.CRankPeriod) string {
	if period == nil {
		return ""
	}
	return buildRankPeriodHref(path, typeName, category, period.PeriodKey)
}

func buildRankPeriodHref(path, typeName string, category int64, key string) string {
	q := url.Values{}
	q.Set("type", typeName)
	q.Set("category", strconv.FormatInt(category, 10))
	if key != "" {
		q.Set("key", key)
	}
	return fmt.Sprintf("%s?%s", path, q.Encode())
}

func rankPeriodCardJavIDs(cards []*movie.RankPeriodMovieCard) []string {
	out := make([]string, 0, len(cards))
	for _, card := range cards {
		if card == nil || card.Movie == nil {
			continue
		}
		out = append(out, card.Movie.JavId)
	}
	return out
}
