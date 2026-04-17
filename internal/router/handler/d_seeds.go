package handler

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/spiderx"
)

type dSeedListQuery struct {
	Page       int64  `form:"p"`
	PageSize   int64  `form:"ps"`
	Sort       string `form:"sort"`
	Order      string `form:"order"`
	Name       string `form:"name"`
	Active     string `form:"active"`
	SearchType string `form:"search_type"`
	NameType   string `form:"name_type"`
	LastStatus string `form:"last_status"`
}

type dSeedSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type dSeedSortQuery struct {
	ByName                          dSeedSortLink
	ByActive                        dSeedSortLink
	BySearchType                    dSeedSortLink
	ByNameType                      dSeedSortLink
	ByMovieTotal                    dSeedSortLink
	ByMovieLatestReleasingMovieName dSeedSortLink
	ByMovieLatestReleasingDate      dSeedSortLink
	ByMovieLastAddedTime            dSeedSortLink
	ByLastInsertCount               dSeedSortLink
	ByPageNow                       dSeedSortLink
	ByLastQueryTime                 dSeedSortLink
}

type dSeedListRow struct {
	Id                             int64
	Name                           string
	CardsHref                      string
	ActiveValue                    int64
	ActiveText                     string
	ActiveClass                    string
	SearchTypeValue                int64
	SearchTypeText                 string
	NameTypeValue                  int64
	NameTypeText                   string
	PageNow                        int64
	Offset                         int64
	StartPage                      int64
	EndPage                        int64
	LastQueryTime                  int64
	LastStatusValue                int64
	LastStatusText                 string
	LastStatusClass                string
	LastError                      string
	MovieTotal                     int64
	MovieLatestReleasingMovieJavId string
	MovieLatestReleasingMovieName  string
	MovieLatestReleasingMovieHref  string
	MovieLastAddedTime             int64
	LastInsertCount                int64
	MovieLatestReleasingDate       int64
	CreatedOn                      int64
	UpdatedOn                      int64
}

func (h *CrawlerPages) DSeedListPage(c *gin.Context) {
	var q dSeedListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 200 {
		q.PageSize = 50
	}
	q.Sort = normalizeDSeedSortField(q.Sort)
	q.Order = normalizeDSeedSortOrder(q.Order)

	filter, err := buildDSeedListFilter(q)
	if err != nil {
		c.String(http.StatusBadRequest, "筛选参数错误: %v", err)
		return
	}

	offset := (q.Page - 1) * q.PageSize
	rows, err := h.deps.SeedModel.ListPage(c.Request.Context(), offset, q.PageSize, buildDSeedOrderByForPage(q.Sort, q.Order), filter)
	if err != nil {
		c.String(http.StatusInternalServerError, "d_seed 列表加载失败: %v", err)
		return
	}
	total, err := h.deps.SeedModel.CountPage(c.Request.Context(), filter)
	if err != nil {
		c.String(http.StatusInternalServerError, "d_seed 总数加载失败: %v", err)
		return
	}

	items := make([]*dSeedListRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, buildDSeedListRowFromModel(row))
	}

	pi := BuildPageInfo(c, total, q.Page, q.PageSize, pageWindow)
	c.HTML(http.StatusOK, "page.d_seed_list", gin.H{
		"Title":           "DSeed",
		"PageTitle":       "DSeed 列表",
		"PageNote":        "查看 d_seed 查询种子的当前状态、分页断点和电影覆盖情况。",
		"Items":           items,
		"Total":           total,
		"PageInfo":        pi,
		"pageInfo":        pi,
		"Query":           q,
		"DSeedSort":       buildDSeedSortQuery(c, q.Sort, q.Order),
		"QuickNavCurrent": "d_seed_list",
	})
}

func buildDSeedListRowFromModel(row *spiderx.DSeed) *dSeedListRow {
	if row == nil {
		return nil
	}
	return &dSeedListRow{
		Id:                             row.Id,
		Name:                           row.Name,
		CardsHref:                      buildDSeedCardsHref(row.Name, row.NameType),
		ActiveValue:                    row.Active,
		ActiveText:                     dSeedActiveText(row.Active),
		ActiveClass:                    dSeedActiveClass(row.Active),
		SearchTypeValue:                row.SearchType,
		SearchTypeText:                 dSeedSearchTypeText(row.SearchType),
		NameTypeValue:                  row.NameType,
		NameTypeText:                   dSeedNameTypeText(row.NameType),
		PageNow:                        row.PageNow,
		Offset:                         row.Offset,
		StartPage:                      row.StartPage,
		EndPage:                        row.EndPage,
		LastQueryTime:                  row.LastQueryTime,
		LastStatusValue:                row.LastStatus,
		LastStatusText:                 dSeedStatusText(row.LastStatus),
		LastStatusClass:                dSeedStatusClass(row.LastStatus),
		LastError:                      strings.TrimSpace(row.LastError),
		MovieTotal:                     row.MovieTotal,
		MovieLatestReleasingMovieJavId: row.MovieLatestReleasingMovieJavId,
		MovieLatestReleasingMovieName:  row.MovieLatestReleasingMovieName,
		MovieLatestReleasingMovieHref:  buildMovieDetailHrefByName(row.MovieLatestReleasingMovieName),
		MovieLastAddedTime:             row.MovieLastAddedTime,
		LastInsertCount:                row.LastInsertCount,
		MovieLatestReleasingDate:       row.MovieLatestReleasingDate,
		CreatedOn:                      row.CreatedOn,
		UpdatedOn:                      row.UpdatedOn,
	}
}

func buildDSeedCardsHref(name string, nameType int64) string {
	queryName := strings.TrimSpace(name)
	if queryName == "" {
		return ""
	}

	q := url.Values{}
	switch nameType {
	case spiderx.QueryNamePrefix:
		q.Set("pn", queryName)
	case spiderx.QueryNameLabel:
		q.Set("lj", queryName)
	default:
		return ""
	}
	return "/cards?" + q.Encode()
}

func buildMovieDetailHrefByName(movieName string) string {
	name := strings.TrimSpace(movieName)
	if name == "" {
		return ""
	}
	return "/movie/" + url.PathEscape(name)
}

func buildDSeedListFilter(q dSeedListQuery) (spiderx.DSeedListFilter, error) {
	filter := spiderx.DSeedListFilter{
		NameKeyword: strings.TrimSpace(q.Name),
	}

	if v, ok, err := parseDSeedOptionalInt(q.Active); err != nil {
		return spiderx.DSeedListFilter{}, err
	} else if ok {
		filter.Active = &v
	}
	if v, ok, err := parseDSeedOptionalInt(q.SearchType); err != nil {
		return spiderx.DSeedListFilter{}, err
	} else if ok {
		filter.SearchType = &v
	}
	if v, ok, err := parseDSeedOptionalInt(q.NameType); err != nil {
		return spiderx.DSeedListFilter{}, err
	} else if ok {
		filter.NameType = &v
	}
	if v, ok, err := parseDSeedOptionalInt(q.LastStatus); err != nil {
		return spiderx.DSeedListFilter{}, err
	} else if ok {
		filter.LastStatus = &v
	}

	return filter, nil
}

func parseDSeedOptionalInt(raw string) (int64, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false, err
	}
	return v, true, nil
}

func normalizeDSeedSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "name", "active", "search_type", "name_type", "movie_total", "movie_latest_releasing_movie_name", "movie_latest_releasing_date", "movie_last_added_time", "last_insert_count", "page_now", "last_query_time", "last_status", "updated_on", "created_on":
		return strings.TrimSpace(raw)
	default:
		return "updated_on"
	}
}

func normalizeDSeedSortOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}

func buildDSeedOrderByForPage(sortField string, sortOrder string) string {
	fieldMap := map[string]string{
		"name":                              "`name`",
		"active":                            "`active`",
		"search_type":                       "`search_type`",
		"name_type":                         "`name_type`",
		"movie_total":                       "`movie_total`",
		"movie_latest_releasing_movie_name": "`movie_latest_releasing_movie_name`",
		"movie_latest_releasing_date":       "`movie_latest_releasing_date`",
		"movie_last_added_time":             "`movie_last_added_time`",
		"last_insert_count":                 "`last_insert_count`",
		"page_now":                          "`page_now`",
		"last_query_time":                   "`last_query_time`",
		"last_status":                       "`last_status`",
		"updated_on":                        "`updated_on`",
		"created_on":                        "`created_on`",
	}
	column, ok := fieldMap[sortField]
	if !ok {
		column = "`updated_on`"
	}
	order := "DESC"
	if sortOrder == "asc" {
		order = "ASC"
	}
	return column + " " + order + ", `id` DESC"
}

func buildDSeedSortQuery(c *gin.Context, currentField string, currentOrder string) *dSeedSortQuery {
	makeHref := func(field string) dSeedSortLink {
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
		return dSeedSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &dSeedSortQuery{
		ByName:                          makeHref("name"),
		ByActive:                        makeHref("active"),
		BySearchType:                    makeHref("search_type"),
		ByNameType:                      makeHref("name_type"),
		ByMovieTotal:                    makeHref("movie_total"),
		ByMovieLatestReleasingMovieName: makeHref("movie_latest_releasing_movie_name"),
		ByMovieLatestReleasingDate:      makeHref("movie_latest_releasing_date"),
		ByMovieLastAddedTime:            makeHref("movie_last_added_time"),
		ByLastInsertCount:               makeHref("last_insert_count"),
		ByPageNow:                       makeHref("page_now"),
		ByLastQueryTime:                 makeHref("last_query_time"),
	}
}

func dSeedActiveText(v int64) string {
	switch v {
	case spiderx.QueryActive:
		return "启用"
	case spiderx.QueryInactive:
		return "停用"
	default:
		return "未知"
	}
}

func dSeedActiveClass(v int64) string {
	switch v {
	case spiderx.QueryActive:
		return "d-seed-status-badge d-seed-status-active"
	case spiderx.QueryInactive:
		return "d-seed-status-badge d-seed-status-inactive"
	default:
		return "d-seed-status-badge d-seed-status-unknown"
	}
}

func dSeedSearchTypeText(v int64) string {
	switch v {
	case spiderx.QueryByOffset:
		return "偏移量"
	case spiderx.QueryByStartEnd:
		return "起止页"
	default:
		return "未知"
	}
}

func dSeedNameTypeText(v int64) string {
	switch v {
	case spiderx.QueryNamePrefix:
		return "Prefix"
	case spiderx.QueryNameLabel:
		return "Label"
	default:
		return "未知"
	}
}

func dSeedStatusText(v int64) string {
	switch v {
	case consts.SeedStatusIdle:
		return "空闲"
	case consts.SeedStatusOK:
		return "成功"
	case consts.SeedStatusEmpty:
		return "空页"
	case consts.SeedStatusError:
		return "错误"
	default:
		return "未知"
	}
}

func dSeedStatusClass(v int64) string {
	switch v {
	case consts.SeedStatusIdle:
		return "text-bg-secondary"
	case consts.SeedStatusOK:
		return "text-bg-success"
	case consts.SeedStatusEmpty:
		return "text-bg-warning"
	case consts.SeedStatusError:
		return "text-bg-danger"
	default:
		return "text-bg-dark"
	}
}
