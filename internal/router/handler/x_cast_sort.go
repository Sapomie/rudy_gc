package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type castSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type castSortQuery struct {
	ByName        castSortLink
	ByChinese     castSortLink
	ByAge         castSortLink
	ByHeight      castSortLink
	ByMovieNumber castSortLink
	ByOwned       castSortLink
	ByScTimes     castSortLink
	ByComeTimes   castSortLink
	ByLastSc      castSortLink
	ByHighestRank castSortLink
}

func buildCastSortQuery(c *gin.Context, currentField, currentOrder string) *castSortQuery {
	makeHref := func(field string) castSortLink {
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

		return castSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &castSortQuery{
		ByName:        makeHref("name"),
		ByChinese:     makeHref("chinese"),
		ByAge:         makeHref("age"),
		ByHeight:      makeHref("height"),
		ByMovieNumber: makeHref("movie_number"),
		ByOwned:       makeHref("owned_movie_number"),
		ByScTimes:     makeHref("sc_times"),
		ByComeTimes:   makeHref("come_times"),
		ByLastSc:      makeHref("last_sc_time"),
		ByHighestRank: makeHref("highest_rank"),
	}
}

func normalizeCastSortFieldForPage(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "name", "chinese", "age", "height", "movie_number", "owned_movie_number", "sc_times", "come_times", "last_sc_time", "highest_rank":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "owned_movie_number"
	}
}

func normalizeCastSortOrderForPage(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "asc") {
		return "asc"
	}
	return "desc"
}
