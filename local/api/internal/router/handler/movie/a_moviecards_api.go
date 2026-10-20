package movie

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"rudy_gc/local/api/internal/router/handler/common"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/service/movie"
	"rudy_gc/local/api/internal/service/wkv"
	"rudy_gc/local/api/internal/types"
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
		"/listcard",
	)
}

func (h *MovieCardAPI) ListMovieCardToday(c *gin.Context) {
	h.movieCardAPI(c,
		types.ListMovieFullRequest{
			OrderBy:          consts.OrderByReleasingDate,
			ReleasingDateEnd: time.Now().Format("2006-01-02"),
		},
		"MovieCard",
		"/listcardtoday",
	)
}

func (h *MovieCardAPI) ListMovieCardHasRank(c *gin.Context) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = "/listcardrank"
	defer func() {
		c.Request.URL.Path = originalPath
	}()

	req, err := common.ParseMovieCardRequest(c, types.ListMovieFullRequest{
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
		"/listcardowned",
	)
}

func (h *MovieCardAPI) ListMovieCardNeedDownload(c *gin.Context) {
	h.movieCardAPI(c,
		types.ListMovieFullRequest{
			NeedDownload: consts.MovieNeedDownLoadOK,
			OrderBy:      consts.OrderByReleasingDate,
		},
		"MovieCard",
		"/listcardneed",
	)
}

func (h *MovieCardAPI) movieCardAPI(c *gin.Context, base types.ListMovieFullRequest, title, frontendPath string) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = frontendPath
	defer func() {
		c.Request.URL.Path = originalPath
	}()

	req, err := common.ParseMovieCardRequest(c, base)
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
	c.Request.URL.Path = "/listcardrandom"
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
	if n > common.MaxPageSize {
		n = common.MaxPageSize
	}

	curOD := common.NormalizeOrderBy(c.Query("od"), req.OrderBy)
	req.OrderBy = curOD
	req.Order = common.NormalizeOrder(c.Query("order"))

	resp, err := h.movieSvc.ListMovieFullRandom(c.Request.Context(), &req, int64(n))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	h.enqueueJavIDsNonBlocking(resp.JavIds)

	pi := common.BuildPageInfo(c, resp.Total, 1, int64(n), common.PageWindow)
	ownedQ := common.BuildOwnedFilterInfo(c)
	sortQ := common.BuildSortQuery(c, curOD)

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
		"MovieCardFilter": common.BuildMovieCardFilterView(c, req, curOD, &n),
	})
}

func (h *MovieCardAPI) ListMovieCardAgg(c *gin.Context) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = "/listcardagg"
	defer func() {
		c.Request.URL.Path = originalPath
	}()

	req, err := common.ParseMovieCardRequest(c, types.ListMovieFullRequest{OrderBy: consts.OrderByReleasingDate})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数解析错误: " + err.Error()})
		return
	}

	filterView := common.BuildMovieCardFilterView(c, req, req.OrderBy, nil)
	filterView.Action = "/listcardagg"
	filterView.ClearHref = "/listcardagg"

	if !hasMovieCardAggEffectiveFilter(c) {
		c.JSON(http.StatusOK, gin.H{
			"Title":           "MovieCard Agg",
			"ViewType":        "cards-agg",
			"total":           0,
			"Total":           0,
			"TotalLabel":      "--",
			"Skipped":         true,
			"SkipReason":      "no_filter",
			"Message":         "请选择筛选条件后再执行聚合。",
			"Limit":           movie.CardsAggMaxMovieCount,
			"ownedQuery":      common.BuildOwnedFilterInfo(c),
			"sortQuery":       common.BuildSortQuery(c, req.OrderBy),
			"CurrentSort":     req.OrderBy,
			"MovieCardFilter": filterView,
		})
		return
	}

	agg, err := h.movieSvc.BuildMovieCardsAgg(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":           "MovieCard Agg",
		"ViewType":        "cards-agg",
		"total":           agg.Total,
		"Total":           agg.Total,
		"Skipped":         agg.Skipped,
		"SkipReason":      agg.SkipReason,
		"Message":         agg.Message,
		"Limit":           agg.Limit,
		"Summary":         agg.Summary,
		"Groups":          agg.Groups,
		"Timeline":        agg.Timeline,
		"Distributions":   agg.Distributions,
		"TopMovies":       agg.TopMovies,
		"ownedQuery":      common.BuildOwnedFilterInfo(c),
		"sortQuery":       common.BuildSortQuery(c, req.OrderBy),
		"CurrentSort":     req.OrderBy,
		"MovieCardFilter": filterView,
	})
}

func hasMovieCardAggEffectiveFilter(c *gin.Context) bool {
	q := c.Request.URL.Query()
	filterKeys := []string{
		"cn", "pid", "gn", "dn", "pn", "mn", "ln", "lj", "an", "wd",
		"rs", "re", "mbs", "mbe", "srds", "srde", "lsstmin", "lsstmax",
		"cay", "cao", "vwmin", "vwmax", "smin", "smax",
		"drkmin", "nd", "mowned", "scmin", "scmax", "comin", "comax",
		"md1", "md2", "md3", "md4",
	}
	for _, key := range filterKeys {
		values, ok := q[key]
		if !ok {
			continue
		}
		for _, raw := range values {
			value := strings.TrimSpace(raw)
			if value == "" {
				continue
			}
			if key == "mowned" && value == "1" {
				continue
			}
			if key == "nd" && value == "0" {
				continue
			}
			return true
		}
	}
	return false
}

func (h *MovieCardAPI) ListMovieCardDayRank(c *gin.Context) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = "/listcarddayrank"
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
	if pageSize <= 0 || pageSize > common.MaxPageSize {
		pageSize = common.DefaultPageSize
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

	pi := common.BuildPageInfo(c, total, page, pageSize, common.PageWindow)

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
	c.Request.URL.Path = "/listcardperiodrank"
	defer func() {
		c.Request.URL.Path = originalPath
	}()

	typeName := strings.TrimSpace(c.DefaultQuery("type", consts.RankPeriodTypeName(consts.RankPeriodTypeMonth)))
	periodType := consts.RankPeriodTypeFromName(typeName)
	category := parseInt64OrDefault(c.DefaultQuery("category", strconv.FormatInt(consts.BestCategoryMonth, 10)), consts.BestCategoryMonth)
	periodKey := strings.TrimSpace(c.Query("key"))
	page := parseInt64OrDefault(c.DefaultQuery("p", "1"), 1)
	pageSize := parseInt64OrDefault(c.DefaultQuery("ps", strconv.FormatInt(common.DefaultPageSize, 10)), common.DefaultPageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > common.MaxPageSize {
		pageSize = common.DefaultPageSize
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

	pi := common.BuildPageInfo(c, view.Total, page, pageSize, common.PageWindow)
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
	PageInfo        *common.PageInfo
	OwnedQuery      *common.OwnedQuery
	SortQuery       *common.SortQuery
	CurrentSort     string
	MovieCardFilter *common.MovieCardFilterView
	PageDateEdit    *movieCardPageDateEditView
}

type movieCardPageDateEditView struct {
	ItemKey      string
	Label        string
	Value        string
	DisplayValue string
}

func (h *MovieCardAPI) loadMovieCardPageData(c *gin.Context, req types.ListMovieFullRequest, action, clearHref string) (*movieCardPageData, error) {
	resp, err := h.movieSvc.ListMovieFull(c.Request.Context(), &req)
	if err != nil {
		return nil, err
	}

	h.enqueueJavIDsNonBlocking(resp.JavIds)

	pi := common.BuildPageInfo(c, resp.Total, req.Page, req.PageSize, common.PageWindow)
	ownedQ := common.BuildOwnedFilterInfo(c)
	sortQ := common.BuildSortQuery(c, req.OrderBy)
	filterView := common.BuildMovieCardFilterView(c, req, req.OrderBy, nil)
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
	for _, id := range common.UniqueNonEmpty(ids) {
		select {
		case h.detailJobs <- id:
		default:
			continue
		}
	}
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
