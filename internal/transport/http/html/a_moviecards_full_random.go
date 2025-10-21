package html

import (
	"net/http"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"strconv"

	"github.com/gin-gonic/gin"
)

/* ======================== 随机卡片 ======================== */

// /moviecard/random?n=6
func (h *MovieHTMLHandler) ListMovieCardFullRandom(c *gin.Context) {
	// 基础默认值与常规列表一致
	req := types.ListMovieFullRequest{OrderBy: consts.OrderByReleasingDate}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}

	// n：随机抽取条数（默认 6，上限沿用页面上限）
	const defaultN = 6
	nStr := c.DefaultQuery("n", strconv.Itoa(defaultN))
	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		n = defaultN
	}
	if n > maxPageSize {
		n = maxPageSize
	}

	// 排序字段仍允许外部传（但随机抽样不依赖排序）
	curOD := normalizeOrderBy(c.Query("od"), req.OrderBy)
	req.OrderBy = curOD

	resp, err := h.movieSvc.ListMovieFullRandom(c.Request.Context(), &req, int64(n))
	if err != nil {
		c.String(http.StatusInternalServerError, "查询失败: %v", err)
		return
	}

	// 异步投递 javIds（非阻塞 + 去重）
	h.enqueueJavIDsNonBlocking(resp.JavIds)

	// 展示为“随机页”：页号固定 1
	pi := BuildPageInfo(c, resp.Total, 1, int64(n), pageWindow)
	ownedQ := buildOwnedFilterInfo(c)
	sortQ := buildSortQuery(c, curOD)

	c.HTML(http.StatusOK, "page.list_movie_card", gin.H{
		"Title":       "MovieCard (Random)",
		"movies":      resp.List,
		"total":       resp.Total,
		"PageInfo":    pi,
		"pageInfo":    pi,
		"ownedQuery":  ownedQ,
		"sortQuery":   sortQ,
		"CurrentSort": curOD,
	})
}
