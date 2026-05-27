package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

type mediaListQuery struct {
	Page            int64  `form:"p"`
	PageSize        int64  `form:"ps"`
	Sort            string `form:"sort"`
	Order           string `form:"order"`
	Keyword         string `form:"keyword"`
	SizeMin         string `form:"size_min"`
	SizeMax         string `form:"size_max"`
	HeightMin       string `form:"height_min"`
	HeightMax       string `form:"height_max"`
	DurationMin     string `form:"duration_min"`
	DurationMax     string `form:"duration_max"`
	BitRateMin      string `form:"bit_rate_min"`
	BitRateMax      string `form:"bit_rate_max"`
	FrameAverageMin string `form:"frame_average_min"`
	FrameAverageMax string `form:"frame_average_max"`
	SelfMake        string `form:"self_make"`
	HasMask         string `form:"has_mask"`
	ScTimesMin      string `form:"sc_times_min"`
	ScTimesMax      string `form:"sc_times_max"`
	LastScFrom      string `form:"last_sc_from"`
	LastScTo        string `form:"last_sc_to"`
	BirthFrom       string `form:"birth_from"`
	BirthTo         string `form:"birth_to"`
	ReleasingFrom   string `form:"releasing_from"`
	ReleasingTo     string `form:"releasing_to"`
}

func (h *MovieHTMLHandler) MediaListPage(c *gin.Context) {
	var q mediaListQuery
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
	q.Sort = normalizeFilmSortFieldForPage(q.Sort)
	q.Order = normalizeFilmSortOrderForPage(q.Order)

	filter, err := buildMediaListFilter(q)
	if err != nil {
		c.String(http.StatusBadRequest, "筛选参数错误: %v", err)
		return
	}

	items, total, err := h.mediaSvc.ListMediaPageWithFilter(
		c.Request.Context(),
		q.Page,
		q.PageSize,
		buildMediaOrderByForPage(q.Sort, q.Order),
		filter,
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "媒体列表加载失败: %v", err)
		return
	}

	pi := BuildPageInfo(c, total, q.Page, q.PageSize, pageWindow)
	sortQ := buildFilmSortQuery(c, q.Sort, q.Order)
	c.HTML(http.StatusOK, "page.w_media_list", gin.H{
		"Title":         "Medias",
		"Items":         items,
		"Total":         total,
		"PageInfo":      pi,
		"pageInfo":      pi,
		"FilmSortQuery": sortQ,
		"Query":         q,
	})
}

func (h *MovieHTMLHandler) ProbeMediaMeta(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "media id 无效"})
		return
	}

	result, err := h.mediaSvc.ProbeMediaMetaByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) || errors.Is(err, types.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "media 不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":     true,
		"result": result,
	})
}

func buildMediaListFilter(q mediaListQuery) (types.MediaListFilter, error) {
	filter := types.MediaListFilter{
		MovieNameKeyword: strings.TrimSpace(q.Keyword),
	}

	if v, ok, err := parseOptionalNonNegativeFloat64(q.SizeMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.SizeMin = gbToBytes(v)
		filter.HasSizeMin = true
	}
	if v, ok, err := parseOptionalNonNegativeFloat64(q.SizeMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.SizeMax = gbToBytes(v)
		filter.HasSizeMax = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.HeightMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.HeightMin = v
		filter.HasHeightMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.HeightMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.HeightMax = v
		filter.HasHeightMax = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.DurationMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.DurationMin = minutesToSeconds(v)
		filter.HasDurationMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.DurationMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.DurationMax = minutesToSeconds(v)
		filter.HasDurationMax = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.BitRateMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.BitRateMin = v
		filter.HasBitRateMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.BitRateMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.BitRateMax = v
		filter.HasBitRateMax = true
	}

	if v, ok, err := parseOptionalNonNegativeFloat64(q.FrameAverageMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.FrameAverageMin = v
		filter.HasFrameAverageMin = true
	}
	if v, ok, err := parseOptionalNonNegativeFloat64(q.FrameAverageMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.FrameAverageMax = v
		filter.HasFrameAverageMax = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.SelfMake); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.SelfMake = v
		filter.HasSelfMake = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.HasMask); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.HasMask = v
		filter.HasHasMask = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.ScTimesMin); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.ScTimesMin = v
		filter.HasScTimesMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.ScTimesMax); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.ScTimesMax = v
		filter.HasScTimesMax = true
	}

	if v, ok, err := parseOptionalDateUnixStart(q.LastScFrom); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.LastScFrom = v
		filter.HasLastScFrom = true
	}
	if v, ok, err := parseOptionalDateUnixEnd(q.LastScTo); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.LastScTo = v
		filter.HasLastScTo = true
	}

	if v, ok, err := parseOptionalDateUnixStart(q.BirthFrom); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.BirthTimeFrom = v
		filter.HasBirthTimeFrom = true
	}
	if v, ok, err := parseOptionalDateUnixEnd(q.BirthTo); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.BirthTimeTo = v
		filter.HasBirthTimeTo = true
	}

	if v, ok, err := parseOptionalDateUnixStart(q.ReleasingFrom); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.ReleasingDateFrom = v
		filter.HasReleasingDateFrom = true
	}
	if v, ok, err := parseOptionalDateUnixEnd(q.ReleasingTo); err != nil {
		return types.MediaListFilter{}, err
	} else if ok {
		filter.ReleasingDateTo = v
		filter.HasReleasingDateTo = true
	}

	return filter, nil
}

func buildMediaOrderByForPage(sortField, sortOrder string) string {
	field := normalizeFilmSortFieldForPage(sortField)
	order := "DESC"
	if normalizeFilmSortOrderForPage(sortOrder) == "asc" {
		order = "ASC"
	}

	switch field {
	case "movie_name":
		return "wm.movie_name " + order + ", wm.birth_time DESC, wm.id DESC"
	case "size":
		return "wm.size " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "height":
		return "wm.height " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "duration":
		return "wm.duration " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "bit_rate":
		return "wm.bit_rate " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "frame_average":
		return "wm.frame_average " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "self_make":
		return "wm.self_make " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "has_mask":
		return "wm.has_mask " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "sc_times":
		return "COALESCE(gss.sc_times, 0) " + order + ", COALESCE(gss.last_sc_time, 0) DESC, wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "last_sc_time":
		return "COALESCE(gss.last_sc_time, 0) " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	case "releasing_date":
		return "wm.releasing_date " + order + ", wm.birth_time DESC, wm.movie_name DESC, wm.id DESC"
	default:
		return "wm.birth_time " + order + ", wm.movie_name DESC, wm.id DESC"
	}
}
