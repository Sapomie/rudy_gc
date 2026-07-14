package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/sc"
	"rudy_gc/internal/types"
)

type ScEventAPI struct {
	scSvc *sc.ScService
}

func NewScEventAPI(deps *dep.Dep) *ScEventAPI {
	return &ScEventAPI{
		scSvc: sc.NewService(deps),
	}
}

type scEventUpdateReq struct {
	ComeGListId     int64  `json:"comeGListId"`
	Kind            string `json:"kind"`
	DurationMinutes int64  `json:"duration"`
	Fg              string `json:"fg"`
	Vessel          string `json:"vessel"`
	MovieCast       string `json:"movieCast"`
	Remarks         string `json:"remarks"`
}

type scEventMovieStatusUpdateReq struct {
	IsSc int64 `json:"isSc"`
}

func (h *ScEventAPI) Update(c *gin.Context) {
	name := strings.TrimSpace(c.Param("name"))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 SC 事件名"})
		return
	}

	var req scEventUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}
	if req.DurationMinutes < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "时长不能为负数"})
		return
	}

	item, err := h.scSvc.UpdateEventMeta(c.Request.Context(), &types.ScEventEditForm{
		Name:            name,
		ComeGListId:     req.ComeGListId,
		Kind:            strings.TrimSpace(req.Kind),
		DurationMinutes: req.DurationMinutes,
		Fg:              strings.TrimSpace(req.Fg),
		Vessel:          strings.TrimSpace(req.Vessel),
		MovieCast:       strings.TrimSpace(req.MovieCast),
		Remarks:         strings.TrimSpace(req.Remarks),
	})
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "SC 事件不存在"})
			return
		}
		var badReqErr *types.BadRequestError
		if errors.As(err, &badReqErr) {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": badReqErr.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "SC Event 信息更新成功",
		"item":    item,
	})
}

func (h *ScEventAPI) UpdateMovieStatus(c *gin.Context) {
	name := strings.TrimSpace(c.Param("name"))
	gListID, parseErr := strconv.ParseInt(strings.TrimSpace(c.Param("g_list_id")), 10, 64)
	if name == "" || parseErr != nil || gListID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 SC 事件名或影片记录"})
		return
	}

	var req scEventMovieStatusUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	item, err := h.scSvc.UpdateScEventMovieStatus(c.Request.Context(), name, gListID, req.IsSc)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "SC 事件影片不存在"})
			return
		}
		var badReqErr *types.BadRequestError
		if errors.As(err, &badReqErr) {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": badReqErr.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "SC Event 影片状态更新成功",
		"item":    item,
	})
}
