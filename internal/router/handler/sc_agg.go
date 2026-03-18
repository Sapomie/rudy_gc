package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/sc"
	"rudy_gc/internal/types"
)

func (h *MovieHTMLHandler) ScAggPage(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	quarter, _ := strconv.Atoi(c.Param("q"))
	month, _ := strconv.Atoi(c.Param("month"))
	if year < 0 || quarter < 0 || month < 0 {
		c.String(http.StatusBadRequest, "参数不合法")
		return
	}
	if quarter > 0 && (quarter < 1 || quarter > 4) {
		c.String(http.StatusBadRequest, "季度参数不合法")
		return
	}
	if month > 0 && (month < 1 || month > 12) {
		c.String(http.StatusBadRequest, "月份参数不合法")
		return
	}

	topN := 18
	if v := c.Query("topn"); v != "" {
		topN = atoiDef(v, 18)
	} else if v := c.Query("top"); v != "" {
		topN = atoiDef(v, 18)
	}
	if topN < 1 {
		topN = 1
	}
	if topN > 50 {
		topN = 50
	}

	vm, err := h.scSvc.BuildAggView(c.Request.Context(), sc.AggParams{
		Year:    year,
		Quarter: quarter,
		Month:   month,
		TopN:    topN,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "SC 统计加载失败: %v", err)
		return
	}

	c.HTML(http.StatusOK, "page.sc_agg", gin.H{
		"Title":                 vm.Title,
		"Level":                 vm.Level,
		"Breadcrumbs":           vm.Breadcrumbs,
		"TotalEvents":           vm.TotalEvents,
		"TotalMovieAppearances": vm.TotalMovieAppearances,
		"TotalUniqueMovies":     vm.TotalUniqueMovies,
		"BucketsY":              vm.BucketsY,
		"BucketsQ":              vm.BucketsQ,
		"BucketsM":              vm.BucketsM,
		"RecentTrend":           vm.RecentTrend,
		"TopCasts":              vm.TopCasts,
		"TopLabels":             vm.TopLabels,
		"TopPrefixes":           vm.TopPrefixes,
		"Events":                vm.Events,
		"Movies":                vm.Movies,
	})
}

func (h *MovieHTMLHandler) CastDetailPage(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.String(http.StatusBadRequest, "缺少参数: name")
		return
	}

	vm, err := h.scSvc.BuildActorScPage(c.Request.Context(), name, 0, 24)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			c.String(http.StatusNotFound, "未找到演员: %s", name)
			return
		}
		c.String(http.StatusInternalServerError, "演员页加载失败: %v", err)
		return
	}
	if vm == nil || vm.Actor == nil {
		c.String(http.StatusNotFound, "未找到演员: %s", name)
		return
	}

	c.HTML(http.StatusOK, "page.cast_detail", gin.H{
		"Title": vm.Actor.Name,
		"Page":  vm,
	})
}
