package sc

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/sc"
	"rudy_gc/local/api/internal/types"
)

type ScOptionAPI struct {
	service *sc.Service
}

type scOptionWriteReq struct {
	OptionName string `json:"optionName"`
	SortOrder  int64  `json:"sortOrder"`
	IsEnabled  *int64 `json:"isEnabled"`
}

func NewScOptionAPI(deps *dep.Dep) *ScOptionAPI {
	return &ScOptionAPI{service: sc.NewService(deps)}
}

func (h *ScOptionAPI) List(c *gin.Context) {
	items, err := h.service.ListScOptions(c.Request.Context(), strings.TrimSpace(c.Query("includeDisabled")) == "1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groupScOptions(items))
}

func (h *ScOptionAPI) Create(c *gin.Context) {
	form, ok := bindScOptionWrite(c, c.Param("type"))
	if !ok {
		return
	}
	item, err := h.service.CreateScOption(c.Request.Context(), form)
	if err != nil {
		writeScOptionError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true, "message": "SC 选项新增成功", "item": item})
}

func (h *ScOptionAPI) Update(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "SC 选项 ID 无效"})
		return
	}
	form, ok := bindScOptionWrite(c, c.Param("type"))
	if !ok {
		return
	}
	item, err := h.service.UpdateScOption(c.Request.Context(), id, form)
	if err != nil {
		writeScOptionError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "SC 选项更新成功", "item": item})
}

func (h *ScOptionAPI) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "SC 选项 ID 无效"})
		return
	}
	item, err := h.service.DisableScOption(c.Request.Context(), c.Param("type"), id)
	if err != nil {
		writeScOptionError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "disabled": true, "message": "SC 选项已停用", "item": item})
}

func bindScOptionWrite(c *gin.Context, optionType string) (*types.ScOptionWriteForm, bool) {
	var req scOptionWriteReq
	if err := c.ShouldBindJSON(&req); err != nil || req.IsEnabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请求参数无效"})
		return nil, false
	}
	return &types.ScOptionWriteForm{
		OptionType: strings.TrimSpace(optionType),
		OptionName: req.OptionName,
		SortOrder:  req.SortOrder,
		IsEnabled:  *req.IsEnabled,
	}, true
}

func writeScOptionError(c *gin.Context, err error) {
	if errors.Is(err, types.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "SC 选项不存在"})
		return
	}
	var badRequest *types.BadRequestError
	if errors.As(err, &badRequest) {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": badRequest.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
}

func groupScOptions(items []*types.ScOption) gin.H {
	grouped := map[string][]*types.ScOption{
		sc.ScOptionTypeKind:   {},
		sc.ScOptionTypeFG:     {},
		sc.ScOptionTypeVessel: {},
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		if _, ok := grouped[item.OptionType]; ok {
			grouped[item.OptionType] = append(grouped[item.OptionType], item)
		}
	}

	return gin.H{
		"ok":     true,
		"kind":   grouped[sc.ScOptionTypeKind],
		"fg":     grouped[sc.ScOptionTypeFG],
		"vessel": grouped[sc.ScOptionTypeVessel],
	}
}
