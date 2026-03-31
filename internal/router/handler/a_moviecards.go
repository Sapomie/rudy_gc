// internal/handler/html/movie_card_handlers.go
package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

// 分页参数常量
const (
	defaultPageSize = 18
	maxPageSize     = 20000
	pageWindow      = 3
)

/* ======================== 页面入口 ======================== */

// /moviecardtoday：只显示今天前上映
func (h *MovieHTMLHandler) ListMovieCardToday(c *gin.Context) {
	h.renderMovieCard(c,
		types.ListMovieFullRequest{
			OrderBy:          consts.OrderByReleasingDate,
			ReleasingDateEnd: time.Now().Format("2006-01-02"),
		},
		"MovieCard", "Movies",
	)
}

// /moviecardrank：在榜（≥1 天），按榜单日期倒序
func (h *MovieHTMLHandler) ListMovieCardHasRank(c *gin.Context) {
	h.renderMovieCard(c,
		types.ListMovieFullRequest{
			OrderBy:       consts.OrderByRankDate,
			DaysInRankMin: 1,
		},
		"MovieCard", "Movies",
	)
}

// /moviecardowned：仅已拥有，按拍摄/生成时间倒序
func (h *MovieHTMLHandler) ListMovieCardOwned(c *gin.Context) {
	h.renderMovieCard(c,
		types.ListMovieFullRequest{
			Owned:   consts.OwnedAllNotRemoved,
			OrderBy: consts.OrderByBirthTime,
		},
		"MovieCard", "Movies",
	)
}

// /cardsmediamowned：仅 media 已拥有，按 media 下载时间倒序
func (h *MovieHTMLHandler) ListMovieCardMediaOwned(c *gin.Context) {
	h.renderMovieCard(c,
		types.ListMovieFullRequest{
			MediaOwned: consts.OwnedAllNotRemoved,
			OrderBy:    consts.OrderByMediaBirthTime,
		},
		"MovieCard", "Movies",
	)
}

// /moviecardneeddownload：需要下载 OK，按上映日倒序
func (h *MovieHTMLHandler) ListMovieCardNeedDownload(c *gin.Context) {
	h.renderMovieCard(c,
		types.ListMovieFullRequest{
			NeedDownload: consts.MovieNeedDownLoadOK,
			OrderBy:      consts.OrderByReleasingDate,
		},
		"MovieCard", "Movies",
	)
}

/* ======================== 渲染核心 ======================== */

type movieCardPageData struct {
	Movies          []*types.MovieType
	Total           int64
	PageInfo        *PageInfo
	OwnedQuery      *OwnedQuery
	SortQuery       *SortQuery
	CurrentSort     string
	MovieCardFilter *movieCardFilterView
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

func (h *MovieHTMLHandler) loadMovieCardPageData(c *gin.Context, req types.ListMovieFullRequest, action, clearHref string) (*movieCardPageData, error) {
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

func (h *MovieHTMLHandler) renderMovieCard(c *gin.Context, base types.ListMovieFullRequest, title, fieldName string) {
	req, err := parseMovieCardRequest(c, base)
	if err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}

	data, err := h.loadMovieCardPageData(c, req, "", "")
	if err != nil {
		c.String(http.StatusInternalServerError, "查询失败: %v", err)
		return
	}

	c.HTML(http.StatusOK, "page.list_movie_card", gin.H{
		"Title":           title,
		"movies":          data.Movies,
		"total":           data.Total,
		"PageInfo":        data.PageInfo,
		"pageInfo":        data.PageInfo,
		"ownedQuery":      data.OwnedQuery,
		"sortQuery":       data.SortQuery,
		"CurrentSort":     data.CurrentSort,
		"MovieCardFilter": data.MovieCardFilter,
	})
}

/* ======================== 工具函数 ======================== */

// 去重 + 非阻塞逐个发送
func (h *MovieHTMLHandler) enqueueJavIDsNonBlocking(ids []string) {
	if h.detailJobs == nil || len(ids) == 0 {
		return
	}
	for _, id := range uniqueNonEmpty(ids) {
		select {
		case h.detailJobs <- id:
		default:
			continue // 通道满时跳过，避免阻塞
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
