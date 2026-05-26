package rank

import (
	"net/http"
	"strconv"

	"rudy-gc-api/internal/consts"
	"rudy-gc-api/internal/dep"
	ranksvc "rudy-gc-api/internal/service/rank"
	"rudy-gc-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func Register(api *gin.RouterGroup, d *dep.Dep) {
	handler := &Handler{service: ranksvc.New(d)}
	api.GET("/rank/day", handler.Day)
	api.GET("/rank/period", handler.Period)
}

type Handler struct {
	service *ranksvc.Service
}

func (h *Handler) Day(c *gin.Context) {
	date := c.Query("date")
	page := parseInt64(c.DefaultQuery("p", "1"), 1)
	pageSize := parseInt64(c.DefaultQuery("ps", "18"), 18)
	data, err := h.service.Day(c.Request.Context(), date, page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "rank_day_failed", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, data)
}

func (h *Handler) Period(c *gin.Context) {
	typeName := c.DefaultQuery("type", consts.RankPeriodTypeName(consts.RankPeriodTypeMonth))
	category := consts.ParseCategory(c.DefaultQuery("category", "1"), consts.BestCategoryMonth)
	page := parseInt64(c.DefaultQuery("p", "1"), 1)
	pageSize := parseInt64(c.DefaultQuery("ps", "18"), 18)
	key := c.Query("key")
	data, err := h.service.Period(c.Request.Context(), typeName, category, page, pageSize, key)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "rank_period_failed", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, data)
}

func parseInt64(raw string, fallback int64) int64 {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
