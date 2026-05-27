package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type itemSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type itemSortQuery struct {
	ByJavID               itemSortLink
	ByName                itemSortLink
	ByHasDetail           itemSortLink
	ByHasDownloadCover    itemSortLink
	ByHasChinese          itemSortLink
	ByDetailNeedScan      itemSortLink
	ByDetailBirthTime     itemSortLink
	ByLastQueryDetailTime itemSortLink
}

func buildItemSortQuery(c *gin.Context, currentField, currentOrder string) *itemSortQuery {
	makeHref := func(field string) itemSortLink {
		q := cloneValues(c)
		q.Set("p", "1")
		q.Set("sort", field)

		if currentField == field && currentOrder == "desc" {
			q.Set("order", "asc")
		} else {
			q.Set("order", "desc")
		}

		href := c.Request.URL.Path
		if enc := q.Encode(); enc != "" {
			href += "?" + enc
		}

		return itemSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &itemSortQuery{
		ByJavID:               makeHref("jav_id"),
		ByName:                makeHref("name"),
		ByHasDetail:           makeHref("has_detail"),
		ByHasDownloadCover:    makeHref("has_download_cover"),
		ByHasChinese:          makeHref("has_chinese"),
		ByDetailNeedScan:      makeHref("detail_need_scan"),
		ByDetailBirthTime:     makeHref("detail_birth_time"),
		ByLastQueryDetailTime: makeHref("last_query_detail_time"),
	}
}

func normalizeItemSortField(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "jav_id", "name", "has_detail", "has_download_cover", "has_chinese", "detail_need_scan", "detail_birth_time", "last_query_detail_time":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "last_query_detail_time"
	}
}

func normalizeItemSortOrder(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "asc") {
		return "asc"
	}
	return "desc"
}

func buildItemOrderByForPage(sortField, sortOrder string) string {
	field := normalizeItemSortField(sortField)
	order := "DESC"
	if normalizeItemSortOrder(sortOrder) == "asc" {
		order = "ASC"
	}

	column := "`last_query_detail_time`"
	switch field {
	case "jav_id":
		column = "`jav_id`"
	case "name":
		column = "`name`"
	case "has_detail":
		column = "`has_detail`"
	case "has_download_cover":
		column = "`has_download_cover`"
	case "has_chinese":
		column = "`has_chinese`"
	case "detail_need_scan":
		column = "`detail_need_scan`"
	case "detail_birth_time":
		column = "`detail_birth_time`"
	case "last_query_detail_time":
		column = "`last_query_detail_time`"
	}

	if column == "`jav_id`" || column == "`name`" {
		return column + " " + order + ", `id` DESC"
	}
	return column + " " + order + ", `id` DESC"
}
