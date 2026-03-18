package html

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/types"
)

type castListQuery struct {
	Page         int64  `form:"p"`
	PageSize     int64  `form:"ps"`
	Sort         string `form:"sort"`
	Order        string `form:"order"`
	OwnedMin     string `form:"owned_min"`
	OwnedMax     string `form:"owned_max"`
	ScTimesMin   string `form:"sc_times_min"`
	ScTimesMax   string `form:"sc_times_max"`
	ComeTimesMin string `form:"come_times_min"`
	ComeTimesMax string `form:"come_times_max"`
	LastScFrom   string `form:"last_sc_from"`
	LastScTo     string `form:"last_sc_to"`
}

func (h *MovieHTMLHandler) CastListPage(c *gin.Context) {
	var q castListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.String(http.StatusBadRequest, "参数解析错误: %v", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 24
	}
	q.Sort = normalizeCastSortFieldForPage(q.Sort)
	q.Order = normalizeCastSortOrderForPage(q.Order)

	filter, err := buildCastListFilter(q)
	if err != nil {
		c.String(http.StatusBadRequest, "筛选参数错误: %v", err)
		return
	}

	items, total, err := h.deps.CastRepo.ListPage(c.Request.Context(), int(q.Page), int(q.PageSize), q.Sort, q.Order, filter)
	if err != nil {
		c.String(http.StatusInternalServerError, "演员列表加载失败: %v", err)
		return
	}

	pi := BuildPageInfo(c, total, q.Page, q.PageSize, pageWindow)
	sortQ := buildCastSortQuery(c, q.Sort, q.Order)

	c.HTML(http.StatusOK, "page.cast_list", gin.H{
		"Title":         "Casts",
		"Items":         items,
		"Total":         total,
		"PageInfo":      pi,
		"pageInfo":      pi,
		"CastSortQuery": sortQ,
		"Query":         q,
	})
}

func buildCastListFilter(q castListQuery) (types.CastListFilter, error) {
	filter := types.CastListFilter{}

	if v, ok, err := parseOptionalNonNegativeInt64(q.OwnedMin); err != nil {
		return types.CastListFilter{}, err
	} else if ok {
		filter.OwnedMin = v
		filter.HasOwnedMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.OwnedMax); err != nil {
		return types.CastListFilter{}, err
	} else if ok {
		filter.OwnedMax = v
		filter.HasOwnedMax = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.ScTimesMin); err != nil {
		return types.CastListFilter{}, err
	} else if ok {
		filter.ScTimesMin = v
		filter.HasScTimesMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.ScTimesMax); err != nil {
		return types.CastListFilter{}, err
	} else if ok {
		filter.ScTimesMax = v
		filter.HasScTimesMax = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.ComeTimesMin); err != nil {
		return types.CastListFilter{}, err
	} else if ok {
		filter.ComeTimesMin = v
		filter.HasComeTimesMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.ComeTimesMax); err != nil {
		return types.CastListFilter{}, err
	} else if ok {
		filter.ComeTimesMax = v
		filter.HasComeTimesMax = true
	}

	if strings.TrimSpace(q.LastScFrom) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(q.LastScFrom), time.Local)
		if err != nil {
			return types.CastListFilter{}, err
		}
		filter.LastScFrom = t.Unix()
		filter.HasLastScFrom = true
	}

	if strings.TrimSpace(q.LastScTo) != "" {
		t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(q.LastScTo), time.Local)
		if err != nil {
			return types.CastListFilter{}, err
		}
		filter.LastScTo = t.Add(24*time.Hour - time.Second).Unix()
		filter.HasLastScTo = true
	}

	return filter, nil
}

func parseOptionalNonNegativeInt64(raw string) (int64, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}

	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false, err
	}
	if v < 0 {
		return 0, false, fmt.Errorf("必须是非负整数")
	}
	return v, true, nil
}
