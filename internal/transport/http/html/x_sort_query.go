package html

import (
	"net/url"

	"rudy_gc/internal/consts"

	"github.com/gin-gonic/gin"
)

// ---- 允许的排序键（对应 od 参数）----
var allowedOrderBy = map[string]struct{}{
	consts.OrderByDetailUpdateTime: {}, // ✅ 按 detail 更新时间
	consts.OrderByCastAgeDesc:      {},
	consts.OrderByCastAgeAsc:       {},
	consts.OrderByViewerWatched:    {},
	consts.OrderByReleasingDate:    {},
	consts.OrderByRankDate:         {},
	consts.OrderByHighestRank:      {},
	consts.OrderByDaysInRank:       {},
	consts.OrderByBirthTime:        {},
	consts.OrderByScTimes:          {},
	consts.OrderByComeTimes:        {},
	consts.OrderByLastScTime:       {},
}

// 归一化 od 参数，非法则回退为 fallback
func normalizeOrderBy(od string, fallback string) string {
	if _, ok := allowedOrderBy[od]; ok {
		return od
	}
	return fallback
}

// SortLink / SortQuery 结构体
type SortLink struct {
	Href   string
	Active bool
}

type SortQuery struct {
	ByDetailUpdate SortLink // du  按 detail 更新时间
	ByReleasing    SortLink // rd  上映日期
	ByViewer       SortLink // vw  观看人数
	ByDaysInRank   SortLink // drk 在榜天数
	ByHighestRank  SortLink // hrk 最高排名
	ByRankDate     SortLink // rk  首次上榜日期
	ByScTimes      SortLink // sc  评分次数
	ByLastScTime   SortLink // lsct 最后评分时间
	ByBirthTime    SortLink // bt  拍摄时间
	ByCastAgeDesc  SortLink // cad 年龄倒序
	ByCastAgeAsc   SortLink // caa 年龄正序
	ByComeTimes    SortLink // co  出现次数
}

// 构造带 od 的链接（保留其它参数，并把 p 重置为 1）
func buildSortQuery(c *gin.Context, currentOD string) *SortQuery {
	makeHref := func(od string) (href string, active bool) {
		q := cloneValues(c)
		q.Set("od", od)
		q.Set("p", "1")
		href = c.Request.URL.Path
		if enc := q.Encode(); enc != "" {
			href += "?" + enc
		}
		return href, currentOD == od
	}

	du, a0 := makeHref(consts.OrderByDetailUpdateTime)
	rd, a1 := makeHref(consts.OrderByReleasingDate)
	vw, a2 := makeHref(consts.OrderByViewerWatched)
	drk, a3 := makeHref(consts.OrderByDaysInRank)
	hrk, a4 := makeHref(consts.OrderByHighestRank)
	rk, a5 := makeHref(consts.OrderByRankDate)
	sc, a6 := makeHref(consts.OrderByScTimes)
	lsct, a7 := makeHref(consts.OrderByLastScTime)
	bt, a8 := makeHref(consts.OrderByBirthTime)
	cad, a9 := makeHref(consts.OrderByCastAgeDesc)
	caa, a10 := makeHref(consts.OrderByCastAgeAsc)
	co, a11 := makeHref(consts.OrderByComeTimes)

	return &SortQuery{
		ByDetailUpdate: SortLink{Href: du, Active: a0},
		ByReleasing:    SortLink{Href: rd, Active: a1},
		ByViewer:       SortLink{Href: vw, Active: a2},
		ByDaysInRank:   SortLink{Href: drk, Active: a3},
		ByHighestRank:  SortLink{Href: hrk, Active: a4},
		ByRankDate:     SortLink{Href: rk, Active: a5},
		ByScTimes:      SortLink{Href: sc, Active: a6},
		ByLastScTime:   SortLink{Href: lsct, Active: a7},
		ByBirthTime:    SortLink{Href: bt, Active: a8},
		ByCastAgeDesc:  SortLink{Href: cad, Active: a9},
		ByCastAgeAsc:   SortLink{Href: caa, Active: a10},
		ByComeTimes:    SortLink{Href: co, Active: a11},
	}
}

// 复制 URL 查询参数
func cloneValues(c *gin.Context) url.Values {
	orig := c.Request.URL.Query()
	q := make(url.Values, len(orig))
	for k, vv := range orig {
		cp := make([]string, len(vv))
		copy(cp, vv)
		q[k] = cp
	}
	return q
}
