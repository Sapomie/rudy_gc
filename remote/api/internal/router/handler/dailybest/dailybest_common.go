package dailybest

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	dailybestservice "rudy_gc/remote/api/internal/service/dailybest"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func parsePageQuery(c *gin.Context) (dailybesttypes.PageQuery, error) {
	pageRaw, hasPage := c.GetQuery("p")
	pageSizeRaw, hasPageSize := c.GetQuery("ps")
	if !hasPage && !hasPageSize {
		return dailybesttypes.PageQuery{}, nil
	}
	page := int64(1)
	pageSize := dailybesttypes.DefaultPageSize
	var err error
	if strings.TrimSpace(pageRaw) != "" {
		page, err = strconv.ParseInt(strings.TrimSpace(pageRaw), 10, 64)
		if err != nil {
			return dailybesttypes.PageQuery{}, errors.New("p 必须为整数")
		}
	}
	if strings.TrimSpace(pageSizeRaw) != "" {
		pageSize, err = strconv.ParseInt(strings.TrimSpace(pageSizeRaw), 10, 64)
		if err != nil {
			return dailybesttypes.PageQuery{}, errors.New("ps 必须为整数")
		}
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > dailybesttypes.MaxPageSize {
		pageSize = dailybesttypes.DefaultPageSize
	}
	return dailybesttypes.PageQuery{Page: page, PageSize: pageSize, Enabled: true}, nil
}

func buildPageInfo(total int64, page dailybesttypes.PageQuery) *contracts.DailyBestPageInfo {
	if !page.Enabled {
		return nil
	}
	pageTotal := (total + page.PageSize - 1) / page.PageSize
	if pageTotal <= 0 {
		pageTotal = 1
	}
	return &contracts.DailyBestPageInfo{
		Total: total, Page: page.Page, PageSize: page.PageSize, PageTotal: pageTotal,
	}
}

func parseNonNegativeInt64(raw, field string) (int64, error) {
	if strings.TrimSpace(raw) == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value < 0 {
		return 0, errors.New(field + " 必须为非负整数")
	}
	return value, nil
}

func parseDayNumber(dayNumberRaw, dateRaw string) (int64, error) {
	dayNumber, err := parseNonNegativeInt64(dayNumberRaw, "day_number")
	if err != nil {
		return 0, err
	}
	dateRaw = strings.TrimSpace(dateRaw)
	if dateRaw == "" {
		return dayNumber, nil
	}
	dateDayNumber, err := contracts.RankDayNumber(dateRaw)
	if err != nil || dateDayNumber <= 0 {
		return 0, errors.New("日期必须为 2023-01-01 及之后的 YYYY-MM-DD")
	}
	if dayNumber > 0 && dayNumber != dateDayNumber {
		return 0, errors.New("day_number 与日期不一致")
	}
	return dateDayNumber, nil
}

func parsePositiveInt64(raw, field string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value <= 0 {
		return 0, errors.New(field + " 必须为正整数")
	}
	return value, nil
}

func writeServiceError(c *gin.Context, resource string, err error) {
	if errors.Is(err, dailybestservice.ErrNotFound) {
		writeError(c, http.StatusNotFound, resource+"_not_found", "数据不存在")
		return
	}
	writeError(c, http.StatusInternalServerError, resource+"_query_failed", err.Error())
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, contracts.DailyBestAPIError{Code: code, Message: message})
}
