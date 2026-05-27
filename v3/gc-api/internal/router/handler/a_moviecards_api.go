package handler

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/dep"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/movie"
	"rudy_gc/internal/service/wkv"
	"rudy_gc/internal/types"
)

const (
	defaultPageSize = 18
	maxPageSize     = 20000
	pageWindow      = 3
)

type MovieCardAPI struct {
	movieSvc   *movie.Service
	wkvSvc     *wkv.Service
	detailJobs chan string
}

func NewMovieCardAPI(deps *dep.Dep) *MovieCardAPI {
	return &MovieCardAPI{
		movieSvc:   movie.NewService(deps),
		wkvSvc:     wkv.NewService(deps),
		detailJobs: deps.DetailJobs,
	}
}

func (h *MovieCardAPI) ListMovieCardFull(c *gin.Context) {
	h.movieCardAPI(c,
		types.ListMovieFullRequest{OrderBy: consts.OrderByReleasingDate},
		"MovieCard",
		"/cards",
	)
}

func (h *MovieCardAPI) ListMovieCardToday(c *gin.Context) {
	h.movieCardAPI(c,
		types.ListMovieFullRequest{
			OrderBy:          consts.OrderByReleasingDate,
			ReleasingDateEnd: time.Now().Format("2006-01-02"),
		},
		"MovieCard",
		"/cardstoday",
	)
}

func (h *MovieCardAPI) ListMovieCardHasRank(c *gin.Context) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = "/cardshasrank"
	defer func() {
		c.Request.URL.Path = originalPath
	}()

	req, err := parseMovieCardRequest(c, types.ListMovieFullRequest{
		OrderBy:       consts.OrderByRankDate,
		DaysInRankMin: 1,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数解析错误: " + err.Error()})
		return
	}

	data, err := h.loadMovieCardPageData(c, req, "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	hrkTimeValue, err := h.wkvSvc.GetValue(c.Request.Context(), wkv.ItemKeyHRKTime)
	if err != nil {
		hrkTimeValue = ""
	}
	data.PageDateEdit = &movieCardPageDateEditView{
		ItemKey:      wkv.ItemKeyHRKTime,
		Label:        "HRK时间",
		Value:        hrkTimeValue,
		DisplayValue: displayWKvDateValue(hrkTimeValue),
	}

	h.respondMovieCardPage(c, "MovieCard", data)
}

func (h *MovieCardAPI) ListMovieCardMediaOwned(c *gin.Context) {
	h.movieCardAPI(c,
		types.ListMovieFullRequest{
			MediaOwned: consts.OwnedAllNotRemoved,
			OrderBy:    consts.OrderByMediaBirthTime,
		},
		"MovieCard",
		"/cardsmediamowned",
	)
}

func (h *MovieCardAPI) ListMovieCardNeedDownload(c *gin.Context) {
	h.movieCardAPI(c,
		types.ListMovieFullRequest{
			NeedDownload: consts.MovieNeedDownLoadOK,
			OrderBy:      consts.OrderByReleasingDate,
		},
		"MovieCard",
		"/cardsneeddownload",
	)
}

func (h *MovieCardAPI) movieCardAPI(c *gin.Context, base types.ListMovieFullRequest, title, frontendPath string) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = frontendPath
	defer func() {
		c.Request.URL.Path = originalPath
	}()

	req, err := parseMovieCardRequest(c, base)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数解析错误: " + err.Error()})
		return
	}

	data, err := h.loadMovieCardPageData(c, req, "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	h.respondMovieCardPage(c, title, data)
}

func (h *MovieCardAPI) ListMovieCardFullRandom(c *gin.Context) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = "/cardsrandom"
	defer func() {
		c.Request.URL.Path = originalPath
	}()

	req := types.ListMovieFullRequest{OrderBy: consts.OrderByReleasingDate}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数解析错误: " + err.Error()})
		return
	}

	const defaultN = 6
	nStr := c.DefaultQuery("n", strconv.Itoa(defaultN))
	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		n = defaultN
	}
	if n > maxPageSize {
		n = maxPageSize
	}

	curOD := normalizeOrderBy(c.Query("od"), req.OrderBy)
	req.OrderBy = curOD

	resp, err := h.movieSvc.ListMovieFullRandom(c.Request.Context(), &req, int64(n))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	h.enqueueJavIDsNonBlocking(resp.JavIds)

	pi := BuildPageInfo(c, resp.Total, 1, int64(n), pageWindow)
	ownedQ := buildOwnedFilterInfo(c)
	sortQ := buildSortQuery(c, curOD)

	c.JSON(http.StatusOK, gin.H{
		"Title":           "MovieCard (Random)",
		"ViewType":        "cards",
		"movies":          resp.List,
		"total":           resp.Total,
		"PageInfo":        pi,
		"pageInfo":        pi,
		"ownedQuery":      ownedQ,
		"sortQuery":       sortQ,
		"CurrentSort":     curOD,
		"MovieCardFilter": buildMovieCardFilterView(c, req, curOD, &n),
	})
}

func (h *MovieCardAPI) ListMovieCardDayRank(c *gin.Context) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = "/moviecarddayrank"
	defer func() {
		c.Request.URL.Path = originalPath
	}()

	ctx := c.Request.Context()
	dateStr := strings.TrimSpace(c.Query("date"))

	minDay, err := h.movieSvc.FindEarliestRankDayNumber(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询最早排行日期失败: " + err.Error()})
		return
	}
	maxDay, err := h.movieSvc.FindLatestRankDayNumber(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询最近排行日期失败: " + err.Error()})
		return
	}
	if maxDay <= 0 {
		c.JSON(http.StatusOK, gin.H{"Title": "MovieCard Rank", "ViewType": "rank-day", "message": "暂无排行数据"})
		return
	}
	if minDay <= 0 {
		minDay = 1
	}

	var dayNumber int64
	if dateStr == "" {
		dayNumber = maxDay
		dateStr = consts.GetDateStringByRankDayNumber(dayNumber)
	} else {
		dayNumber = consts.GetRankDayNumber(dateStr)
		if dayNumber < minDay {
			dayNumber = minDay
		}
		if dayNumber > maxDay {
			dayNumber = maxDay
		}
		dateStr = consts.GetDateStringByRankDayNumber(dayNumber)
	}

	page := parseInt64OrDefault(c.DefaultQuery("p", "1"), 1)
	pageSize := parseInt64OrDefault(c.DefaultQuery("ps", "0"), 0)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = defaultPageSize
	}

	movies, total, err := h.movieSvc.ListMovieTypesByRankDay(ctx, dayNumber, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询排行影片失败: " + err.Error()})
		return
	}

	prevDay := dayNumber - 1
	if prevDay < minDay {
		prevDay = minDay
	}
	nextDay := dayNumber + 1
	if nextDay > maxDay {
		nextDay = maxDay
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randDay := minDay
	if maxDay > minDay {
		randDay = rng.Int63n(maxDay-minDay+1) + minDay
	}

	pi := BuildPageInfo(c, total, page, pageSize, pageWindow)

	c.JSON(http.StatusOK, gin.H{
		"Title":        fmt.Sprintf("MovieCard Rank %s", dateStr),
		"ViewType":     "rank-day",
		"movies":       movies,
		"total":        total,
		"PageInfo":     pi,
		"pageInfo":     pi,
		"rankDate":     dateStr,
		"prevDate":     consts.GetDateStringByRankDayNumber(prevDay),
		"nextDate":     consts.GetDateStringByRankDayNumber(nextDay),
		"randDate":     consts.GetDateStringByRankDayNumber(randDay),
		"prevDisabled": dayNumber <= minDay,
		"nextDisabled": dayNumber >= maxDay,
	})
}

func (h *MovieCardAPI) ListMovieCardPeriodRank(c *gin.Context) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = "/moviecardperiodrank"
	defer func() {
		c.Request.URL.Path = originalPath
	}()

	typeName := strings.TrimSpace(c.DefaultQuery("type", consts.RankPeriodTypeName(consts.RankPeriodTypeMonth)))
	periodType := consts.RankPeriodTypeFromName(typeName)
	category := parseInt64OrDefault(c.DefaultQuery("category", strconv.FormatInt(consts.BestCategoryMonth, 10)), consts.BestCategoryMonth)
	periodKey := strings.TrimSpace(c.Query("key"))
	page := parseInt64OrDefault(c.DefaultQuery("p", "1"), 1)
	pageSize := parseInt64OrDefault(c.DefaultQuery("ps", strconv.FormatInt(defaultPageSize, 10)), defaultPageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > maxPageSize {
		pageSize = defaultPageSize
	}

	view, err := h.movieSvc.BuildRankPeriodPage(c.Request.Context(), movie.RankPeriodPageRequest{
		PeriodType: periodType,
		PeriodKey:  periodKey,
		Category:   category,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			c.JSON(http.StatusOK, gin.H{"Title": "MovieCard Period Rank", "ViewType": "rank-period", "message": "暂无周期排行数据"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询周期排行失败: " + err.Error()})
		return
	}

	h.enqueueJavIDsNonBlocking(rankPeriodCardJavIDs(view.Cards))

	pi := BuildPageInfo(c, view.Total, page, pageSize, pageWindow)
	currentTypeName := consts.RankPeriodTypeName(view.Period.PeriodType)
	currentCategory := view.Period.Category

	c.JSON(http.StatusOK, gin.H{
		"Title":           view.Title,
		"ViewType":        "rank-period",
		"cards":           view.Cards,
		"total":           view.Total,
		"PageInfo":        pi,
		"pageInfo":        pi,
		"periodKey":       view.Period.PeriodKey,
		"periodTypeLabel": view.PeriodTypeLabel,
		"categoryLabel":   view.CategoryLabel,
		"rangeStart":      view.RangeStart,
		"rangeEnd":        view.RangeEnd,
		"typeLinks":       buildRankPeriodTypeLinks(c, currentTypeName, currentCategory),
		"categoryLinks":   buildRankPeriodCategoryLinks(c, currentTypeName, currentCategory),
		"latestHref":      buildRankPeriodHref(c.Request.URL.Path, currentTypeName, currentCategory, ""),
		"prevHref":        rankPeriodNavHref(c.Request.URL.Path, currentTypeName, currentCategory, view.PrevPeriod),
		"nextHref":        rankPeriodNavHref(c.Request.URL.Path, currentTypeName, currentCategory, view.NextPeriod),
		"prevDisabled":    view.PrevPeriod == nil,
		"nextDisabled":    view.NextPeriod == nil,
	})
}

func (h *MovieCardAPI) respondMovieCardPage(c *gin.Context, title string, data *movieCardPageData) {
	c.JSON(http.StatusOK, gin.H{
		"Title":           title,
		"ViewType":        "cards",
		"movies":          data.Movies,
		"total":           data.Total,
		"PageInfo":        data.PageInfo,
		"pageInfo":        data.PageInfo,
		"ownedQuery":      data.OwnedQuery,
		"sortQuery":       data.SortQuery,
		"CurrentSort":     data.CurrentSort,
		"MovieCardFilter": data.MovieCardFilter,
		"PageDateEdit":    data.PageDateEdit,
	})
}

type movieCardPageData struct {
	Movies          []*types.MovieType
	Total           int64
	PageInfo        *PageInfo
	OwnedQuery      *OwnedQuery
	SortQuery       *SortQuery
	CurrentSort     string
	MovieCardFilter *movieCardFilterView
	PageDateEdit    *movieCardPageDateEditView
}

type movieCardPageDateEditView struct {
	ItemKey      string
	Label        string
	Value        string
	DisplayValue string
}

func parseMovieCardRequest(c *gin.Context, base types.ListMovieFullRequest) (types.ListMovieFullRequest, error) {
	req := base
	if err := c.ShouldBindQuery(&req); err != nil {
		return types.ListMovieFullRequest{}, err
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > maxPageSize {
		req.PageSize = defaultPageSize
	}
	req.OrderBy = normalizeOrderBy(c.Query("od"), req.OrderBy)
	req.Order = normalizeOrder(c.Query("order"))
	return req, nil
}

func (h *MovieCardAPI) loadMovieCardPageData(c *gin.Context, req types.ListMovieFullRequest, action, clearHref string) (*movieCardPageData, error) {
	resp, err := h.movieSvc.ListMovieFull(c.Request.Context(), &req)
	if err != nil {
		return nil, err
	}

	h.enqueueJavIDsNonBlocking(resp.JavIds)

	pi := BuildPageInfo(c, resp.Total, req.Page, req.PageSize, pageWindow)
	ownedQ := buildOwnedFilterInfo(c)
	sortQ := buildSortQuery(c, req.OrderBy)
	filterView := buildMovieCardFilterView(c, req, req.OrderBy, nil)
	if action != "" {
		filterView.Action = action
	}
	if clearHref != "" {
		filterView.ClearHref = clearHref
	}

	return &movieCardPageData{
		Movies:          resp.List,
		Total:           resp.Total,
		PageInfo:        pi,
		OwnedQuery:      ownedQ,
		SortQuery:       sortQ,
		CurrentSort:     req.OrderBy,
		MovieCardFilter: filterView,
	}, nil
}

func (h *MovieCardAPI) enqueueJavIDsNonBlocking(ids []string) {
	if h.detailJobs == nil || len(ids) == 0 {
		return
	}
	for _, id := range uniqueNonEmpty(ids) {
		select {
		case h.detailJobs <- id:
		default:
			continue
		}
	}
}

func uniqueNonEmpty(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s = strings.TrimSpace(s); s == "" {
			continue
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

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

func buildOwnedFilterInfo(c *gin.Context) *OwnedQuery {
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
	consts.OrderByDetailUpdateTime: {},
	consts.OrderByCastAgeDesc:      {},
	consts.OrderByCastAgeAsc:       {},
	consts.OrderByViewerWatched:    {},
	consts.OrderByReleasingDate:    {},
	consts.OrderByRankDate:         {},
	consts.OrderByHighestRank:      {},
	consts.OrderByDaysInRank:       {},
	consts.OrderByBirthTime:        {},
	consts.OrderByMediaBirthTime:   {},
	consts.OrderByScTimes:          {},
	consts.OrderByComeTimes:        {},
	consts.OrderByLastScTime:       {},
}

func normalizeOrderBy(od string, fallback string) string {
	if _, ok := allowedOrderBy[od]; ok {
		return od
	}
	return fallback
}

func normalizeOrder(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "asc", "desc":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return ""
	}
}

type SortLink struct {
	Href   string
	Active bool
}

type SortQuery struct {
	ByDetailUpdate   SortLink
	ByReleasing      SortLink
	ByViewer         SortLink
	ByDaysInRank     SortLink
	ByHighestRank    SortLink
	ByRankDate       SortLink
	ByScTimes        SortLink
	ByLastScTime     SortLink
	ByBirthTime      SortLink
	ByMediaBirthTime SortLink
	ByCastAgeDesc    SortLink
	ByCastAgeAsc     SortLink
	ByComeTimes      SortLink
}

func buildSortQuery(c *gin.Context, currentOD string) *SortQuery {
	makeHref := func(od string) (href string, active bool) {
		q := cloneValues(c)
		q.Set("od", od)
		q.Set("p", "1")
		href = c.Request.URL.Path
		if enc := q.Encode(); enc != "" {
			href += "?" + enc
		}
		return href, currentOD == od
	}

	du, a0 := makeHref(consts.OrderByDetailUpdateTime)
	rd, a1 := makeHref(consts.OrderByReleasingDate)
	vw, a2 := makeHref(consts.OrderByViewerWatched)
	drk, a3 := makeHref(consts.OrderByDaysInRank)
	hrk, a4 := makeHref(consts.OrderByHighestRank)
	rk, a5 := makeHref(consts.OrderByRankDate)
	sc, a6 := makeHref(consts.OrderByScTimes)
	lsct, a7 := makeHref(consts.OrderByLastScTime)
	bt, a8 := makeHref(consts.OrderByBirthTime)
	mbt, a9 := makeHref(consts.OrderByMediaBirthTime)
	cad, a10 := makeHref(consts.OrderByCastAgeDesc)
	caa, a11 := makeHref(consts.OrderByCastAgeAsc)
	co, a12 := makeHref(consts.OrderByComeTimes)

	return &SortQuery{
		ByDetailUpdate:   SortLink{Href: du, Active: a0},
		ByReleasing:      SortLink{Href: rd, Active: a1},
		ByViewer:         SortLink{Href: vw, Active: a2},
		ByDaysInRank:     SortLink{Href: drk, Active: a3},
		ByHighestRank:    SortLink{Href: hrk, Active: a4},
		ByRankDate:       SortLink{Href: rk, Active: a5},
		ByScTimes:        SortLink{Href: sc, Active: a6},
		ByLastScTime:     SortLink{Href: lsct, Active: a7},
		ByBirthTime:      SortLink{Href: bt, Active: a8},
		ByMediaBirthTime: SortLink{Href: mbt, Active: a9},
		ByCastAgeDesc:    SortLink{Href: cad, Active: a10},
		ByCastAgeAsc:     SortLink{Href: caa, Active: a11},
		ByComeTimes:      SortLink{Href: co, Active: a12},
	}
}

func cloneValues(c *gin.Context) url.Values {
	orig := c.Request.URL.Query()
	q := make(url.Values, len(orig))
	for k, vv := range orig {
		cp := make([]string, len(vv))
		copy(cp, vv)
		q[k] = cp
	}
	return q
}

type movieCardFilterView struct {
	Action    string
	ClearHref string

	OrderBy  string
	Order    string
	PageSize string

	RandomCount    string
	Explicit       map[string]bool
	HideDirs       bool
	TextDateInputs bool

	CastNames    string
	PersonIds    string
	GenreNames   string
	DirectorName string
	PrefixName   string
	MakerName    string
	LabelName    string
	LabelJavID   string
	AlbumName    string
	Word         string

	ReleasingDateStart  string
	ReleasingDateEnd    string
	MediaBirthTimeStart string
	MediaBirthTimeEnd   string

	CastAgeMin string
	CastAgeMax string

	StartRankingDateStart string
	StartRankingDateEnd   string

	DaysInRankMin string
	NeedDownload  string
	MediaOwned    string

	ViewWatchedMin string
	ViewWatchedMax string
	ScoreMin       string
	ScoreMax       string

	LastScTimeMin string
	LastScTimeMax string
	ScTimesMin    string
	ScTimesMax    string
	ComeTimesMin  string
	ComeTimesMax  string

	MediaDir1 string
	MediaDir2 string
	MediaDir3 string
	MediaDir4 string
}

func buildMovieCardFilterView(c *gin.Context, req types.ListMovieFullRequest, currentOD string, randomN *int) *movieCardFilterView {
	path := c.Request.URL.Path
	view := &movieCardFilterView{
		Action:         path,
		ClearHref:      buildMovieCardFilterClearHref(path, randomN),
		OrderBy:        currentOD,
		Order:          queryOrFallbackString(c, "order", req.Order),
		PageSize:       strconv.FormatInt(req.PageSize, 10),
		Explicit:       buildMovieCardFilterExplicit(c, randomN != nil),
		HideDirs:       false,
		TextDateInputs: true,

		CastNames:    queryOrFallbackString(c, "cn", req.CastNames),
		PersonIds:    queryOrFallbackString(c, "pid", req.PersonIds),
		GenreNames:   queryOrFallbackString(c, "gn", req.GenreNames),
		DirectorName: queryOrFallbackString(c, "dn", req.DirectorName),
		PrefixName:   queryOrFallbackString(c, "pn", req.PrefixName),
		MakerName:    queryOrFallbackString(c, "mn", req.MakerName),
		LabelName:    queryOrFallbackString(c, "ln", req.LabelName),
		LabelJavID:   queryOrFallbackString(c, "lj", req.LabelJavID),
		AlbumName:    queryOrFallbackString(c, "an", req.AlbumName),
		Word:         queryOrFallbackString(c, "wd", req.Word),

		ReleasingDateStart:  queryOrFallbackString(c, "rs", req.ReleasingDateStart),
		ReleasingDateEnd:    queryOrFallbackString(c, "re", req.ReleasingDateEnd),
		MediaBirthTimeStart: queryOrFallbackString(c, "mbs", req.MediaBirthTimeStart),
		MediaBirthTimeEnd:   queryOrFallbackString(c, "mbe", req.MediaBirthTimeEnd),

		CastAgeMin: queryOrFallbackFloat(c, "cay", req.CastAgeMin),
		CastAgeMax: queryOrFallbackFloat(c, "cao", req.CastAgeMax),

		StartRankingDateStart: queryOrFallbackString(c, "srds", req.StartRankingDateStart),
		StartRankingDateEnd:   queryOrFallbackString(c, "srde", req.StartRankingDateEnd),

		DaysInRankMin: queryOrFallbackInt(c, "drkmin", req.DaysInRankMin, false),
		NeedDownload:  queryOrFallbackInt(c, "nd", req.NeedDownload, false),
		MediaOwned:    queryOrFallbackInt(c, "mowned", req.MediaOwned, false),

		ViewWatchedMin: queryOrFallbackInt(c, "vwmin", req.ViewWatchedMin, false),
		ViewWatchedMax: queryOrFallbackInt(c, "vwmax", req.ViewWatchedMax, false),
		ScoreMin:       queryOrFallbackFloat(c, "smin", req.ScoreMin),
		ScoreMax:       queryOrFallbackFloat(c, "smax", req.ScoreMax),

		LastScTimeMin: queryOrFallbackString(c, "lsctmin", req.LastScTimeMin),
		LastScTimeMax: queryOrFallbackString(c, "lsctmax", req.LastScTimeMax),
		ScTimesMin:    queryOrFallbackInt(c, "scmin", req.ScTimesMin, false),
		ScTimesMax:    queryOrFallbackIntPtr(c, "scmax", req.ScTimesMax),
		ComeTimesMin:  queryOrFallbackInt(c, "comin", req.ComeTimesMin, false),
		ComeTimesMax:  queryOrFallbackIntPtr(c, "comax", req.ComeTimesMax),

		MediaDir1: queryOrFallbackString(c, "md1", req.MediaDir1),
		MediaDir2: queryOrFallbackString(c, "md2", req.MediaDir2),
		MediaDir3: queryOrFallbackString(c, "md3", req.MediaDir3),
		MediaDir4: queryOrFallbackString(c, "md4", req.MediaDir4),
	}

	if randomN != nil {
		view.RandomCount = strconv.Itoa(*randomN)
	}
	return view
}

func buildMovieCardFilterExplicit(c *gin.Context, hasRandom bool) map[string]bool {
	keys := []string{
		"ps", "od", "order",
		"cn", "pid", "gn", "dn", "pn", "mn", "ln", "lj", "an", "wd",
		"nd", "drkmin",
		"mowned",
		"md1", "md2", "md3", "md4",
		"rs", "re", "mbs", "mbe", "srds", "srde", "lsctmin", "lsctmax",
		"cay", "cao", "vwmin", "vwmax", "smin", "smax", "scmin", "scmax", "comin", "comax",
	}
	if hasRandom {
		keys = append(keys, "n")
	}

	out := make(map[string]bool, len(keys))
	for _, key := range keys {
		_, ok := c.GetQuery(key)
		out[key] = ok
	}
	return out
}

func buildMovieCardFilterClearHref(path string, randomN *int) string {
	if randomN == nil {
		return path
	}
	return path + "?n=" + strconv.Itoa(*randomN)
}

func queryOrFallbackString(c *gin.Context, key, fallback string) string {
	if v, ok := c.GetQuery(key); ok {
		return strings.TrimSpace(v)
	}
	return fallback
}

func queryOrFallbackInt(c *gin.Context, key string, fallback int64, showZeroFallback bool) string {
	if v, ok := c.GetQuery(key); ok {
		return strings.TrimSpace(v)
	}
	if fallback == 0 && !showZeroFallback {
		return ""
	}
	return strconv.FormatInt(fallback, 10)
}

func queryOrFallbackIntPtr(c *gin.Context, key string, fallback *int64) string {
	if v, ok := c.GetQuery(key); ok {
		return strings.TrimSpace(v)
	}
	if fallback == nil {
		return ""
	}
	return strconv.FormatInt(*fallback, 10)
}

func queryOrFallbackFloat(c *gin.Context, key string, fallback float64) string {
	if v, ok := c.GetQuery(key); ok {
		return strings.TrimSpace(v)
	}
	if fallback == 0 {
		return ""
	}
	return strconv.FormatFloat(fallback, 'f', -1, 64)
}

func displayWKvDateValue(v string) string {
	if strings.TrimSpace(v) == "" {
		return "未设置"
	}
	return strings.TrimSpace(v)
}

func parseInt64OrDefault(s string, def int64) int64 {
	if s == "" {
		return def
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return n
}

type rankPeriodSwitch struct {
	Label  string
	Href   string
	Active bool
}

func buildRankPeriodTypeLinks(c *gin.Context, currentType string, category int64) []rankPeriodSwitch {
	periodTypes := []int64{
		consts.RankPeriodTypeWeek,
		consts.RankPeriodTypeMonth,
		consts.RankPeriodTypeQuarter,
		consts.RankPeriodTypeYear,
	}
	out := make([]rankPeriodSwitch, 0, len(periodTypes))
	for _, periodType := range periodTypes {
		typeName := consts.RankPeriodTypeName(periodType)
		out = append(out, rankPeriodSwitch{
			Label:  consts.RankPeriodTypeLabel(periodType),
			Href:   buildRankPeriodHref(c.Request.URL.Path, typeName, category, ""),
			Active: typeName == currentType,
		})
	}
	return out
}

func buildRankPeriodCategoryLinks(c *gin.Context, currentType string, currentCategory int64) []rankPeriodSwitch {
	categories := []int64{
		consts.BestCategoryMonth,
		consts.BestCategoryAllTime,
	}
	out := make([]rankPeriodSwitch, 0, len(categories))
	for _, category := range categories {
		out = append(out, rankPeriodSwitch{
			Label:  consts.BestCategoryLabel(category),
			Href:   buildRankPeriodHref(c.Request.URL.Path, currentType, category, ""),
			Active: category == currentCategory,
		})
	}
	return out
}

func rankPeriodNavHref(path, typeName string, category int64, period *moviex.CRankPeriod) string {
	if period == nil {
		return ""
	}
	return buildRankPeriodHref(path, typeName, category, period.PeriodKey)
}

func buildRankPeriodHref(path, typeName string, category int64, key string) string {
	q := url.Values{}
	q.Set("type", typeName)
	q.Set("category", strconv.FormatInt(category, 10))
	if key != "" {
		q.Set("key", key)
	}
	return fmt.Sprintf("%s?%s", path, q.Encode())
}

func rankPeriodCardJavIDs(cards []*movie.RankPeriodMovieCard) []string {
	out := make([]string, 0, len(cards))
	for _, card := range cards {
		if card == nil || card.Movie == nil {
			continue
		}
		out = append(out, card.Movie.JavId)
	}
	return out
}
