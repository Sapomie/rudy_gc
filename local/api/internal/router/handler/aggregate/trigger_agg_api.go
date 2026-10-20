package aggregate

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/moviereleaseagg"
	"rudy_gc/local/api/internal/service/shthreadagg"
	"rudy_gc/local/api/internal/service/wmediaagg"
)

type AggTriggerAPI struct {
	wMediaAggSvc   *wmediaagg.Service
	releaseAggSvc  *moviereleaseagg.Service
	shThreadAggSvc *shthreadagg.Service
}

type mediaAggBackfillResponse struct {
	Message string                    `json:"message"`
	Result  *wmediaagg.BackfillResult `json:"result"`
}

type movieReleaseAggBackfillResponse struct {
	Message string                          `json:"message"`
	Result  *moviereleaseagg.BackfillResult `json:"result"`
}

type shThreadAggBackfillResponse struct {
	Message string                      `json:"message"`
	Result  *shthreadagg.BackfillResult `json:"result"`
}

func NewAggTriggerAPI(deps *dep.Dep) *AggTriggerAPI {
	return &AggTriggerAPI{
		wMediaAggSvc:   wmediaagg.NewService(deps),
		releaseAggSvc:  moviereleaseagg.NewService(deps),
		shThreadAggSvc: shthreadagg.NewService(deps),
	}
}

func (h *AggTriggerAPI) Page(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Title":     "聚合回填",
		"ViewType":  "agg-triggers",
		"PageTitle": "聚合回填",
		"PageNote":  "这里处理 WMedia、上映日和 SH 帖子时间聚合的手动回填。",
	})
}

func (h *AggTriggerAPI) BackfillSHThreadAgg(c *gin.Context) {
	result, err := h.shThreadAggSvc.BackfillAll(c.Request.Context())
	resp := shThreadAggBackfillResponse{
		Message: "SH 帖子时间聚合回填完成",
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
