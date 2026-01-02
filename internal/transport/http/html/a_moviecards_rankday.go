package html

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
)

// /moviecarddayrank：按某天榜单排名顺序显示
func (h *MovieHTMLHandler) ListMovieCardDayRank(c *gin.Context) {
	ctx := c.Request.Context()
	dateStr := strings.TrimSpace(c.Query("date"))

	minDay, err := h.movieSvc.FindEarliestRankDayNumber(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "查询最早排行日期失败: %v", err)
		return
	}
	maxDay, err := h.movieSvc.FindLatestRankDayNumber(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "查询最近排行日期失败: %v", err)
		return
	}
	if maxDay <= 0 {
		c.String(http.StatusOK, "暂无排行数据")
		return
	}
	if minDay <= 0 {
		minDay = 1
	}

	var dayNumber int64
	if dateStr == "" {
		dayNumber = maxDay
		dateStr = consts.GetDateStringByRankDayNumber(dayNumber)
	} else {
		dayNumber = consts.GetRankDayNumber(dateStr)
		if dayNumber < minDay {
			dayNumber = minDay
		}
		if dayNumber > maxDay {
			dayNumber = maxDay
		}
		dateStr = consts.GetDateStringByRankDayNumber(dayNumber)
	}

	page := parseInt64OrDefault(c.DefaultQuery("p", "1"), 1)
	pageSize := parseInt64OrDefault(c.DefaultQuery("ps", "0"), 0)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = defaultPageSize
	}

	movies, total, err := h.movieSvc.ListMovieTypesByRankDay(ctx, dayNumber, page, pageSize)
	if err != nil {
		c.String(http.StatusInternalServerError, "查询排行影片失败: %v", err)
		return
	}

	prevDay := dayNumber - 1
	if prevDay < minDay {
		prevDay = minDay
	}
	nextDay := dayNumber + 1
	if nextDay > maxDay {
		nextDay = maxDay
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randDay := minDay
	if maxDay > minDay {
		randDay = rng.Int63n(maxDay-minDay+1) + minDay
	}

	pi := BuildPageInfo(c, total, page, pageSize, pageWindow)

	c.HTML(http.StatusOK, "page.list_movie_card_rank_day", gin.H{
		"Title":        fmt.Sprintf("MovieCard Rank %s", dateStr),
		"movies":       movies,
		"total":        total,
		"PageInfo":     pi,
		"pageInfo":     pi,
		"rankDate":     dateStr,
		"prevDate":     consts.GetDateStringByRankDayNumber(prevDay),
		"nextDate":     consts.GetDateStringByRankDayNumber(nextDay),
		"randDate":     consts.GetDateStringByRankDayNumber(randDay),
		"prevDisabled": dayNumber <= minDay,
		"nextDisabled": dayNumber >= maxDay,
	})
}

func parseInt64OrDefault(s string, def int64) int64 {
	if s == "" {
		return def
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return n
}
