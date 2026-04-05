package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

type filmListQuery struct {
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

func (h *MovieHTMLHandler) FilmListPage(c *gin.Context) {
	var q filmListQuery
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

	filter, err := buildFilmListFilter(q)
	if err != nil {
		c.String(http.StatusBadRequest, "筛选参数错误: %v", err)
		return
	}

	items, total, err := h.vfilmSvc.ListFilmPageWithFilter(
		c.Request.Context(),
		q.Page,
		q.PageSize,
		buildFilmOrderByForPage(q.Sort, q.Order),
		filter,
	)
	if err != nil {
		c.String(http.StatusInternalServerError, "影片列表加载失败: %v", err)
		return
	}

	pi := BuildPageInfo(c, total, q.Page, q.PageSize, pageWindow)
	sortQ := buildFilmSortQuery(c, q.Sort, q.Order)
	c.HTML(http.StatusOK, "page.v_film_list", gin.H{
		"Title":         "Films",
		"Items":         items,
		"Total":         total,
		"PageInfo":      pi,
		"pageInfo":      pi,
		"FilmSortQuery": sortQ,
		"Query":         q,
	})
}

func (h *MovieHTMLHandler) ProbeFilmMeta(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "film id 无效"})
		return
	}

	result, err := h.vfilmSvc.ProbeFilmMetaByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) || errors.Is(err, types.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "error": "film 不存在"})
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

func buildFilmListFilter(q filmListQuery) (types.FilmListFilter, error) {
	filter := types.FilmListFilter{
		MovieNameKeyword: strings.TrimSpace(q.Keyword),
	}

	if v, ok, err := parseOptionalNonNegativeFloat64(q.SizeMin); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.SizeMin = gbToBytes(v)
		filter.HasSizeMin = true
	}
	if v, ok, err := parseOptionalNonNegativeFloat64(q.SizeMax); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.SizeMax = gbToBytes(v)
		filter.HasSizeMax = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.HeightMin); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.HeightMin = v
		filter.HasHeightMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.HeightMax); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.HeightMax = v
		filter.HasHeightMax = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.DurationMin); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.DurationMin = minutesToSeconds(v)
		filter.HasDurationMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.DurationMax); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.DurationMax = minutesToSeconds(v)
		filter.HasDurationMax = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.BitRateMin); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.BitRateMin = v
		filter.HasBitRateMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.BitRateMax); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.BitRateMax = v
		filter.HasBitRateMax = true
	}

	if v, ok, err := parseOptionalNonNegativeFloat64(q.FrameAverageMin); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.FrameAverageMin = v
		filter.HasFrameAverageMin = true
	}
	if v, ok, err := parseOptionalNonNegativeFloat64(q.FrameAverageMax); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.FrameAverageMax = v
		filter.HasFrameAverageMax = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.SelfMake); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.SelfMake = v
		filter.HasSelfMake = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.HasMask); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.HasMask = v
		filter.HasHasMask = true
	}

	if v, ok, err := parseOptionalNonNegativeInt64(q.ScTimesMin); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.ScTimesMin = v
		filter.HasScTimesMin = true
	}
	if v, ok, err := parseOptionalNonNegativeInt64(q.ScTimesMax); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.ScTimesMax = v
		filter.HasScTimesMax = true
	}

	if v, ok, err := parseOptionalDateUnixStart(q.LastScFrom); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.LastScFrom = v
		filter.HasLastScFrom = true
	}
	if v, ok, err := parseOptionalDateUnixEnd(q.LastScTo); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.LastScTo = v
		filter.HasLastScTo = true
	}

	if v, ok, err := parseOptionalDateUnixStart(q.BirthFrom); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.BirthTimeFrom = v
		filter.HasBirthTimeFrom = true
	}
	if v, ok, err := parseOptionalDateUnixEnd(q.BirthTo); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.BirthTimeTo = v
		filter.HasBirthTimeTo = true
	}

	if v, ok, err := parseOptionalDateUnixStart(q.ReleasingFrom); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.ReleasingDateFrom = v
		filter.HasReleasingDateFrom = true
	}
	if v, ok, err := parseOptionalDateUnixEnd(q.ReleasingTo); err != nil {
		return types.FilmListFilter{}, err
	} else if ok {
		filter.ReleasingDateTo = v
		filter.HasReleasingDateTo = true
	}

	return filter, nil
}

func buildFilmOrderByForPage(sortField, sortOrder string) string {
	field := normalizeFilmSortFieldForPage(sortField)
	order := "DESC"
	if normalizeFilmSortOrderForPage(sortOrder) == "asc" {
		order = "ASC"
	}

	switch field {
	case "movie_name":
		return "f.movie_name " + order + ", f.birth_time DESC, f.id DESC"
	case "size":
		return "f.size " + order + ", f.birth_time DESC, f.movie_name DESC, f.id DESC"
	case "height":
		return "f.height " + order + ", f.birth_time DESC, f.movie_name DESC, f.id DESC"
	case "duration":
		return "f.duration " + order + ", f.birth_time DESC, f.movie_name DESC, f.id DESC"
	case "bit_rate":
		return "f.bit_rate " + order + ", f.birth_time DESC, f.movie_name DESC, f.id DESC"
	case "frame_average":
		return "f.frame_average " + order + ", f.birth_time DESC, f.movie_name DESC, f.id DESC"
	case "self_make":
		return "f.self_make " + order + ", f.birth_time DESC, f.movie_name DESC, f.id DESC"
	case "has_mask":
		return "f.has_mask " + order + ", f.birth_time DESC, f.movie_name DESC, f.id DESC"
	case "sc_times":
		return "COALESCE(gss.sc_times, 0) " + order + ", COALESCE(gss.last_sc_time, 0) DESC, f.birth_time DESC, f.movie_name DESC, f.id DESC"
	case "last_sc_time":
		return "COALESCE(gss.last_sc_time, 0) " + order + ", f.birth_time DESC, f.movie_name DESC, f.id DESC"
	case "releasing_date":
		return "f.releasing_date " + order + ", f.birth_time DESC, f.movie_name DESC, f.id DESC"
	default:
		return "f.birth_time " + order + ", f.movie_name DESC, f.id DESC"
	}
}

func parseOptionalNonNegativeFloat64(raw string) (float64, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}

	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false, err
	}
	if v < 0 {
		return 0, false, fmt.Errorf("必须是非负数")
	}
	return v, true, nil
}

func parseOptionalDateUnixStart(raw string) (int64, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}

	t, err := time.ParseInLocation("2006-01-02", raw, time.Local)
	if err != nil {
		return 0, false, err
	}
	return t.Unix(), true, nil
}

func parseOptionalDateUnixEnd(raw string) (int64, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}

	t, err := time.ParseInLocation("2006-01-02", raw, time.Local)
	if err != nil {
		return 0, false, err
	}
	return t.Add(24*time.Hour - time.Second).Unix(), true, nil
}

func gbToBytes(v float64) int64 {
	return int64(v * 1024 * 1024 * 1024)
}

func minutesToSeconds(v int64) int64 {
	return v * 60
}
