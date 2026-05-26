package movie

import (
	"net/http"

	"rudy-gc-api/internal/dep"
	moviesvc "rudy-gc-api/internal/service/movie"
	"rudy-gc-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func Register(api *gin.RouterGroup, d *dep.Dep) {
	handler := &Handler{service: moviesvc.New(d)}
	api.GET("/movie/:movie", handler.Detail)
}

type Handler struct {
	service *moviesvc.Service
}

func (h *Handler) Detail(c *gin.Context) {
	movieName := c.Param("movie")
	if movieName == "" {
		response.Fail(c, http.StatusBadRequest, "bad_request", "缺少 movie 参数")
		return
	}
	data, err := h.service.Detail(c.Request.Context(), movieName)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "movie_detail_failed", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, data)
}
