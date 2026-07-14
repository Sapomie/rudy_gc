package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/dep"
	mediaservice "rudy_gc/internal/service/media"
)

type FC2CardsAPI struct {
	deps     *dep.Dep
	mediaSvc *mediaservice.Service
}

type fc2CardsQuery struct {
	Page     int64 `form:"p"`
	PageSize int64 `form:"ps"`
}

func NewFC2CardsAPI(deps *dep.Dep) *FC2CardsAPI {
	return &FC2CardsAPI{deps: deps, mediaSvc: mediaservice.NewService(deps)}
}

func (h *FC2CardsAPI) List(c *gin.Context) {
	h.list(c, "/fclistcard", "FC2 Cards", false, false)
}

func (h *FC2CardsAPI) ListOwned(c *gin.Context) {
	h.list(c, "/fclistcardowned", "FC2 Cards MediaOwned", true, false)
}

func (h *FC2CardsAPI) ListRandom(c *gin.Context) {
	h.list(c, "/fclistcardrandom", "FC2 Cards Random", false, true)
}

func (h *FC2CardsAPI) list(c *gin.Context, frontendPath, title string, ownedOnly, randomOrder bool) {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = frontendPath
	defer func() {
		c.Request.URL.Path = originalPath
	}()

	var query fc2CardsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		jsonBadRequest(c, "FC2卡片列表参数无效", err)
		return
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = defaultPageSize
	}

	total, err := h.deps.FC2MovieModel.CountCards(c.Request.Context(), ownedOnly)
	if err != nil {
		jsonServerError(c, "FC2卡片数量加载失败", err)
		return
	}
	movieRows, err := h.deps.FC2MovieModel.ListCards(c.Request.Context(), (query.Page-1)*query.PageSize, query.PageSize, ownedOnly, randomOrder)
	if err != nil {
		jsonServerError(c, "FC2卡片列表加载失败", err)
		return
	}

	rows, err := h.mediaSvc.BuildFC2CardViews(c.Request.Context(), movieRows)
	if err != nil {
		jsonServerError(c, "FC2卡片数据加载失败", err)
		return
	}
	pageInfo := BuildPageInfo(c, total, query.Page, query.PageSize, pageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":       title,
		"ViewType":    "fc2-cards",
		"movies":      rows,
		"total":       total,
		"PageInfo":    pageInfo,
		"pageInfo":    pageInfo,
		"RandomOrder": randomOrder,
	})
}
