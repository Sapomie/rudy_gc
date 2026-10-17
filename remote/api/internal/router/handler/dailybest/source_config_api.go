package dailybest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	dailybestservice "rudy_gc/remote/api/internal/service/dailybest"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

func (h *DailyBestAPI) GetSourceConfig(c *gin.Context) {
	config, err := h.service.GetSourceConfig(c.Request.Context())
	if err != nil {
		writeServiceError(c, "source_config", err)
		return
	}
	c.JSON(http.StatusOK, config)
}

func (h *DailyBestAPI) UpdateSourceConfig(c *gin.Context) {
	var input dailybesttypes.UpdateSourceConfigRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_source_config", "请求参数无效")
		return
	}
	config, err := h.service.UpdateSourceConfig(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, dailybestservice.ErrInvalidSourceConfig) {
			writeError(c, http.StatusBadRequest, "invalid_source_config", err.Error())
			return
		}
		writeServiceError(c, "source_config", err)
		return
	}
	c.JSON(http.StatusOK, config)
}
