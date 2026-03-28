package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/movie"
	"rudy_gc/internal/svc"
	"rudy_gc/internal/types"
)

type EItemAPI struct {
	deps     *svc.Deps
	movieSvc *movie.Service
}

func NewEItemAPI(deps *svc.Deps) *EItemAPI {
	return &EItemAPI{
		deps:     deps,
		movieSvc: movie.NewService(deps),
	}
}

type eItemStatusUpdateReq struct {
	HasDetail        *string `json:"hasDetail"`
	HasDownloadCover *string `json:"hasDownloadCover"`
	HasChinese       *string `json:"hasChinese"`
	DetailNeedScan   *string `json:"detailNeedScan"`
}

type eItemBatchStatusUpdateReq struct {
	Ids []int64 `json:"ids"`
	eItemStatusUpdateReq
}

func (h *EItemAPI) Delete(c *gin.Context) {
	itemID, err := parseEItemID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "无效的 item id"})
		return
	}

	item, err := h.deps.ItemModel.FindOne(c.Request.Context(), itemID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "条目不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "条目不存在"})
		return
	}

	if err := h.movieSvc.DeleteMovieByJavID(c.Request.Context(), item.JavId, item.Name, "e_items_page"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":    true,
		"id":    item.Id,
		"javId": item.JavId,
		"name":  item.Name,
	})
}

func (h *EItemAPI) UpdateStatus(c *gin.Context) {
	itemID, err := parseEItemID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "无效的 item id"})
		return
	}

	item, err := h.deps.ItemModel.FindOne(c.Request.Context(), itemID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "条目不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "条目不存在"})
		return
	}

	var req eItemStatusUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid request"})
		return
	}

	changed, err := applyEItemStatusUpdate(item, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	if changed {
		item.UpdatedOn = time.Now().Unix()
		if err := h.deps.ItemModel.Update(c.Request.Context(), item); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"item": buildItemListRow(mapEItemModelToType(item), h.deps.Config.Fetcher.JavAddress),
	})
}

func (h *EItemAPI) UpdateStatusBatch(c *gin.Context) {
	var req eItemBatchStatusUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "invalid request"})
		return
	}

	ids, err := normalizeEItemIDs(req.Ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if req.HasDetail == nil && req.HasDownloadCover == nil && req.HasChinese == nil && req.DetailNeedScan == nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "缺少状态字段"})
		return
	}

	items := make([]*moviex.EItem, 0, len(ids))
	for _, id := range ids {
		item, err := h.deps.ItemModel.FindOne(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "条目不存在"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
			return
		}
		if item == nil {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "条目不存在"})
			return
		}
		items = append(items, item)
	}

	rows := make([]*types.ItemListRow, 0, len(items))
	for _, item := range items {
		changed, err := applyEItemStatusUpdate(item, req.eItemStatusUpdateReq)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
			return
		}
		if changed {
			item.UpdatedOn = time.Now().Unix()
			if err := h.deps.ItemModel.Update(c.Request.Context(), item); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
				return
			}
		}
		rows = append(rows, buildItemListRow(mapEItemModelToType(item), h.deps.Config.Fetcher.JavAddress))
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":    true,
		"items": rows,
	})
}

func parseEItemID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func normalizeEItemIDs(ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return nil, errors.New("缺少条目 id")
	}

	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, errors.New("存在无效的 item id")
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func applyEItemStatusUpdate(item *moviex.EItem, req eItemStatusUpdateReq) (bool, error) {
	if req.HasDetail == nil && req.HasDownloadCover == nil && req.HasChinese == nil && req.DetailNeedScan == nil {
		return false, errors.New("缺少状态字段")
	}

	changed := false

	if req.HasDetail != nil {
		v, err := parseRequiredItemDetailFilter(*req.HasDetail)
		if err != nil {
			return false, err
		}
		if item.HasDetail != v {
			item.HasDetail = v
			changed = true
		}
	}

	if req.HasDownloadCover != nil {
		v, err := parseRequiredItemCoverFilter(*req.HasDownloadCover)
		if err != nil {
			return false, err
		}
		if item.HasDownloadCover != v {
			item.HasDownloadCover = v
			changed = true
		}
	}

	if req.HasChinese != nil {
		v, err := parseRequiredItemChineseFilter(*req.HasChinese)
		if err != nil {
			return false, err
		}
		if item.HasChinese != v {
			item.HasChinese = v
			changed = true
		}
	}

	if req.DetailNeedScan != nil {
		v, err := parseRequiredItemNeedScanFilter(*req.DetailNeedScan)
		if err != nil {
			return false, err
		}
		if item.DetailNeedScan != v {
			item.DetailNeedScan = v
			changed = true
		}
	}

	return changed, nil
}

func mapEItemModelToType(item *moviex.EItem) *types.Item {
	if item == nil {
		return nil
	}

	return &types.Item{
		Id:                  item.Id,
		Name:                item.Name,
		JavId:               item.JavId,
		Prefix:              item.Prefix,
		SearchType:          item.SearchType,
		CoverUrl:            item.CoverUrl,
		SearchBy:            item.SearchBy,
		HasDetail:           item.HasDetail,
		HasDownloadCover:    item.HasDownloadCover,
		HasChinese:          item.HasChinese,
		DetailNeedScan:      item.DetailNeedScan,
		DetailBirthTime:     item.DetailBirthTime,
		LastQueryDetailTime: item.LastQueryDetailTime,
		CreatedOn:           item.CreatedOn,
		UpdatedOn:           item.UpdatedOn,
	}
}
