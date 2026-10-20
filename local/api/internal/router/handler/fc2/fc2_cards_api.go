package fc2

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/router/handler/common"
	mediaservice "rudy_gc/local/api/internal/service/media"
)

type FC2CardsAPI struct {
	deps     *dep.Dep
	mediaSvc *mediaservice.Service
}

type fc2CardsQuery struct {
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
	Sort     string `form:"sort"`
	Order    string `form:"order"`
}

type fc2CardsSortQuery struct {
	ByPostTime           common.SortLink
	ByFC2Key             common.SortLink
	ByLastSelectedScTime common.SortLink
	ByBirthTime          common.SortLink
	ByScTimes            common.SortLink
}

func NewFC2CardsAPI(deps *dep.Dep) *FC2CardsAPI {
	return &FC2CardsAPI{deps: deps, mediaSvc: mediaservice.NewService(deps)}
}

func (h *FC2CardsAPI) List(c *gin.Context) {
	h.list(c, "/fclistcard", "FC2 Cards", false, false, "birth_time")
}

func (h *FC2CardsAPI) ListOwned(c *gin.Context) {
	h.list(c, "/fclistcardowned", "FC2 Cards MediaOwned", true, false, "post_time")
}

func (h *FC2CardsAPI) ListRandom(c *gin.Context) {
	h.list(c, "/fclistcardrandom", "FC2 Cards Random", false, true, "post_time")
}

func (h *FC2CardsAPI) list(c *gin.Context, frontendPath, title string, ownedOnly, randomOrder bool, defaultSort string) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = frontendPath
	defer func() {
		c.Request.URL.Path = originalPath
	}()

	var query fc2CardsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		common.JSONBadRequest(c, "FC2卡片列表参数无效", err)
		return
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = common.DefaultPageSize
	}
	query.Sort = normalizeFC2CardSort(query.Sort, defaultSort)
	query.Order = normalizeFC2Order(query.Order)

	total, err := h.deps.FC2MovieModel.CountCards(c.Request.Context(), ownedOnly)
	if err != nil {
		common.JSONServerError(c, "FC2卡片数量加载失败", err)
		return
	}
	movieRows, err := h.deps.FC2MovieModel.ListCards(
		c.Request.Context(),
		(query.Page-1)*query.PageSize,
		query.PageSize,
		ownedOnly,
		randomOrder,
		query.Sort,
		query.Order,
	)
	if err != nil {
		common.JSONServerError(c, "FC2卡片列表加载失败", err)
		return
	}

	rows, err := h.mediaSvc.BuildFC2CardViews(c.Request.Context(), movieRows)
	if err != nil {
		common.JSONServerError(c, "FC2卡片数据加载失败", err)
		return
	}
	pageInfo := common.BuildPageInfo(c, total, query.Page, query.PageSize, common.PageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":       title,
		"ViewType":    "fc2-cards",
		"movies":      rows,
		"total":       total,
		"PageInfo":    pageInfo,
		"pageInfo":    pageInfo,
		"RandomOrder": randomOrder,
		"Query":       query,
		"SortQuery":   buildFC2CardsSortQuery(c, query.Sort, query.Order),
	})
}

func normalizeFC2CardSort(value, fallback string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "post_time", "fc2_key", "last_selected_sc_time", "birth_time", "sc_times":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return fallback
	}
}

func buildFC2CardsSortQuery(c *gin.Context, currentField, currentOrder string) *fc2CardsSortQuery {
	makeLink := func(field string) common.SortLink {
		active := currentField == field
		nextOrder := "desc"
		if active && currentOrder == "desc" {
			nextOrder = "asc"
		}
		query := common.CloneValues(c)
		query.Set("sort", field)
		query.Set("order", nextOrder)
		query.Set("p", "1")
		href := c.Request.URL.Path
		if encoded := query.Encode(); encoded != "" {
			href += "?" + encoded
		}
		return common.SortLink{
			Href:   href,
			Active: active,
			Desc:   active && currentOrder == "desc",
		}
	}

	return &fc2CardsSortQuery{
		ByPostTime:           makeLink("post_time"),
		ByFC2Key:             makeLink("fc2_key"),
		ByLastSelectedScTime: makeLink("last_selected_sc_time"),
		ByBirthTime:          makeLink("birth_time"),
		ByScTimes:            makeLink("sc_times"),
	}
}
