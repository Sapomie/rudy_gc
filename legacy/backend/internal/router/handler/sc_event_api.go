package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/service/sc"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"
)

type ScEventAPI struct {
	scSvc *sc.ScService
}

func NewScEventAPI(deps *svc.Deps) *ScEventAPI {
	return &ScEventAPI{
		scSvc: sc.NewService(deps),
	}
}

type scEventUpdateReq struct {
	ComeMovieJavId  string `json:"comeMovieJavId"`
	Kind            string `json:"kind"`
	DurationMinutes int64  `json:"duration"`
	Fg              string `json:"fg"`
	Vessel          string `json:"vessel"`
	MovieCast       string `json:"movieCast"`
	Remarks         string `json:"remarks"`
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
		ComeMovieJavId:  strings.TrimSpace(req.ComeMovieJavId),
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
