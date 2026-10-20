package crawler

import (
	"errors"
	"net/http"
	"net/url"
	"rudy_gc/local/api/internal/router/handler/common"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/spiderx"
)

type DSeedAPI struct {
	deps *dep.Dep
}

func NewDSeedAPI(deps *dep.Dep) *DSeedAPI {
	return &DSeedAPI{deps: deps}
}

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

func (h *DSeedAPI) List(c *gin.Context) {
	defer common.WithFrontendPath(c, "/seedlist")()

	var q dSeedListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.JSONBadRequest(c, "参数解析错误", err)
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
		common.JSONBadRequest(c, "筛选参数错误", err)
		return
	}

	offset := (q.Page - 1) * q.PageSize
	rows, err := h.deps.SeedModel.ListPage(c.Request.Context(), offset, q.PageSize, buildDSeedOrderByForPage(q.Sort, q.Order), filter)
	if err != nil {
		common.JSONServerError(c, "d_seed 列表加载失败", err)
		return
	}
	total, err := h.deps.SeedModel.CountPage(c.Request.Context(), filter)
	if err != nil {
		common.JSONServerError(c, "d_seed 总数加载失败", err)
		return
	}

	items := make([]*dSeedListRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, buildDSeedListRowFromModel(row))
	}

	pi := common.BuildPageInfo(c, total, q.Page, q.PageSize, common.PageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":           "DSeed",
		"ViewType":        "d-seeds",
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
	return "/listcard?" + q.Encode()
}

func buildMovieDetailHrefByName(movieName string) string {
	name := strings.TrimSpace(movieName)
	if name == "" {
		return ""
	}
	return "/moviedetail/" + url.PathEscape(name)
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

type dSeedUpsertReq struct {
	Name       string `json:"name"`
	Active     int64  `json:"active"`
	SearchType int64  `json:"search_type"`
	NameType   int64  `json:"name_type"`
	PageNow    int64  `json:"page_now"`
	Offset     int64  `json:"offset"`
	StartPage  int64  `json:"start_page"`
	EndPage    int64  `json:"end_page"`
	LastStatus int64  `json:"last_status"`
	LastError  string `json:"last_error"`
}

func (h *DSeedAPI) Create(c *gin.Context) {
	var req dSeedUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	row, err := buildDSeedModelForCreate(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	if _, err := h.deps.SeedModel.FindOneByName(c.Request.Context(), row.Name); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "查询名已存在"})
		return
	} else if !errors.Is(err, spiderx.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	res, err := h.deps.SeedModel.Insert(c.Request.Context(), row)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	if id > 0 {
		row.Id = id
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "d_seed 条目创建成功",
		"item":    buildDSeedListRowFromModel(row),
	})
}

func (h *DSeedAPI) Update(c *gin.Context) {
	id, err := parseDSeedID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "无效的 d_seed id"})
		return
	}

	row, err := h.deps.SeedModel.FindOne(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, spiderx.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "条目不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if row == nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "条目不存在"})
		return
	}

	var req dSeedUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	if err := applyDSeedUpdate(row, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	exist, err := h.deps.SeedModel.FindOneByName(c.Request.Context(), row.Name)
	if err == nil && exist != nil && exist.Id != row.Id {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "查询名已存在"})
		return
	} else if err != nil && !errors.Is(err, spiderx.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	if err := h.deps.SeedModel.Update(c.Request.Context(), row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "d_seed 条目更新成功",
		"item":    buildDSeedListRowFromModel(row),
	})
}

func parseDSeedID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid d_seed id")
	}
	return id, nil
}

func buildDSeedModelForCreate(req dSeedUpsertReq) (*spiderx.DSeed, error) {
	now := time.Now().Unix()
	row := &spiderx.DSeed{
		LastQueryTime: 0,
		CreatedOn:     now,
		UpdatedOn:     now,
	}
	if err := applyDSeedUpdate(row, req); err != nil {
		return nil, err
	}
	return row, nil
}

func applyDSeedUpdate(row *spiderx.DSeed, req dSeedUpsertReq) error {
	if row == nil {
		return errors.New("条目不存在")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errors.New("查询名不能为空")
	}
	if len(name) > 128 {
		return errors.New("查询名过长")
	}
	if req.Active != spiderx.QueryInactive && req.Active != spiderx.QueryActive {
		return errors.New("启用状态无效")
	}
	if req.SearchType != spiderx.QueryByOffset && req.SearchType != spiderx.QueryByStartEnd {
		return errors.New("查询类型无效")
	}
	if req.NameType != spiderx.QueryNamePrefix && req.NameType != spiderx.QueryNameLabel {
		return errors.New("名称类型无效")
	}
	if req.LastStatus < 0 || req.LastStatus > 3 {
		return errors.New("最后状态无效")
	}
	if req.PageNow < 0 || req.Offset < 0 || req.StartPage < 0 || req.EndPage < 0 {
		return errors.New("分页字段不能为负数")
	}

	row.Name = name
	row.Active = req.Active
	row.SearchType = req.SearchType
	row.NameType = req.NameType
	row.PageNow = req.PageNow
	row.Offset = req.Offset
	row.StartPage = req.StartPage
	row.EndPage = req.EndPage
	row.LastStatus = req.LastStatus
	row.LastError = strings.TrimSpace(req.LastError)
	row.UpdatedOn = time.Now().Unix()
	return nil
}
