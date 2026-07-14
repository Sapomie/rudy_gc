package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/repository/moviex"
	"rudy_gc/internal/service/shthreadagg"
)

type SHThreadAPI struct {
	deps       *dep.Dep
	aggService *shthreadagg.Service
}

type shThreadBatchDeleteRequest struct {
	ThreadTIDs []string `json:"thread_tids"`
}

func NewSHThreadAPI(deps *dep.Dep) *SHThreadAPI {
	return &SHThreadAPI{
		deps:       deps,
		aggService: shthreadagg.NewService(deps),
	}
}

func (h *SHThreadAPI) BatchDelete(c *gin.Context) {
	var req shThreadBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "删除参数无效: " + err.Error()})
		return
	}
	threadTIDs := normalizeSHThreadDeleteTIDs(req.ThreadTIDs)
	if len(threadTIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要删除的帖子"})
		return
	}

	targetMap, err := h.deps.ShSehuatangThreadModel.FindThreadAggTargetsByTIDs(c.Request.Context(), threadTIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取帖子聚合范围失败: " + err.Error()})
		return
	}
	targets := make([]moviex.ShThreadAggTarget, 0, len(targetMap))
	for _, threadTID := range threadTIDs {
		if target, ok := targetMap[threadTID]; ok {
			targets = append(targets, target)
		}
	}

	deletedCount, err := h.deps.ShSehuatangThreadModel.DeleteByThreadTIDs(c.Request.Context(), threadTIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除帖子失败: " + err.Error()})
		return
	}
	if err := h.aggService.RebuildByTargetsAndLogEvent(c.Request.Context(), "shpool_delete", targets...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":           "帖子已删除，但聚合数据刷新失败: " + err.Error(),
			"requested_count": len(threadTIDs),
			"deleted_count":   deletedCount,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":              true,
		"requested_count": len(threadTIDs),
		"deleted_count":   deletedCount,
	})
}

func normalizeSHThreadDeleteTIDs(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
