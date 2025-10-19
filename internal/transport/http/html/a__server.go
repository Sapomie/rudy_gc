package html

import (
	"rudy_gc/internal/svc"

	"rudy_gc/internal/domain/movie"
)

// -------- Handler 结构 --------
type MovieHTMLHandler struct {
	svc        *movie.Service
	detailJobs chan string // ✅ 单 ID 通道
	//bestTrigger chan struct{}
}

// 依赖注入
func NewMovieHTMLHandler(deps *svc.Deps) *MovieHTMLHandler {
	return &MovieHTMLHandler{
		svc:        movie.NewMovieService(deps),
		detailJobs: deps.DetailJobs, // deps.DetailJobs = make(chan string, 200)
		//bestTrigger: deps.BestTrigger,
	}
}
