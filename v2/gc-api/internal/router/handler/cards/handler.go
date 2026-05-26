package cards

import (
	"net/http"

	"rudy-gc-api/internal/dep"
	cardssvc "rudy-gc-api/internal/service/cards"
	"rudy-gc-api/internal/types"
	"rudy-gc-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func Register(api *gin.RouterGroup, d *dep.Dep) {
	handler := &Handler{service: cardssvc.New(d)}
	api.GET("/cards", handler.List)
}

type Handler struct {
	service *cardssvc.Service
}

func (h *Handler) List(c *gin.Context) {
	var req types.CardsListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	data, err := h.service.List(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "list_cards_failed", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, data)
}
