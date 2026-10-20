package movie

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/router/handler/common"
	movieservice "rudy_gc/local/api/internal/service/movie"
)

func (h *MovieCardAPI) ListMovieCardPeriodRelease(c *gin.Context) {
	h.listMovieCardPeriod(c, "release", "/listcardperiodrelease", "上映周期卡片")
}

func (h *MovieCardAPI) ListMovieCardPeriodBirth(c *gin.Context) {
	h.listMovieCardPeriod(c, "birth", "/listcardperiodbirth", "下载周期卡片")
}

func (h *MovieCardAPI) listMovieCardPeriod(c *gin.Context, basis, path, title string) {
	params := movieservice.CardPeriodParams{
		Basis: basis, PeriodType: "day", PeriodPage: 1, PeriodPageSize: common.DefaultPageSize,
		PeriodSort: "time", PeriodOrder: "desc", Page: 1, PageSize: common.DefaultPageSize,
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "周期卡片参数错误: " + err.Error()})
		return
	}
	params.Basis = basis
	defaultSort := consts.OrderByReleasingDate
	if params.Basis == "birth" {
		defaultSort = consts.OrderByBirthTime
	}
	params.OrderBy = common.NormalizeOrderBy(params.OrderBy, defaultSort)
	result, err := h.movieSvc.BuildCardPeriod(c.Request.Context(), params)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, movieservice.ErrCardPeriodParams) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	// Pin the automatic selection in all subsequent links.
	params = result.Params
	original := *c.Request.URL
	query := original.Query()
	query.Del("basis")
	query.Set("period_type", params.PeriodType)
	query.Set("period_p", fmt.Sprint(params.PeriodPage))
	query.Set("period_ps", fmt.Sprint(params.PeriodPageSize))
	query.Set("period_sort", params.PeriodSort)
	query.Set("period_order", params.PeriodOrder)
	query.Set("key", params.Key)
	query.Set("p", fmt.Sprint(params.Page))
	query.Set("ps", fmt.Sprint(params.PageSize))
	query.Set("od", params.OrderBy)
	query.Set("order", params.Order)
	if params.Basis == "release" {
		query.Set("mowned", fmt.Sprint(params.MediaOwned))
	}
	c.Request.URL.Path, c.Request.URL.RawQuery = path, query.Encode()
	defer func() { *c.Request.URL = original }()
	pageInfo := common.BuildPageInfo(c, result.Total, params.Page, params.PageSize, common.PageWindow)
	periodPageInfo := buildCardPeriodPageInfo(c, result.PeriodTotal, params.PeriodPage, params.PeriodPageSize)
	response := gin.H{
		"Title": title, "Basis": params.Basis, "PeriodType": params.PeriodType,
		"Periods": result.Periods, "PeriodTotal": result.PeriodTotal, "SelectedPeriod": result.SelectedPeriod,
		"movies": result.Movies, "total": result.Total, "PageInfo": pageInfo, "PeriodPageInfo": periodPageInfo,
		"sortQuery": common.BuildSortQuery(c, result.OrderBy), "Query": query,
	}
	if params.Basis == "release" {
		response["ownedQuery"] = common.BuildOwnedFilterInfo(c)
	}
	c.JSON(http.StatusOK, response)
}

func buildCardPeriodPageInfo(c *gin.Context, total, page, size int64) *common.PageInfo {
	info := common.BuildPageInfo(c, total, page, size, common.PageWindow)
	translate := func(href string) string {
		u, _ := url.Parse(href)
		q := u.Query()
		q.Set("period_p", q.Get("p"))
		q.Set("p", "1")
		q.Del("key")
		u.RawQuery = q.Encode()
		return u.String()
	}
	info.StartHref, info.PrevHref, info.NextHref, info.EndHref = translate(info.StartHref), translate(info.PrevHref), translate(info.NextHref), translate(info.EndHref)
	for i := range info.Links {
		if info.Links[i].Href != "" {
			info.Links[i].Href = translate(info.Links[i].Href)
		}
	}
	return info
}
