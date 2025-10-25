package html

import (
	"rudy_gc/internal/consts"

	"github.com/gin-gonic/gin"
)

type sortLink struct {
	Href   string
	Active bool
}
type dirSortQuery struct {
	ByBirthTime  sortLink
	ByLastScTime sortLink
	ByScTimes    sortLink
	ByComeTimes  sortLink
	ByReleasing  sortLink
}

// 目录页排序栏：只构造 bt/lsct/sc/co/rd 5 个
func buildDirSortQuery(c *gin.Context, current string) *dirSortQuery {
	// 复制当前 URL query，修改 od，并且切换排序时回到第一页 p=1
	makeHref := func(od string) (href string, active bool) {
		q := c.Request.URL.Query()
		q.Set("p", "1")
		if od == "" {
			q.Del("od")
		} else {
			q.Set("od", od)
		}
		path := c.Request.URL.Path
		enc := q.Encode()
		if enc != "" {
			href = path + "?" + enc
		} else {
			href = path
		}
		active = (current == od || (current == "" && od == "")) // 简单相等判断
		return
	}

	btHref, btAct := makeHref(consts.OrderByBirthTime)
	lsctHref, lsctAct := makeHref(consts.OrderByLastScTime)
	scHref, scAct := makeHref(consts.OrderByScTimes)
	coHref, coAct := makeHref(consts.OrderByComeTimes)
	rdHref, rdAct := makeHref(consts.OrderByReleasingDate)

	return &dirSortQuery{
		ByBirthTime:  sortLink{Href: btHref, Active: btAct},
		ByLastScTime: sortLink{Href: lsctHref, Active: lsctAct},
		ByScTimes:    sortLink{Href: scHref, Active: scAct},
		ByComeTimes:  sortLink{Href: coHref, Active: coAct},
		ByReleasing:  sortLink{Href: rdHref, Active: rdAct},
	}
}
