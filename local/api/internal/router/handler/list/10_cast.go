package list

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/router/handler/common"
	"rudy_gc/local/api/internal/types"
)

type castListQuery struct {
	Page               int64  `form:"p"`
	PageSize           int64  `form:"ps"`
	Sort               string `form:"sort"`
	Order              string `form:"order"`
	Keyword            string `form:"keyword"`
	OwnedMin           string `form:"wowned_min"`
	OwnedMax           string `form:"wowned_max"`
	LegacyMin          string `form:"owned_min"`
	LegacyMax          string `form:"owned_max"`
	ScTimesMin         string `form:"sc_times_min"`
	ScTimesMax         string `form:"sc_times_max"`
	ComeTimesMin       string `form:"come_times_min"`
	ComeTimesMax       string `form:"come_times_max"`
	LastSelectedScFrom string `form:"last_selected_sc_from"`
	LastSelectedScTo   string `form:"last_selected_sc_to"`
}

func (h *ListPagesAPI) CastList(c *gin.Context) {
	defer common.WithFrontendPath(c, "/castlist")()

	var q castListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.JSONBadRequest(c, "参数解析错误", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 24
	}
	if strings.TrimSpace(q.OwnedMin) == "" {
		q.OwnedMin = strings.TrimSpace(q.LegacyMin)
	}
	if strings.TrimSpace(q.OwnedMax) == "" {
		q.OwnedMax = strings.TrimSpace(q.LegacyMax)
	}
	q.Sort = normalizeCastSortFieldForPage(q.Sort)
	q.Order = normalizeCastSortOrderForPage(q.Order)

	filter, err := buildCastListFilter(q)
	if err != nil {
		common.JSONBadRequest(c, "筛选参数错误", err)
		return
	}
	total, err := h.deps.PersonModel.CountAll(c.Request.Context(), filter)
	if err != nil {
		common.JSONServerError(c, "演员列表加载失败", err)
		return
	}

	offset := (q.Page - 1) * q.PageSize
	items, err := h.deps.PersonModel.ListPage(
		c.Request.Context(),
		offset,
		q.PageSize,
		buildCastOrderByForPage(q.Sort, q.Order),
		filter,
	)
	if err != nil {
		common.JSONServerError(c, "演员列表加载失败", err)
		return
	}

	pi := common.BuildPageInfo(c, total, q.Page, q.PageSize, common.PageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":         "Casts",
		"ViewType":      "casts",
		"Items":         items,
		"Total":         total,
		"PageInfo":      pi,
		"pageInfo":      pi,
		"CastSortQuery": buildCastSortQuery(c, q.Sort, q.Order),
		"Query":         q,
	})
}

func buildCastListFilter(q castListQuery) (types.PersonListFilter, error) {
	filter := types.PersonListFilter{
		Keyword: strings.TrimSpace(q.Keyword),
	}

	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.OwnedMin); err != nil {
		return types.PersonListFilter{}, err
	} else if ok {
		filter.OwnedMin = v
		filter.HasOwnedMin = true
	}
	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.OwnedMax); err != nil {
		return types.PersonListFilter{}, err
	} else if ok {
		filter.OwnedMax = v
		filter.HasOwnedMax = true
	}
	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.ScTimesMin); err != nil {
		return types.PersonListFilter{}, err
	} else if ok {
		filter.ScTimesMin = v
		filter.HasScTimesMin = true
	}
	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.ScTimesMax); err != nil {
		return types.PersonListFilter{}, err
	} else if ok {
		filter.ScTimesMax = v
		filter.HasScTimesMax = true
	}
	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.ComeTimesMin); err != nil {
		return types.PersonListFilter{}, err
	} else if ok {
		filter.ComeTimesMin = v
		filter.HasComeTimesMin = true
	}
	if v, ok, err := common.ParseOptionalNonNegativeInt64(q.ComeTimesMax); err != nil {
		return types.PersonListFilter{}, err
	} else if ok {
		filter.ComeTimesMax = v
		filter.HasComeTimesMax = true
	}

	if strings.TrimSpace(q.LastSelectedScFrom) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(q.LastSelectedScFrom), time.Local)
		if err != nil {
			return types.PersonListFilter{}, err
		}
		filter.LastSelectedScFrom = t.Unix()
		filter.HasLastSelectedScFrom = true
	}

	if strings.TrimSpace(q.LastSelectedScTo) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(q.LastSelectedScTo), time.Local)
		if err != nil {
			return types.PersonListFilter{}, err
		}
		filter.LastSelectedScTo = t.Add(24*time.Hour - time.Second).Unix()
		filter.HasLastSelectedScTo = true
	}

	return filter, nil
}

func buildCastOrderByForPage(sortField, sortOrder string) string {
	field := normalizeCastSortFieldForPage(sortField)
	order := "DESC"
	if normalizeCastSortOrderForPage(sortOrder) == "asc" {
		order = "ASC"
	}

	column := "p.owned_w_media_number"
	switch field {
	case "name":
		column = "p.name"
	case "chinese":
		column = "p.chinese"
	case "age":
		if order == "ASC" {
			return "CASE WHEN p.birth_day > 0 THEN 0 ELSE 1 END ASC, p.birth_day DESC, p.owned_w_media_number DESC, p.movie_number DESC, p.name ASC, p.id DESC"
		}
		return "CASE WHEN p.birth_day > 0 THEN 0 ELSE 1 END ASC, p.birth_day ASC, p.owned_w_media_number DESC, p.movie_number DESC, p.name ASC, p.id DESC"
	case "height":
		return "CASE WHEN p.height > 0 THEN 0 ELSE 1 END ASC, p.height " + order + ", p.owned_w_media_number DESC, p.movie_number DESC, p.name ASC, p.id DESC"
	case "movie_number":
		column = "p.movie_number"
	case "owned_w_media_number":
		column = "p.owned_w_media_number"
	case "sc_times":
		column = "p.sc_times"
	case "come_times":
		column = "p.come_times"
	case "last_selected_sc_time":
		column = "p.last_selected_sc_time"
	case "highest_rank":
		column = "p.highest_rank"
	}

	if column == "p.owned_w_media_number" {
		return column + " " + order + ", p.movie_number DESC, p.name ASC, p.id DESC"
	}
	if column == "p.name" {
		return column + " " + order + ", p.owned_w_media_number DESC, p.movie_number DESC, p.id DESC"
	}
	if column == "p.chinese" {
		return "CASE WHEN p.chinese <> '' THEN 0 ELSE 1 END ASC, " + column + " " + order + ", p.owned_w_media_number DESC, p.movie_number DESC, p.name ASC, p.id DESC"
	}
	return column + " " + order + ", p.owned_w_media_number DESC, p.movie_number DESC, p.name ASC, p.id DESC"
}

type castSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type castSortQuery struct {
	ByName           castSortLink
	ByChinese        castSortLink
	ByAge            castSortLink
	ByHeight         castSortLink
	ByMovieNumber    castSortLink
	ByWOwned         castSortLink
	ByScTimes        castSortLink
	ByComeTimes      castSortLink
	ByLastSelectedSc castSortLink
	ByHighestRank    castSortLink
}

func buildCastSortQuery(c *gin.Context, currentField, currentOrder string) *castSortQuery {
	makeHref := func(field string) castSortLink {
		q := common.CloneValues(c)
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
		ByName:           makeHref("name"),
		ByChinese:        makeHref("chinese"),
		ByAge:            makeHref("age"),
		ByHeight:         makeHref("height"),
		ByMovieNumber:    makeHref("movie_number"),
		ByWOwned:         makeHref("owned_w_media_number"),
		ByScTimes:        makeHref("sc_times"),
		ByComeTimes:      makeHref("come_times"),
		ByLastSelectedSc: makeHref("last_selected_sc_time"),
		ByHighestRank:    makeHref("highest_rank"),
	}
}

func normalizeCastSortFieldForPage(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "owned_movie_number":
		return "owned_w_media_number"
	case "name", "chinese", "age", "height", "movie_number", "owned_w_media_number", "sc_times", "come_times", "last_selected_sc_time", "highest_rank":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "owned_w_media_number"
	}
}

func normalizeCastSortOrderForPage(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "asc") {
		return "asc"
	}
	return "desc"
}
