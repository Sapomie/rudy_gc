package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/wkv"
	"rudy_gc/internal/svc"
)

type WKvAPI struct {
	wkvSvc *wkv.Service
}

func NewWKvAPI(deps *svc.Deps) *WKvAPI {
	return &WKvAPI{
		wkvSvc: wkv.NewService(deps),
	}
}

type upsertWKvDateRequest struct {
	ItemKey   string `json:"item_key"`
	ItemValue string `json:"item_value"`
}

func (h *WKvAPI) UpsertDate(c *gin.Context) {
	var req upsertWKvDateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	value, err := h.wkvSvc.UpsertDate(c.Request.Context(), req.ItemKey, req.ItemValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "保存成功",
		"item_key":   req.ItemKey,
		"item_value": value,
	})
}
