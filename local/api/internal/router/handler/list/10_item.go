package list

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/router/handler/common"
	"rudy_gc/local/api/internal/types"
)

type itemListQuery struct {
	Page                    int64  `form:"p"`
	PageSize                int64  `form:"ps"`
	Sort                    string `form:"sort"`
	Order                   string `form:"order"`
	JavID                   string `form:"jav_id"`
	Name                    string `form:"name"`
	HasDetail               string `form:"has_detail"`
	HasDownloadCover        string `form:"has_download_cover"`
	HasChinese              string `form:"has_chinese"`
	DetailNeedScan          string `form:"detail_need_scan"`
	DetailBirthTimeFrom     string `form:"detail_birth_time_from"`
	DetailBirthTimeTo       string `form:"detail_birth_time_to"`
	LastQueryDetailTimeFrom string `form:"last_query_detail_time_from"`
	LastQueryDetailTimeTo   string `form:"last_query_detail_time_to"`
}

func (h *ListPagesAPI) EItemList(c *gin.Context) {
	defer common.WithFrontendPath(c, "/eitemlist")()

	var q itemListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		common.JSONBadRequest(c, "参数解析错误", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 50
	}
	q.Sort = normalizeItemSortField(q.Sort)
	q.Order = normalizeItemSortOrder(q.Order)

	filter, err := buildItemListFilter(q)
	if err != nil {
		common.JSONBadRequest(c, "筛选参数错误", err)
		return
	}

	total, err := h.deps.ItemModel.CountAll(c.Request.Context(), filter)
	if err != nil {
		common.JSONServerError(c, "e_item 列表加载失败", err)
		return
	}

	offset := (q.Page - 1) * q.PageSize
	items, err := h.deps.ItemModel.ListPage(c.Request.Context(), offset, q.PageSize, buildItemOrderByForPage(q.Sort, q.Order), filter)
	if err != nil {
		common.JSONServerError(c, "e_item 列表加载失败", err)
		return
	}
	rows := h.buildItemListRows(items)

	pi := common.BuildPageInfo(c, total, q.Page, q.PageSize, common.PageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":         "E Items",
		"ViewType":      "e-items",
		"Items":         rows,
		"Total":         total,
		"PageInfo":      pi,
		"pageInfo":      pi,
		"Query":         q,
		"ItemSortQuery": buildItemSortQuery(c, q.Sort, q.Order),
	})
}

func (h *ListPagesAPI) buildItemListRows(items []*types.Item) []*types.ItemListRow {
	if len(items) == 0 {
		return []*types.ItemListRow{}
	}

	rows := make([]*types.ItemListRow, 0, len(items))
	for _, item := range items {
		row := buildItemListRow(item, h.deps.Config.Fetcher.JavAddress)
		if row != nil {
			rows = append(rows, row)
		}
	}
	return rows
}

func buildItemListRow(item *types.Item, javAddress string) *types.ItemListRow {
	if item == nil {
		return nil
	}

	hasDetailValue, hasDetailText := itemDetailStatusMeta(item.HasDetail)
	hasDownloadCoverValue, hasDownloadCoverText := itemCoverStatusMeta(item.HasDownloadCover)
	hasChineseValue, hasChineseText := itemChineseStatusMeta(item.HasChinese)
	detailNeedScanValue, detailNeedScanText := itemNeedScanStatusMeta(item.DetailNeedScan)

	return &types.ItemListRow{
		Id:                    item.Id,
		JavId:                 item.JavId,
		JavUrl:                buildItemJavURL(javAddress, item.JavId),
		Name:                  item.Name,
		HasDetail:             item.HasDetail,
		HasDetailValue:        hasDetailValue,
		HasDetailText:         hasDetailText,
		HasDownloadCover:      item.HasDownloadCover,
		HasDownloadCoverValue: hasDownloadCoverValue,
		HasDownloadCoverText:  hasDownloadCoverText,
		HasChinese:            item.HasChinese,
		HasChineseValue:       hasChineseValue,
		HasChineseText:        hasChineseText,
		DetailNeedScan:        item.DetailNeedScan,
		DetailNeedScanValue:   detailNeedScanValue,
		DetailNeedScanText:    detailNeedScanText,
		DetailBirthTime:       item.DetailBirthTime,
		LastQueryDetailTime:   item.LastQueryDetailTime,
	}
}

func buildItemJavURL(javAddress, javID string) string {
	if strings.TrimSpace(javAddress) == "" || strings.TrimSpace(javID) == "" {
		return ""
	}
	return "https://" + javAddress + "/cn/?v=" + javID
}

func itemDetailStatusMeta(status int64) (string, string) {
	switch status {
	case consts.ItemDetailNone:
		return "none", "无详情"
	case consts.ItemDetailOK:
		return "ok", "已有详情"
	default:
		return "", "-"
	}
}

func itemCoverStatusMeta(status int64) (string, string) {
	switch status {
	case consts.ItemCoverNone:
		return "none", "无本地封面"
	case consts.ItemCoverOK:
		return "ok", "已有本地封面"
	case consts.ItemCoverWrong:
		return "wrong", "错误封面链接"
	default:
		return "", "-"
	}
}

func itemChineseStatusMeta(status int64) (string, string) {
	switch status {
	case consts.ItemChineseNone:
		return "none", "无中文字幕"
	case consts.ItemChineseOK:
		return "ok", "有中文字幕"
	case consts.ItemChineseError:
		return "error", "翻译错误"
	case consts.ItemChineseSensitive:
		return "sensitive", "含敏感词"
	default:
		return "", "-"
	}
}

func itemNeedScanStatusMeta(status int64) (string, string) {
	switch status {
	case consts.ItemDetailStatusNeedScan:
		return "need", "需要扫描"
	case consts.ItemDetailStatusNoNeedScan:
		return "no_need", "无需扫描"
	case consts.ItemDetailStatusWrongContent:
		return "wrong", "内容错误"
	default:
		return "", "-"
	}
}

func buildItemListFilter(q itemListQuery) (types.ItemListFilter, error) {
	filter := types.ItemListFilter{
		JavID: strings.TrimSpace(q.JavID),
		Name:  strings.TrimSpace(q.Name),
	}

	if v, ok, err := parseItemDetailFilter(q.HasDetail); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.HasDetail = v
		filter.HasDetailSet = true
	}
	if v, ok, err := parseItemCoverFilter(q.HasDownloadCover); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.HasDownloadCover = v
		filter.HasDownloadCoverSet = true
	}
	if v, ok, err := parseItemChineseFilter(q.HasChinese); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.HasChinese = v
		filter.HasChineseSet = true
	}
	if v, ok, err := parseItemNeedScanFilter(q.DetailNeedScan); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.DetailNeedScan = v
		filter.DetailNeedScanSet = true
	}

	if ts, ok, err := common.ParseOptionalDateStart(q.DetailBirthTimeFrom); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.DetailBirthTimeFrom = ts
		filter.HasDetailBirthTimeFrom = true
	}
	if ts, ok, err := common.ParseOptionalDateEnd(q.DetailBirthTimeTo); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.DetailBirthTimeTo = ts
		filter.HasDetailBirthTimeTo = true
	}
	if ts, ok, err := common.ParseOptionalDateStart(q.LastQueryDetailTimeFrom); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.LastQueryDetailTimeFrom = ts
		filter.HasLastQueryDetailTimeFrom = true
	}
	if ts, ok, err := common.ParseOptionalDateEnd(q.LastQueryDetailTimeTo); err != nil {
		return types.ItemListFilter{}, err
	} else if ok {
		filter.LastQueryDetailTimeTo = ts
		filter.HasLastQueryDetailTimeTo = true
	}

	return filter, nil
}

func parseItemDetailFilter(raw string) (int64, bool, error) {
	switch strings.TrimSpace(raw) {
	case "":
		return 0, false, nil
	case "none":
		return consts.ItemDetailNone, true, nil
	case "ok":
		return consts.ItemDetailOK, true, nil
	default:
		return 0, false, fmt.Errorf("详情状态无效")
	}
}

func parseItemCoverFilter(raw string) (int64, bool, error) {
	switch strings.TrimSpace(raw) {
	case "":
		return 0, false, nil
	case "none":
		return consts.ItemCoverNone, true, nil
	case "ok":
		return consts.ItemCoverOK, true, nil
	case "wrong":
		return consts.ItemCoverWrong, true, nil
	default:
		return 0, false, fmt.Errorf("封面状态无效")
	}
}

func parseItemChineseFilter(raw string) (int64, bool, error) {
	switch strings.TrimSpace(raw) {
	case "":
		return 0, false, nil
	case "none":
		return consts.ItemChineseNone, true, nil
	case "ok":
		return consts.ItemChineseOK, true, nil
	case "error":
		return consts.ItemChineseError, true, nil
	case "sensitive":
		return consts.ItemChineseSensitive, true, nil
	default:
		return 0, false, fmt.Errorf("中文字幕状态无效")
	}
}

func parseItemNeedScanFilter(raw string) (int64, bool, error) {
	switch strings.TrimSpace(raw) {
	case "":
		return 0, false, nil
	case "need":
		return consts.ItemDetailStatusNeedScan, true, nil
	case "no_need":
		return consts.ItemDetailStatusNoNeedScan, true, nil
	case "wrong":
		return consts.ItemDetailStatusWrongContent, true, nil
	default:
		return 0, false, fmt.Errorf("详情扫描状态无效")
	}
}

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
		q := common.CloneValues(c)
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
