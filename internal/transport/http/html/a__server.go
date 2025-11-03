package html

import (
	"rudy_gc/internal/consts"
	"rudy_gc/internal/domain/sc"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"

	"rudy_gc/internal/domain/movie"

	"github.com/gin-gonic/gin"
)

// -------- Handler 结构 --------
type MovieHTMLHandler struct {
	movieSvc   *movie.MovieService
	scSvc      *sc.ScService
	detailJobs chan string // ✅ 单 ID 通道
}

// 依赖注入
func NewMovieHTMLHandler(deps *svc.Deps) *MovieHTMLHandler {
	return &MovieHTMLHandler{
		movieSvc:   movie.NewMovieService(deps),
		scSvc:      sc.NewScService(deps),
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
