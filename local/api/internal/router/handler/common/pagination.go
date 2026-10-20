package common

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/consts"
)

const (
	DefaultPageSize = 18
	MaxPageSize     = 20000
	PageWindow      = 3
)

type PageLink struct {
	Label    string
	Page     int64
	Href     string
	Active   bool
	Disabled bool
	Ellipsis bool
}

type PageInfo struct {
	Total     int64
	Page      int64
	PageSize  int64
	PageTotal int64
	StartHref string
	PrevHref  string
	NextHref  string
	EndHref   string
	Links     []PageLink
}

type OwnedLink struct {
	Href   string
	Active bool
}

type OwnedQuery struct {
	MediaAll      OwnedLink
	MediaOwned    OwnedLink
	MediaNotOwned OwnedLink
}

func BuildOwnedFilterInfo(c *gin.Context) *OwnedQuery {
	return BuildOwnedFilterInfoWithDefaults(c, "")
}

func BuildOwnedFilterInfoWithDefaults(c *gin.Context, defaultMediaOwned string) *OwnedQuery {
	curMediaOwned := c.Query("mowned")
	if curMediaOwned == "" {
		curMediaOwned = defaultMediaOwned
	}

	makeHref := func(mutator func(q mapSetter)) string {
		q := c.Request.URL.Query()
		q.Set("p", "1")
		mutator(q)
		path := c.Request.URL.Path
		if enc := q.Encode(); enc != "" {
			return path + "?" + enc
		}
		return path
	}

	mediaAllHref := makeHref(func(q mapSetter) {
		q.Set("mowned", "1")
	})
	mediaOwnedHref := makeHref(func(q mapSetter) {
		q.Set("mowned", "3")
	})
	mediaNotOwnedHref := makeHref(func(q mapSetter) {
		q.Set("mowned", "7")
	})

	return &OwnedQuery{
		MediaAll:      OwnedLink{Href: mediaAllHref, Active: curMediaOwned == "" || curMediaOwned == "1"},
		MediaOwned:    OwnedLink{Href: mediaOwnedHref, Active: curMediaOwned == "3"},
		MediaNotOwned: OwnedLink{Href: mediaNotOwnedHref, Active: curMediaOwned == "7"},
	}
}

type mapSetter interface {
	Del(key string)
	Set(key, value string)
	Encode() string
}

func BuildPageInfo(c *gin.Context, total, page, pageSize int64, window int) *PageInfo {
	if pageSize <= 0 {
		pageSize = 20
	}
	pageTotal := (total + pageSize - 1) / pageSize
	if pageTotal <= 0 {
		pageTotal = 1
	}
	if page < 1 {
		page = 1
	}
	if page > pageTotal {
		page = pageTotal
	}
	makeHref := func(p int64) string {
		q := c.Request.URL.Query()
		q.Set("p", fmt.Sprint(p))
		u := c.Request.URL.Path
		qs := q.Encode()
		if qs != "" {
			u += "?" + qs
		}
		return u
	}

	links := make([]PageLink, 0, 12)
	start := int64(1)
	prev := page - 1
	if prev < 1 {
		prev = 1
	}
	links = append(links,
		PageLink{Label: "Start", Page: start, Href: makeHref(start), Disabled: page == 1},
		PageLink{Label: "Prev", Page: prev, Href: makeHref(prev), Disabled: page == 1},
	)

	if window <= 0 {
		window = 5
	}
	left := page - int64(window)
	right := page + int64(window)
	if left > 1 {
		links = append(links, PageLink{Label: "1", Page: 1, Href: makeHref(1)})
		if left > 2 {
			links = append(links, PageLink{Label: "...", Ellipsis: true, Disabled: true})
		}
		left = max(left, 2)
	}
	if right < pageTotal {
		right = min(right, pageTotal-1)
	}
	for p := left; p <= right; p++ {
		if p < 1 || p > pageTotal {
			continue
		}
		links = append(links, PageLink{
			Label:  fmt.Sprint(p),
			Page:   p,
			Href:   makeHref(p),
			Active: p == page,
		})
	}
	if right < pageTotal {
		if right < pageTotal-1 {
			links = append(links, PageLink{Label: "...", Ellipsis: true, Disabled: true})
		}
		links = append(links, PageLink{Label: fmt.Sprint(pageTotal), Page: pageTotal, Href: makeHref(pageTotal)})
	}

	next := page + 1
	if next > pageTotal {
		next = pageTotal
	}
	links = append(links,
		PageLink{Label: "Next", Page: next, Href: makeHref(next), Disabled: page == pageTotal},
		PageLink{Label: "End", Page: pageTotal, Href: makeHref(pageTotal), Disabled: page == pageTotal},
	)

	return &PageInfo{
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		PageTotal: pageTotal,
		StartHref: makeHref(start),
		PrevHref:  makeHref(prev),
		NextHref:  makeHref(next),
		EndHref:   makeHref(pageTotal),
		Links:     links,
	}
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

var allowedOrderBy = map[string]struct{}{
	consts.OrderByDetailUpdateTime:   {},
	consts.OrderByCastAgeDesc:        {},
	consts.OrderByCastAgeAsc:         {},
	consts.OrderByViewerWatched:      {},
	consts.OrderByReleasingDate:      {},
	consts.OrderByRankDate:           {},
	consts.OrderByHighestRank:        {},
	consts.OrderByDaysInRank:         {},
	consts.OrderByBirthTime:          {},
	consts.OrderByMediaBirthTime:     {},
	consts.OrderByScTimes:            {},
	consts.OrderByComeTimes:          {},
	consts.OrderByLastSelectedScTime: {},
}

func NormalizeOrderBy(orderBy, fallback string) string {
	if _, ok := allowedOrderBy[orderBy]; ok {
		return orderBy
	}
	return fallback
}

func NormalizeOrder(order string) string {
	switch strings.ToLower(strings.TrimSpace(order)) {
	case "asc", "desc":
		return strings.ToLower(strings.TrimSpace(order))
	default:
		return ""
	}
}

type SortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type SortQuery struct {
	ByDetailUpdate       SortLink
	ByReleasing          SortLink
	ByViewer             SortLink
	ByDaysInRank         SortLink
	ByHighestRank        SortLink
	ByRankDate           SortLink
	ByScTimes            SortLink
	ByLastSelectedScTime SortLink
	ByBirthTime          SortLink
	ByMediaBirthTime     SortLink
	ByCastAgeDesc        SortLink
	ByCastAgeAsc         SortLink
	ByComeTimes          SortLink
}

func BuildSortQuery(c *gin.Context, currentOrderBy string) *SortQuery {
	currentOrder := movieCardEffectiveOrder(currentOrderBy, c.Query("order"))
	makeLink := func(orderBy string) SortLink {
		active := currentOrderBy == orderBy
		nextOrder := movieCardDefaultOrder(orderBy)
		if active {
			nextOrder = oppositeMovieCardOrder(currentOrder)
		}
		q := CloneValues(c)
		q.Set("od", orderBy)
		q.Set("order", nextOrder)
		q.Set("p", "1")
		href := c.Request.URL.Path
		if enc := q.Encode(); enc != "" {
			href += "?" + enc
		}
		return SortLink{
			Href:   href,
			Active: active,
			Desc:   active && currentOrder == "desc",
		}
	}

	return &SortQuery{
		ByDetailUpdate:       makeLink(consts.OrderByDetailUpdateTime),
		ByReleasing:          makeLink(consts.OrderByReleasingDate),
		ByViewer:             makeLink(consts.OrderByViewerWatched),
		ByDaysInRank:         makeLink(consts.OrderByDaysInRank),
		ByHighestRank:        makeLink(consts.OrderByHighestRank),
		ByRankDate:           makeLink(consts.OrderByRankDate),
		ByScTimes:            makeLink(consts.OrderByScTimes),
		ByLastSelectedScTime: makeLink(consts.OrderByLastSelectedScTime),
		ByBirthTime:          makeLink(consts.OrderByBirthTime),
		ByMediaBirthTime:     makeLink(consts.OrderByMediaBirthTime),
		ByCastAgeDesc:        makeLink(consts.OrderByCastAgeDesc),
		ByCastAgeAsc:         makeLink(consts.OrderByCastAgeAsc),
		ByComeTimes:          makeLink(consts.OrderByComeTimes),
	}
}

func movieCardDefaultOrder(orderBy string) string {
	switch orderBy {
	case consts.OrderByHighestRank, consts.OrderByCastAgeAsc:
		return "asc"
	default:
		return "desc"
	}
}

func movieCardEffectiveOrder(orderBy, order string) string {
	if normalized := NormalizeOrder(order); normalized != "" {
		return normalized
	}
	return movieCardDefaultOrder(orderBy)
}

func oppositeMovieCardOrder(order string) string {
	if order == "asc" {
		return "desc"
	}
	return "asc"
}

func CloneValues(c *gin.Context) url.Values {
	original := c.Request.URL.Query()
	values := make(url.Values, len(original))
	for key, source := range original {
		copied := make([]string, len(source))
		copy(copied, source)
		values[key] = copied
	}
	return values
}
