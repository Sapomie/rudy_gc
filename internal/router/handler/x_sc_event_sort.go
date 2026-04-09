package handler

import "github.com/gin-gonic/gin"

type scEventSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type scEventSortQuery struct {
	ByTime      scEventSortLink
	ByMovieCnt  scEventSortLink
	ByComeMovie scEventSortLink
	ByCooldown  scEventSortLink
	ByDuration  scEventSortLink
	ByMovieCast scEventSortLink
	ByVessel    scEventSortLink
	ByFg        scEventSortLink
}

func buildScEventSortQuery(c *gin.Context, currentField, currentOrder string) *scEventSortQuery {
	makeHref := func(field string) scEventSortLink {
		q := cloneValues(c)
		q.Set("p", "1")
		q.Set("sort", field)

		if currentField == field && currentOrder == "desc" {
			q.Set("order", "asc")
		} else {
			q.Set("order", "desc")
		}

		href := c.Request.URL.Path
		if enc := q.Encode(); enc != "" {
			href += "?" + enc
		}

		return scEventSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &scEventSortQuery{
		ByTime:      makeHref("sc_time"),
		ByMovieCnt:  makeHref("movie_number"),
		ByComeMovie: makeHref("come_movie_name"),
		ByCooldown:  makeHref("cooldown"),
		ByDuration:  makeHref("duration"),
		ByMovieCast: makeHref("movie_cast"),
		ByVessel:    makeHref("vessel"),
		ByFg:        makeHref("fg"),
	}
}

func normalizeScEventSortField(v string) string {
	switch v {
	case "movie_number", "come_movie_name", "cooldown", "duration", "movie_cast", "vessel", "fg", "sc_time":
		return v
	default:
		return "sc_time"
	}
}

func normalizeScEventSortOrder(v string) string {
	if v == "asc" {
		return "asc"
	}
	return "desc"
}
