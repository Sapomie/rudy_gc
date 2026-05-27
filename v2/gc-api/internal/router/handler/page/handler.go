package page

import (
	"net/http"
	"strings"

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
	req.Extra = c.Request.URL.Query()
	if strings.TrimSpace(req.Query) == "" {
		req.Query = firstQueryValue(req.Extra, "keyword")
	}
	if strings.TrimSpace(req.OrderBy) == "" {
		req.OrderBy = firstQueryValue(req.Extra, "sort")
	}
	if strings.TrimSpace(req.Order) == "" {
		req.Order = firstQueryValue(req.Extra, "dir")
	}
	data, err := h.service.Load(c.Request.Context(), c.Param("key"), &req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "load_page_failed", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, data)
}

func firstQueryValue(values map[string][]string, key string) string {
	if len(values) == 0 {
		return ""
	}
	items := values[key]
	if len(items) == 0 {
		return ""
	}
	return strings.TrimSpace(items[len(items)-1])
}
