package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"rudy_gc/local/api/internal/router/handler/common"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/service/fetchsite"
	"rudy_gc/local/api/internal/service/loop"
	movieservice "rudy_gc/local/api/internal/service/movie"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

type fetchSitePageQuery struct {
	Page            int64    `form:"p"`
	PageSize        int64    `form:"ps"`
	Sort            string   `form:"sort"`
	Order           string   `form:"order"`
	MediaOwned      string   `form:"mowned"`
	Keyword         string   `form:"keyword"`
	Status          string   `form:"status"`
	Statuses        []string `form:"statuses"`
	TriggerSortKey  string   `form:"trigger_sort_key"`
	LastFetchFrom   string   `form:"last_fetch_from"`
	LastFetchTo     string   `form:"last_fetch_to"`
	ReleaseDateFrom string   `form:"release_date_from"`
	ReleaseDateTo   string   `form:"release_date_to"`
	MediaBirthFrom  string   `form:"media_birth_from"`
	MediaBirthTo    string   `form:"media_birth_to"`
}

type fetchSiteSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type fetchSiteSortQuery struct {
	ByMovieName      fetchSiteSortLink
	ByReleaseDate    fetchSiteSortLink
	ByFetchStatus    fetchSiteSortLink
	ByLastFetchTime  fetchSiteSortLink
	ByResultCount    fetchSiteSortLink
	ByHashCount      fetchSiteSortLink
	ByLatestPublish  fetchSiteSortLink
	ByMediaBirthTime fetchSiteSortLink
}

type shSehuatangSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type shSehuatangThreadPageQuery struct {
	Page             int64    `form:"p"`
	PageSize         int64    `form:"ps"`
	Sort             string   `form:"sort"`
	Order            string   `form:"order"`
	Keyword          string   `form:"keyword"`
	ThreadTID        string   `form:"thread_tid"`
	AuthorName       string   `form:"author_name"`
	InfoHash         string   `form:"info_hash"`
	MovieType        string   `form:"movie_type"`
	MovieTypes       []string `form:"movie_types"`
	SourceForumID    string   `form:"source_forum_id"`
	SourceForumIDs   []string `form:"source_forum_ids"`
	DetailStatus     string   `form:"detail_status"`
	DetailStatuses   []string `form:"detail_statuses"`
	PostTimeFrom     string   `form:"post_time_from"`
	PostTimeTo       string   `form:"post_time_to"`
	LastSeenTimeFrom string   `form:"last_seen_time_from"`
	LastSeenTimeTo   string   `form:"last_seen_time_to"`
}

type shSehuatangThreadSortQuery struct {
	ByThreadTID    shSehuatangSortLink
	ByMovieName    shSehuatangSortLink
	ByPostTime     shSehuatangSortLink
	ByDetailStatus shSehuatangSortLink
	ByLastSeenTime shSehuatangSortLink
	ByCreatedOn    shSehuatangSortLink
	ByUpdatedOn    shSehuatangSortLink
}

type shSehuatangThreadPageRow struct {
	*moviex.ShSehuatangThread
	MovieHref string
}

func (h *CrawlerPagesAPI) FetchSiteJavbusListPageMain(c *gin.Context) {
	defer common.WithFrontendPath(c, "/javbuslist")()

	q, pageQuery, statusSelected, ok := h.parseFetchSitePageQuery(c, "JavBus")
	if !ok {
		return
	}

	result, err := h.fetchSite.BuildJavbusPage(c.Request.Context(), fetchsite.JavbusPageQuery(pageQuery))
	if err != nil {
		common.JSONServerError(c, "JavBus 抓取列表加载失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":                  "JavBus 抓取列表",
		"ViewType":               "fetch-site-javbus-list",
		"PageTitle":              "JavBus 抓取列表",
		"PageNote":               "只展示 t_javbus_magnet_fetch，并支持独立筛选、排序、分页和行级触发。",
		"Query":                  q,
		"Rows":                   result.Items,
		"Total":                  result.Total,
		"total":                  result.Total,
		"SuccessCount":           result.SuccessCount,
		"PendingCount":           result.PendingCount,
		"FailedCount":            result.FailedCount,
		"ownedQuery":             common.BuildOwnedFilterInfo(c),
		"PageInfo":               buildFetchSitePageInfo(c, result.Total, result.Page, result.PageSize),
		"SortQuery":              buildFetchSiteSortQuery(c, q.Sort, q.Order),
		"PeerPageURL":            "/sukebeilist",
		"TaskPageURL":            "/javbusfetch",
		"TasksPageURL":           "/crawltasklist",
		"FilteredJavbusTaskType": loop.TaskSpiderFetchJavbusFilter,
		"StatusSelected":         statusSelected,
		"StatusAllActive":        len(statusSelected) == 0,
	})
}

func (h *CrawlerPagesAPI) FetchSiteSukebeiListPageMain(c *gin.Context) {
	defer common.WithFrontendPath(c, "/sukebeilist")()

	q, pageQuery, statusSelected, ok := h.parseFetchSitePageQuery(c, "Sukebei")
	if !ok {
		return
	}

	result, err := h.fetchSite.BuildSukebeiPage(c.Request.Context(), fetchsite.SukebeiPageQuery(pageQuery))
	if err != nil {
		common.JSONServerError(c, "Sukebei 抓取列表加载失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":                   "Sukebei 抓取列表",
		"ViewType":                "fetch-site-sukebei-list",
		"PageTitle":               "Sukebei 抓取列表",
		"PageNote":                "只展示 t_sukebei_torrent_fetch，并支持独立筛选、排序、分页和行级触发。",
		"Query":                   q,
		"Rows":                    result.Items,
		"Total":                   result.Total,
		"total":                   result.Total,
		"SuccessCount":            result.SuccessCount,
		"PendingCount":            result.PendingCount,
		"FailedCount":             result.FailedCount,
		"ownedQuery":              common.BuildOwnedFilterInfo(c),
		"PageInfo":                buildFetchSitePageInfo(c, result.Total, result.Page, result.PageSize),
		"SortQuery":               buildFetchSiteSortQuery(c, q.Sort, q.Order),
		"PeerPageURL":             "/javbuslist",
		"TaskPageURL":             "/sukebeifetch",
		"TasksPageURL":            "/crawltasklist",
		"FilteredSukebeiTaskType": loop.TaskSpiderFetchSukebeiFilter,
		"StatusSelected":          statusSelected,
		"StatusAllActive":         len(statusSelected) == 0,
	})
}

func (h *CrawlerPagesAPI) parseFetchSitePageQuery(c *gin.Context, label string) (fetchSitePageQuery, fetchsite.SukebeiPageQuery, map[string]bool, bool) {
	var q fetchSitePageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.JSONBadRequest(c, "参数解析错误", err)
		return q, fetchsite.SukebeiPageQuery{}, nil, false
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 50
	}
	q.Sort = normalizeFetchSiteSortField(q.Sort)
	q.Order = normalizeFetchSiteSortOrder(q.Order)

	pageQuery := fetchsite.SukebeiPageQuery{
		Page:     q.Page,
		PageSize: q.PageSize,
		Sort:     q.Sort,
		Order:    q.Order,
		Keyword:  strings.TrimSpace(q.Keyword),
	}
	if v, ok, err := parseOwnedFilterValue(q.MediaOwned); err != nil {
		common.JSONBadRequest(c, "WMedia 库存筛选错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.MediaOwned = v
	}
	if values, ok, err := parseFetchSiteStatuses(q.Statuses, q.Status); err != nil {
		common.JSONBadRequest(c, label+" 状态错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.Statuses = values
		pageQuery.HasStatuses = true
	}
	if ts, ok, err := common.ParseOptionalDateStart(q.LastFetchFrom); err != nil {
		common.JSONBadRequest(c, label+" 最后抓取开始日期错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.LastFetchFrom = ts
		pageQuery.HasLastFetchFrom = true
	}
	if ts, ok, err := common.ParseOptionalDateEnd(q.LastFetchTo); err != nil {
		common.JSONBadRequest(c, label+" 最后抓取结束日期错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.LastFetchTo = ts
		pageQuery.HasLastFetchTo = true
	}
	if ts, ok, err := common.ParseOptionalDateStart(q.ReleaseDateFrom); err != nil {
		common.JSONBadRequest(c, label+" 发行时间开始日期错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.ReleaseDateFrom = ts
		pageQuery.HasReleaseDateFrom = true
	}
	if ts, ok, err := common.ParseOptionalDateEnd(q.ReleaseDateTo); err != nil {
		common.JSONBadRequest(c, label+" 发行时间结束日期错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.ReleaseDateTo = ts
		pageQuery.HasReleaseDateTo = true
	}
	if ts, ok, err := common.ParseOptionalDateStart(q.MediaBirthFrom); err != nil {
		common.JSONBadRequest(c, label+" WMedia 下载时间开始日期错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.MediaBirthFrom = ts
		pageQuery.HasMediaBirthFrom = true
	}
	if ts, ok, err := common.ParseOptionalDateEnd(q.MediaBirthTo); err != nil {
		common.JSONBadRequest(c, label+" WMedia 下载时间结束日期错误", err)
		return q, pageQuery, nil, false
	} else if ok {
		pageQuery.MediaBirthTo = ts
		pageQuery.HasMediaBirthTo = true
	}

	return q, pageQuery, buildFetchSiteStatusSelected(q.Statuses, q.Status), true
}

func (h *CrawlerPagesAPI) FetchSiteSehuatangSHListPageMain(c *gin.Context) {
	defer common.WithFrontendPath(c, "/shpool")()

	var q shSehuatangThreadPageQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.JSONBadRequest(c, "参数解析错误", err)
		return
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 50
	}
	q.Sort = normalizeSHSehuatangThreadSortField(q.Sort)
	q.Order = normalizeSHSehuatangSortOrder(q.Order)
	q.SourceForumIDs = normalizeSHMultiQueryValues(q.SourceForumIDs, q.SourceForumID)
	q.DetailStatuses = normalizeSHMultiQueryValues(q.DetailStatuses, q.DetailStatus)
	q.MovieTypes = moviedomain.NormalizeSehuatangThreadMovieTypes(normalizeSHMultiQueryValues(q.MovieTypes, q.MovieType)...)

	filter := moviex.ShSehuatangThreadListFilter{
		Keyword:        strings.TrimSpace(q.Keyword),
		ThreadTID:      strings.TrimSpace(q.ThreadTID),
		AuthorName:     strings.TrimSpace(q.AuthorName),
		InfoHash:       strings.TrimSpace(q.InfoHash),
		MovieTypes:     q.MovieTypes,
		SourceForumIDs: q.SourceForumIDs,
		DetailStatuses: q.DetailStatuses,
	}
	if ts, ok, err := common.ParseOptionalDateStart(q.PostTimeFrom); err != nil {
		common.JSONBadRequest(c, "发帖时间开始日期错误", err)
		return
	} else if ok {
		filter.PostTimeFrom = ts
		filter.HasPostTimeFrom = true
	}
	if ts, ok, err := common.ParseOptionalDateEnd(q.PostTimeTo); err != nil {
		common.JSONBadRequest(c, "发帖时间结束日期错误", err)
		return
	} else if ok {
		filter.PostTimeTo = ts
		filter.HasPostTimeTo = true
	}
	if ts, ok, err := common.ParseOptionalDateStart(q.LastSeenTimeFrom); err != nil {
		common.JSONBadRequest(c, "最近抓到开始日期错误", err)
		return
	} else if ok {
		filter.LastSeenTimeFrom = ts
		filter.HasLastSeenTimeFrom = true
	}
	if ts, ok, err := common.ParseOptionalDateEnd(q.LastSeenTimeTo); err != nil {
		common.JSONBadRequest(c, "最近抓到结束日期错误", err)
		return
	} else if ok {
		filter.LastSeenTimeTo = ts
		filter.HasLastSeenTimeTo = true
	}

	aggregate, err := h.deps.ShSehuatangThreadModel.Aggregate(c.Request.Context(), filter)
	if err != nil {
		common.JSONServerError(c, "Sehuatang SH 列表聚合失败", err)
		return
	}
	total := aggregate.Total
	offset := (q.Page - 1) * q.PageSize
	rows, err := h.deps.ShSehuatangThreadModel.ListPage(c.Request.Context(), offset, q.PageSize, buildSHSehuatangThreadOrderBy(q.Sort, q.Order), filter)
	if err != nil {
		common.JSONServerError(c, "Sehuatang SH 列表加载失败", err)
		return
	}
	fc2MovieIDs := make([]int64, 0)
	for _, row := range rows {
		if row != nil && row.FC2MovieID > 0 {
			fc2MovieIDs = append(fc2MovieIDs, row.FC2MovieID)
		}
	}
	if len(fc2MovieIDs) > 0 {
		fc2MovieRows, err := h.deps.FC2MovieModel.ListByIDs(c.Request.Context(), fc2MovieIDs)
		if err != nil {
			common.JSONServerError(c, "Sehuatang SH 列表 FC2 影片信息加载失败", err)
			return
		}
		applySHThreadFC2Keys(rows, fc2MovieRows)
	}
	pageRows, err := h.buildSHThreadPageRows(c.Request.Context(), rows)
	if err != nil {
		common.JSONServerError(c, "Sehuatang SH 列表影片链接加载失败", err)
		return
	}
	runs, err := h.deps.ShSehuatangFetchRunModel.ListRecent(c.Request.Context(), 10)
	if err != nil {
		common.JSONServerError(c, "Sehuatang SH 抓取记录加载失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":             "Sehuatang SH 帖子列表",
		"ViewType":          "fetch-site-sehuatang-sh-list",
		"PageTitle":         "Sehuatang SH 帖子列表",
		"PageNote":          "展示 sh_thread 帖子池基础信息、详情状态，并保留最近抓取记录。",
		"Query":             q,
		"Rows":              pageRows,
		"RecentRuns":        runs,
		"Total":             total,
		"Aggregate":         aggregate,
		"PageInfo":          buildFetchSitePageInfo(c, total, q.Page, q.PageSize),
		"SortQuery":         buildSHSehuatangThreadSortQuery(c, q.Sort, q.Order),
		"TaskPageURL":       "/shfetch",
		"DetailTaskPageURL": "/shfill",
		"TasksPageURL":      "/crawltasklist",
	})
}

func applySHThreadFC2Keys(rows []*moviex.ShSehuatangThread, movieRows []*moviex.FC2Movie) {
	keyByMovieID := make(map[int64]string, len(movieRows))
	for _, movieRow := range movieRows {
		if movieRow == nil || movieRow.ID <= 0 || strings.TrimSpace(movieRow.FC2Key) == "" {
			continue
		}
		keyByMovieID[movieRow.ID] = strings.TrimSpace(movieRow.FC2Key)
	}
	for _, row := range rows {
		if row == nil || row.FC2MovieID <= 0 {
			continue
		}
		row.FC2Key = keyByMovieID[row.FC2MovieID]
	}
}

func (h *CrawlerPagesAPI) buildSHThreadPageRows(ctx context.Context, rows []*moviex.ShSehuatangThread) ([]*shSehuatangThreadPageRow, error) {
	movieRefs := make([]movieservice.MovieDetailLinkRef, len(rows))
	for index, row := range rows {
		if row == nil || row.MovieType == moviedomain.SehuatangThreadMovieTypeFC2PPV || strings.TrimSpace(row.FC2Key) != "" {
			continue
		}
		movieRefs[index] = movieservice.MovieDetailLinkRef{
			MovieJavID: row.MovieJavId,
			MovieName:  row.MovieName,
		}
	}
	movieHrefs, err := h.movieSvc.ResolveMovieDetailHrefs(ctx, movieRefs)
	if err != nil {
		return nil, err
	}

	out := make([]*shSehuatangThreadPageRow, 0, len(rows))
	for index, row := range rows {
		if row == nil {
			continue
		}
		movieHref := movieHrefs[index]
		if fc2Key := strings.TrimSpace(row.FC2Key); fc2Key != "" {
			movieHref = "/fcdetail/" + url.PathEscape(fc2Key)
		}
		out = append(out, &shSehuatangThreadPageRow{
			ShSehuatangThread: row,
			MovieHref:         movieHref,
		})
	}
	return out, nil
}

func (h *CrawlerPagesAPI) FetchSiteSehuatangSHDetailPageMain(c *gin.Context) {
	defer common.WithFrontendPath(c, "/shthread/"+c.Param("thread_tid"))()

	threadTID := strings.TrimSpace(c.Param("thread_tid"))
	if threadTID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "thread_tid 不能为空"})
		return
	}

	thread, err := h.deps.ShSehuatangThreadModel.FindOneByThreadTID(c.Request.Context(), threadTID)
	if err != nil {
		if err == moviex.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "帖子不存在"})
			return
		}
		common.JSONServerError(c, "Sehuatang SH 帖子详情加载失败", err)
		return
	}
	items, err := h.deps.ShSehuatangFetchItemModel.ListByThreadTID(c.Request.Context(), threadTID, 50)
	if err != nil {
		common.JSONServerError(c, "Sehuatang SH 抓取记录加载失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":           "Sehuatang SH 帖子详情",
		"ViewType":        "fetch-site-sehuatang-sh-detail",
		"PageTitle":       "Sehuatang SH 帖子详情",
		"Thread":          thread,
		"FetchItems":      items,
		"ListPageURL":     "/shpool",
		"TaskPageURL":     "/shfill",
		"ListTaskPageURL": "/shfetch",
		"TasksPageURL":    "/crawltasklist",
	})
}

func buildFetchSitePageInfo(c *gin.Context, total int64, page int64, pageSize int64) *common.PageInfo {
	return common.BuildPageInfo(c, total, page, pageSize, 2)
}

func buildFetchSiteSortQuery(c *gin.Context, currentField string, currentOrder string) *fetchSiteSortQuery {
	makeHref := func(field string) fetchSiteSortLink {
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
		return fetchSiteSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &fetchSiteSortQuery{
		ByMovieName:      makeHref("movie_name"),
		ByReleaseDate:    makeHref("release_date"),
		ByFetchStatus:    makeHref("fetch_status"),
		ByLastFetchTime:  makeHref("last_fetch_time"),
		ByResultCount:    makeHref("last_result_count"),
		ByHashCount:      makeHref("torrent_hash_count"),
		ByLatestPublish:  makeHref("latest_publish_time"),
		ByMediaBirthTime: makeHref("media_birth_time"),
	}
}

func normalizeFetchSiteSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_name", "release_date", "fetch_status", "last_fetch_time", "last_result_count", "torrent_hash_count", "latest_publish_time", "media_birth_time":
		return strings.TrimSpace(raw)
	default:
		return "last_fetch_time"
	}
}

func normalizeFetchSiteSortOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}

func parseOwnedFilterValue(raw string) (int64, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false, err
	}
	if v < 1 || v > 7 {
		return 0, false, fmt.Errorf("must be between 1 and 7")
	}
	return v, true, nil
}

func parseFetchSiteStatuses(raws []string, legacy string) ([]int64, bool, error) {
	merged := make([]string, 0, len(raws)+1)
	for _, raw := range raws {
		current := strings.TrimSpace(raw)
		if current == "" {
			continue
		}
		merged = append(merged, current)
	}
	if len(merged) == 0 {
		current := strings.TrimSpace(legacy)
		if current != "" {
			merged = append(merged, current)
		}
	}
	if len(merged) == 0 {
		return nil, false, nil
	}

	values := make([]int64, 0, len(merged))
	seen := make(map[int64]struct{}, len(merged))
	for _, raw := range merged {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, false, err
		}
		if v < 1 || v > 4 {
			return nil, false, fmt.Errorf("must be between 1 and 4")
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		values = append(values, v)
	}
	return values, len(values) > 0, nil
}

func buildFetchSiteStatusSelected(raws []string, legacy string) map[string]bool {
	selected := make(map[string]bool)
	for _, raw := range raws {
		current := strings.TrimSpace(raw)
		if current == "" {
			continue
		}
		selected[current] = true
	}
	if len(selected) == 0 {
		current := strings.TrimSpace(legacy)
		if current != "" {
			selected[current] = true
		}
	}
	return selected
}

func buildSHSehuatangThreadSortQuery(c *gin.Context, currentField string, currentOrder string) *shSehuatangThreadSortQuery {
	makeHref := func(field string) shSehuatangSortLink {
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
		return shSehuatangSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &shSehuatangThreadSortQuery{
		ByThreadTID:    makeHref("thread_tid"),
		ByMovieName:    makeHref("movie_name"),
		ByPostTime:     makeHref("post_time"),
		ByDetailStatus: makeHref("detail_status"),
		ByLastSeenTime: makeHref("last_seen_time"),
		ByCreatedOn:    makeHref("created_on"),
		ByUpdatedOn:    makeHref("updated_on"),
	}
}

func normalizeSHSehuatangThreadSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "thread_tid", "movie_name", "post_time", "detail_status", "last_seen_time", "created_on", "updated_on":
		return strings.TrimSpace(raw)
	default:
		return "post_time"
	}
}

func normalizeSHSehuatangSortOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}

func buildSHSehuatangThreadOrderBy(sortField string, sortOrder string) string {
	field := normalizeSHSehuatangThreadSortField(sortField)
	order := normalizeSHSehuatangSortOrder(sortOrder)
	return fmt.Sprintf("`%s` %s, `id` DESC", field, strings.ToUpper(order))
}

func normalizeSHMultiQueryValues(values []string, legacy string) []string {
	out := make([]string, 0, len(values)+1)
	seen := make(map[string]struct{}, len(values)+1)
	appendValue := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	for _, value := range values {
		appendValue(value)
	}
	appendValue(legacy)
	return out
}
