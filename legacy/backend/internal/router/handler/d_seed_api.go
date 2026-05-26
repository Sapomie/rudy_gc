package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/model/modelx/spiderx"
	"rudy_gc/internal/svc"
)

type DSeedAPI struct {
	deps *svc.Deps
}

func NewDSeedAPI(deps *svc.Deps) *DSeedAPI {
	return &DSeedAPI{deps: deps}
}

type dSeedUpsertReq struct {
	Name       string `json:"name"`
	Active     int64  `json:"active"`
	SearchType int64  `json:"search_type"`
	NameType   int64  `json:"name_type"`
	PageNow    int64  `json:"page_now"`
	Offset     int64  `json:"offset"`
	StartPage  int64  `json:"start_page"`
	EndPage    int64  `json:"end_page"`
	LastStatus int64  `json:"last_status"`
	LastError  string `json:"last_error"`
}

func (h *DSeedAPI) Create(c *gin.Context) {
	var req dSeedUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	row, err := buildDSeedModelForCreate(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	if _, err := h.deps.SeedModel.FindOneByName(c.Request.Context(), row.Name); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "查询名已存在"})
		return
	} else if !errors.Is(err, spiderx.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	res, err := h.deps.SeedModel.Insert(c.Request.Context(), row)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	id, _ := res.LastInsertId()
	if id > 0 {
		row.Id = id
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "d_seed 条目创建成功",
		"item":    buildDSeedListRowFromModel(row),
	})
}

func (h *DSeedAPI) Update(c *gin.Context) {
	id, err := parseDSeedID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "无效的 d_seed id"})
		return
	}

	row, err := h.deps.SeedModel.FindOne(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, spiderx.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "条目不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if row == nil {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "条目不存在"})
		return
	}

	var req dSeedUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return
	}

	if err := applyDSeedUpdate(row, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	exist, err := h.deps.SeedModel.FindOneByName(c.Request.Context(), row.Name)
	if err == nil && exist != nil && exist.Id != row.Id {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "查询名已存在"})
		return
	} else if err != nil && !errors.Is(err, spiderx.ErrNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	if err := h.deps.SeedModel.Update(c.Request.Context(), row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "d_seed 条目更新成功",
		"item":    buildDSeedListRowFromModel(row),
	})
}

func parseDSeedID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid d_seed id")
	}
	return id, nil
}

func buildDSeedModelForCreate(req dSeedUpsertReq) (*spiderx.DSeed, error) {
	now := time.Now().Unix()
	row := &spiderx.DSeed{
		LastQueryTime: 0,
		CreatedOn:     now,
		UpdatedOn:     now,
	}
	if err := applyDSeedUpdate(row, req); err != nil {
		return nil, err
	}
	return row, nil
}

func applyDSeedUpdate(row *spiderx.DSeed, req dSeedUpsertReq) error {
	if row == nil {
		return errors.New("条目不存在")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errors.New("查询名不能为空")
	}
	if len(name) > 128 {
		return errors.New("查询名过长")
	}
	if req.Active != spiderx.QueryInactive && req.Active != spiderx.QueryActive {
		return errors.New("启用状态无效")
	}
	if req.SearchType != spiderx.QueryByOffset && req.SearchType != spiderx.QueryByStartEnd {
		return errors.New("查询类型无效")
	}
	if req.NameType != spiderx.QueryNamePrefix && req.NameType != spiderx.QueryNameLabel {
		return errors.New("名称类型无效")
	}
	if req.LastStatus < 0 || req.LastStatus > 3 {
		return errors.New("最后状态无效")
	}
	if req.PageNow < 0 || req.Offset < 0 || req.StartPage < 0 || req.EndPage < 0 {
		return errors.New("分页字段不能为负数")
	}

	row.Name = name
	row.Active = req.Active
	row.SearchType = req.SearchType
	row.NameType = req.NameType
	row.PageNow = req.PageNow
	row.Offset = req.Offset
	row.StartPage = req.StartPage
	row.EndPage = req.EndPage
	row.LastStatus = req.LastStatus
	row.LastError = strings.TrimSpace(req.LastError)
	row.UpdatedOn = time.Now().Unix()
	return nil
}
