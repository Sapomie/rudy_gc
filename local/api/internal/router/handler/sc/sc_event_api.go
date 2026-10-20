package sc

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/sc"
	"rudy_gc/local/api/internal/types"
)

type ScEventAPI struct {
	scSvc *sc.Service
}

func NewScEventAPI(deps *dep.Dep) *ScEventAPI {
	return &ScEventAPI{
		scSvc: sc.NewService(deps),
	}
}

type scEventUpdateReq struct {
	ComeGListId       int64  `json:"comeGListId"`
	KindOptionID      int64  `json:"kindOptionId"`
	DurationMinutes   int64  `json:"duration"`
	FGOptionID        int64  `json:"fgOptionId"`
	VesselOptionID    int64  `json:"vesselOptionId"`
	MovieCastPersonID int64  `json:"movieCastPersonId"`
	Remarks           string `json:"remarks"`
}

type scEventMovieStatusUpdateReq struct {
	IsSc int64 `json:"isSc"`
}

func (h *ScEventAPI) Delete(c *gin.Context) {
	name := strings.TrimSpace(c.Param("name"))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 SC 事件名"})
		return
	}

	result, err := h.scSvc.DeleteEvent(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "SC 事件不存在"})
			return
		}
		response := gin.H{"ok": false, "error": err.Error()}
		if result != nil {
			response["deleted"] = true
			response["name"] = result.Name
			response["deletedMovieCount"] = result.DeletedMovieCount
			response["derivedDataFresh"] = result.DerivedDataFresh
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":                true,
		"deleted":           true,
		"name":              result.Name,
		"deletedMovieCount": result.DeletedMovieCount,
		"derivedDataFresh":  result.DerivedDataFresh,
		"message":           "SC Event 删除成功",
	})
}

func (h *ScEventAPI) Update(c *gin.Context) {
	name := strings.TrimSpace(c.Param("name"))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少 SC 事件名"})
		return
	}

	var req scEventUpdateReq
	var image io.Reader
	if c.ContentType() == "multipart/form-data" {
		form, err := c.MultipartForm()
		if form != nil {
			defer form.RemoveAll()
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "读取图片请求失败"})
			return
		}
		if err := json.Unmarshal([]byte(c.PostForm("data")), &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
			return
		}
		images := form.File["image"]
		if len(images) > 1 {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "每个事件选择一张图片"})
			return
		}
		if len(images) == 1 {
			file, err := images[0].Open()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "读取图片失败"})
				return
			}
			defer file.Close()
			image = file
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
			return
		}
	}
	if req.DurationMinutes < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "时长不能为负数"})
		return
	}

	item, err := h.scSvc.UpdateEventMeta(c.Request.Context(), &types.ScEventEditForm{
		Name:              name,
		ComeGListId:       req.ComeGListId,
		KindOptionID:      req.KindOptionID,
		DurationMinutes:   req.DurationMinutes,
		FGOptionID:        req.FGOptionID,
		VesselOptionID:    req.VesselOptionID,
		MovieCastPersonID: req.MovieCastPersonID,
		Remarks:           strings.TrimSpace(req.Remarks),
		Image:             image,
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
