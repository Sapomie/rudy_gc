package handler

import (
	"rudy_gc/internal/consts"
	"rudy_gc/internal/service/movie"
	"rudy_gc/internal/service/sc"
	"rudy_gc/internal/service/vfilm"
	"rudy_gc/internal/service/wkv"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"

	"github.com/gin-gonic/gin"
)

// -------- Handler 结构 --------
type MovieHTMLHandler struct {
	deps       *svc.Deps
	movieSvc   *movie.Service
	scSvc      *sc.ScService
	vfilmSvc   *vfilm.Service
	wkvSvc     *wkv.Service
	detailJobs chan string // ✅ 单 ID 通道
}

// 依赖注入
func NewMovieHTMLHandler(deps *svc.Deps) *MovieHTMLHandler {
	return &MovieHTMLHandler{
		deps:       deps,
		movieSvc:   movie.NewService(deps),
		scSvc:      sc.NewService(deps),
		vfilmSvc:   vfilm.NewService(deps),
		wkvSvc:     wkv.NewService(deps),
		detailJobs: deps.DetailJobs, // deps.DetailJobs = make(chan string, 200)
		//bestTrigger: deps.BestTrigger,
	}
}

// /moviecard：按上映日倒序
func (h *MovieHTMLHandler) ListMovieCardFull(c *gin.Context) {
	h.renderMovieCard(c,
		types.ListMovieFullRequest{OrderBy: consts.OrderByReleasingDate},
		"MovieCard", "Movies",
	)
}
