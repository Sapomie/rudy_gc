package page

import (
	"net/http"

	"rudy-gc-api/internal/dep"
	pagesvc "rudy-gc-api/internal/service/page"
	"rudy-gc-api/internal/types"
	"rudy-gc-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func Register(api *gin.RouterGroup, d *dep.Dep) {
	handler := &Handler{service: pagesvc.New(d)}
	api.GET("/pages", handler.Summaries)
	api.GET("/pages/:key", handler.Load)
}

type Handler struct {
	service *pagesvc.Service
}

func (h *Handler) Summaries(c *gin.Context) {
	response.JSON(c, http.StatusOK, h.service.Summaries())
}

func (h *Handler) Load(c *gin.Context) {
	var req types.PageListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	data, err := h.service.Load(c.Request.Context(), c.Param("key"), &req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "load_page_failed", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, data)
}
