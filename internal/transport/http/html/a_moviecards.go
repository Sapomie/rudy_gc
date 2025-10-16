package html

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

// 分页参数常量
const (
	defaultPageSize = 18
	maxPageSize     = 200
	pageWindow      = 3
)

// ===== 具体页面：只传差异化默认参数 =====

// /moviecard：按上映日倒序
func (h *MovieHTMLHandler) ListMovieCardFull(c *gin.Context) {
	h.renderMovieCard(c,
		types.ListMovieFullRequest{
			OrderBy: consts.OrderByReleasingDate,
		},
		"MovieCard",
		"Movies",
	)
}

func (h *MovieHTMLHandler) ListMovieCardToday(c *gin.Context) {

	h.renderMovieCard(c,
		types.ListMovieFullRequest{
			OrderBy:          consts.OrderByReleasingDate,
			ReleasingDateEnd: time.Now().Format("2006-01-02"),
		},
		"MovieCard",
		"Movies",
	)
}

// /moviecardrank：在榜（≥1 天），按榜单日期倒序
func (h *MovieHTMLHandler) ListMovieCardHasRank(c *gin.Context) {
	h.renderMovieCard(c,
		types.ListMovieFullRequest{
			OrderBy:       consts.OrderByRankDate,
			DaysInRankMin: 1,
		},
		"MovieCard",
		"Movies",
	)
}

// /moviecardowned：仅已拥有（不含已移除），按拍摄/生成时间倒序
func (h *MovieHTMLHandler) ListMovieCardOwned(c *gin.Context) {
	h.renderMovieCard(c,
		types.ListMovieFullRequest{
			Owned:   consts.OwnedAllNotRemoved,
			OrderBy: consts.OrderByBirthTime,
		},
		"MovieCard",
		"Movies",
	)
}

// /moviecardneeddownload：需要下载 OK，按上映日倒序
func (h *MovieHTMLHandler) ListMovieCardNeedDownload(c *gin.Context) {
	h.renderMovieCard(c,
		types.ListMovieFullRequest{
			NeedDownload: consts.MovieNeedDownLoadOK,
			OrderBy:      consts.OrderByReleasingDate,
		},
		"MovieCard",
		"Movies",
	)
}

// 统一的渲染函数：把公共流程收敛到一处
func (h *MovieHTMLHandler) renderMovieCard(c *gin.Context, base types.ListMovieFullRequest, title, fieldName string) {
	req := base
	if err := c.ShouldBindQuery(&req); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > maxPageSize {
		req.PageSize = defaultPageSize
	}

	// ✅ 读取 od 参数，并对非法值回落到默认
	curOD := normalizeOrderBy(c.Query("od"), req.OrderBy)
	req.OrderBy = curOD

	resp, err := h.svc.ListMovieFull(c.Request.Context(), &req)
	if err != nil {
		c.String(http.StatusInternalServerError, "查询失败: %v", err)
		return
	}

	pi := BuildPageInfo(c, resp.Total, req.Page, req.PageSize, pageWindow) // ⚠️ 建议你的 BuildPageInfo 用 p/ps
	ownedQ := buildOwnedFilterInfo(c)
	sortQ := buildSortQuery(c, curOD)

	c.HTML(http.StatusOK, "base", gin.H{
		"Title":       title,
		"movies":      resp.List,
		"Total":       resp.Total,
		"PageInfo":    pi,
		"pageInfo":    pi,
		"ownedQuery":  ownedQ,
		"sortQuery":   sortQ, // ✅ 模板可用
		"CurrentSort": curOD, // （可选）模板也可直接用
		"total":       resp.Total,
		"fieldName":   fieldName,
	})
}
