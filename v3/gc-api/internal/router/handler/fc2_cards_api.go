package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/model/modelx/moviex"
)

type FC2CardsAPI struct {
	deps *dep.Dep
}

type fc2CardsQuery struct {
	Page     int64 `form:"p"`
	PageSize int64 `form:"ps"`
}

type fc2CardRow struct {
	*moviex.FC2Movie
	ScreenshotURL   string
	VideoURL        string
	PlayButtonClass string
	PlayButtonText  string
	ShowPlayButton  bool
	DownloadTime    int64
	ScTimes         int64
	ComeTimes       int64
	LastScTime      int64
}

func NewFC2CardsAPI(deps *dep.Dep) *FC2CardsAPI {
	return &FC2CardsAPI{deps: deps}
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

	movieIDs := make([]int64, 0, len(movieRows))
	for _, row := range movieRows {
		if row != nil && row.ID > 0 {
			movieIDs = append(movieIDs, row.ID)
		}
	}
	mediaRows, err := h.deps.FC2MediaModel.ListByMovieIDs(c.Request.Context(), movieIDs)
	if err != nil {
		jsonServerError(c, "FC2卡片媒体加载失败", err)
		return
	}
	statRows, err := h.deps.FC2ScStatModel.ListByMovieIDs(c.Request.Context(), movieIDs)
	if err != nil {
		jsonServerError(c, "FC2卡片SC统计加载失败", err)
		return
	}
	mediaIDs := make([]int64, 0, len(mediaRows))
	for _, row := range mediaRows {
		if row != nil && row.ID > 0 {
			mediaIDs = append(mediaIDs, row.ID)
		}
	}
	imageRows, err := h.deps.FC2MediaImageModel.ListScreenshotsByMediaIDs(c.Request.Context(), mediaIDs)
	if err != nil {
		jsonServerError(c, "FC2卡片截图加载失败", err)
		return
	}

	rows := buildFC2CardRows(movieRows, mediaRows, statRows, imageRows)
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

func buildFC2CardRows(movieRows []*moviex.FC2Movie, mediaRows []*moviex.FC2Media, statRows []*moviex.FC2ScStat, imageRows []*moviex.FC2MediaImage) []*fc2CardRow {
	mediaByMovieID := make(map[int64]*moviex.FC2Media, len(mediaRows))
	for _, row := range mediaRows {
		if row != nil && row.FC2MovieID > 0 {
			mediaByMovieID[row.FC2MovieID] = row
		}
	}
	statByMovieID := make(map[int64]*moviex.FC2ScStat, len(statRows))
	for _, row := range statRows {
		if row != nil && row.FC2MovieID > 0 {
			statByMovieID[row.FC2MovieID] = row
		}
	}
	screenshotByID := make(map[int64]*moviex.FC2MediaImage, len(imageRows))
	for _, row := range imageRows {
		if row == nil || row.FC2MediaID <= 0 || strings.TrimSpace(row.ImageType) != "screenshot" {
			continue
		}
		screenshotByID[row.ID] = row
	}

	items := make([]*fc2CardRow, 0, len(movieRows))
	for _, movieRow := range movieRows {
		if movieRow == nil {
			continue
		}
		item := &fc2CardRow{FC2Movie: movieRow}
		if mediaRow := mediaByMovieID[movieRow.ID]; mediaRow != nil {
			mediaView := buildFC2MediaView(mediaRow)
			item.VideoURL = mediaView.VideoURL
			item.PlayButtonClass = mediaView.PlayButtonClass
			item.PlayButtonText = mediaView.PlayButtonText
			item.ShowPlayButton = mediaView.ShowPlayButton
			item.DownloadTime = mediaRow.BirthTime
			if imageRow := screenshotByID[movieRow.CoverImageID]; imageRow != nil && imageRow.FC2MediaID == mediaRow.ID {
				item.ScreenshotURL = filepath.ToSlash(strings.TrimSpace(imageRow.FilePath))
			}
		}
		if statRow := statByMovieID[movieRow.ID]; statRow != nil {
			item.ScTimes = statRow.ScTimes
			item.ComeTimes = statRow.ComeTimes
			item.LastScTime = statRow.LastScTime
		}
		items = append(items, item)
	}
	return items
}
