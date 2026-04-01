package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type PageLink struct {
	Label    string // 显示文本，如 "1" / "Prev" / "..."
	Page     int64  // 对应页码（省略号可为0）
	Href     string // 链接
	Active   bool   // 当前页
	Disabled bool   // 是否禁用
	Ellipsis bool   // 是否省略号
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

	Links []PageLink // 包含 Start / Prev / 数字区 / Next / End
}

type OwnedLink struct {
	Href   string
	Active bool
}
type OwnedQuery struct {
	All           OwnedLink // owned=1
	Owned         OwnedLink // owned=3
	NotOwned      OwnedLink // owned=7
	MediaAll      OwnedLink // mowned=1
	MediaOwned    OwnedLink // mowned=3
	MediaNotOwned OwnedLink // mowned=7
}

func buildOwnedFilterInfo(c *gin.Context) *OwnedQuery {
	curOwned := c.Query("owned")
	curMediaOwned := c.Query("mowned")

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

	allHref := makeHref(func(q mapSetter) {
		q.Set("owned", "1")
	})
	mediaAllHref := makeHref(func(q mapSetter) {
		q.Set("mowned", "1")
	})
	ownedHref := makeHref(func(q mapSetter) {
		q.Set("owned", "3")
	})
	mediaOwnedHref := makeHref(func(q mapSetter) {
		q.Set("mowned", "3")
	})
	notHref := makeHref(func(q mapSetter) {
		q.Set("owned", "7")
	})
	mediaNotOwnedHref := makeHref(func(q mapSetter) {
		q.Set("mowned", "7")
	})

	allAct := curOwned == "" || curOwned == "1"
	ownedAct := curOwned == "3"
	mediaAllAct := curMediaOwned == "" || curMediaOwned == "1"
	mediaOwnedAct := curMediaOwned == "3"
	notAct := curOwned == "7"
	mediaNotAct := curMediaOwned == "7"

	return &OwnedQuery{
		All:           OwnedLink{Href: allHref, Active: allAct},
		Owned:         OwnedLink{Href: ownedHref, Active: ownedAct},
		MediaAll:      OwnedLink{Href: mediaAllHref, Active: mediaAllAct},
		MediaOwned:    OwnedLink{Href: mediaOwnedHref, Active: mediaOwnedAct},
		NotOwned:      OwnedLink{Href: notHref, Active: notAct},
		MediaNotOwned: OwnedLink{Href: mediaNotOwnedHref, Active: mediaNotAct},
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
		q := c.Request.URL.Query() // clone values
		q.Set("p", fmt.Sprint(p))
		u := c.Request.URL.Path
		qs := q.Encode()
		if qs != "" {
			u += "?" + qs
		}
		return u
	}

	links := make([]PageLink, 0, 12)

	// Start / Prev
	start := int64(1)
	prev := page - 1
	if prev < 1 {
		prev = 1
	}
	links = append(links,
		PageLink{Label: "Start", Page: start, Href: makeHref(start), Disabled: page == 1},
		PageLink{Label: "Prev", Page: prev, Href: makeHref(prev), Disabled: page == 1},
	)

	// 数字页窗口（带省略号）
	if window <= 0 {
		window = 5
	}
	left := page - int64(window)
	right := page + int64(window)

	if left > 1 {
		// 1 ...
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

	// Next / End
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
