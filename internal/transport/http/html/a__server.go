package html

import (
	"rudy_gc/internal/domain/sc"
	"rudy_gc/internal/svc"

	"rudy_gc/internal/domain/movie"
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
