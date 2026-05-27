package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/moviereleaseagg"
	"rudy_gc/internal/service/wmediaagg"
	"rudy_gc/internal/svc"
)

type AggTriggerAPI struct {
	wMediaAggSvc  *wmediaagg.Service
	releaseAggSvc *moviereleaseagg.Service
}

type mediaAggBackfillResponse struct {
	Message string                    `json:"message"`
	Result  *wmediaagg.BackfillResult `json:"result"`
}

type movieReleaseAggBackfillResponse struct {
	Message string                          `json:"message"`
	Result  *moviereleaseagg.BackfillResult `json:"result"`
}

func NewAggTriggerAPI(deps *svc.Deps) *AggTriggerAPI {
	return &AggTriggerAPI{
		wMediaAggSvc:  wmediaagg.NewService(deps),
		releaseAggSvc: moviereleaseagg.NewService(deps),
	}
}

func (h *AggTriggerAPI) BackfillWMediaAgg(c *gin.Context) {
	result, err := h.wMediaAggSvc.BackfillAll(c.Request.Context())
	resp := mediaAggBackfillResponse{
		Message: "WMedia 时间聚合回填完成",
		Result:  result,
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"result": resp,
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AggTriggerAPI) BackfillMovieReleaseAgg(c *gin.Context) {
	result, err := h.releaseAggSvc.BackfillAll(c.Request.Context())
	resp := movieReleaseAggBackfillResponse{
		Message: "上映日时间聚合回填完成",
		Result:  result,
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"result": resp,
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}
