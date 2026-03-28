package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type filmSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type filmSortQuery struct {
	ByMovieName     filmSortLink
	BySize          filmSortLink
	ByHeight        filmSortLink
	ByDuration      filmSortLink
	ByBitRate       filmSortLink
	ByFrameAverage  filmSortLink
	BySelfMake      filmSortLink
	ByHasMask       filmSortLink
	ByScTimes       filmSortLink
	ByLastSc        filmSortLink
	ByBirthTime     filmSortLink
	ByReleasingDate filmSortLink
}

func buildFilmSortQuery(c *gin.Context, currentField, currentOrder string) *filmSortQuery {
	makeHref := func(field string) filmSortLink {
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

		return filmSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &filmSortQuery{
		ByMovieName:     makeHref("movie_name"),
		BySize:          makeHref("size"),
		ByHeight:        makeHref("height"),
		ByDuration:      makeHref("duration"),
		ByBitRate:       makeHref("bit_rate"),
		ByFrameAverage:  makeHref("frame_average"),
		BySelfMake:      makeHref("self_make"),
		ByHasMask:       makeHref("has_mask"),
		ByScTimes:       makeHref("sc_times"),
		ByLastSc:        makeHref("last_sc_time"),
		ByBirthTime:     makeHref("birth_time"),
		ByReleasingDate: makeHref("releasing_date"),
	}
}

func normalizeFilmSortFieldForPage(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "movie_name", "size", "height", "duration", "bit_rate", "frame_average", "self_make", "has_mask", "sc_times", "last_sc_time", "birth_time", "releasing_date":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "birth_time"
	}
}

func normalizeFilmSortOrderForPage(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "asc") {
		return "asc"
	}
	return "desc"
}
