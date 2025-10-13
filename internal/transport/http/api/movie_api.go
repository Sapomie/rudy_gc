package html

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/domain/movie"
	"rudy_gc/internal/svc"
)

type MovieHTMLHandler struct {
	svc *movie.Service
}

func NewMovieHTMLHandler(deps *svc.Deps) *MovieHTMLHandler {
	return &MovieHTMLHandler{svc: movie.NewMovieService(deps)}
}

// GET /movies?page=1&size=20
// 渲染 ui/templates/movies/index.html
func (h *MovieHTMLHandler) ListMoviesPage(c *gin.Context) {
	page := parseInt64(c.DefaultQuery("page", "1"), 1)
	size := parseInt64(c.DefaultQuery("size", "20"), 20)

	list, total, err := h.svc.ListMovies(c.Request.Context(), page, size)
	if err != nil {
		// 也可以改成 c.HTML(500, "error.html", gin.H{"message": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询电影列表失败: " + err.Error()})
		return
	}

	// 计算分页
	pages := int64((total + size - 1) / size)
	if pages <= 0 {
		pages = 1
	}
	prev := page - 1
	if prev < 1 {
		prev = 1
	}
	next := page + 1
	if next > pages {
		next = pages
	}

	c.HTML(http.StatusOK, "movies/index.html", gin.H{
		"movies":   list,
		"total":    total,
		"page":     page,
		"size":     size,
		"pages":    pages,
		"hasPrev":  page > 1,
		"hasNext":  page < pages,
		"prevPage": prev,
		"nextPage": next,
	})
}

func parseInt64(s string, def int64) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil || v <= 0 {
		return def
	}
	return v
}
