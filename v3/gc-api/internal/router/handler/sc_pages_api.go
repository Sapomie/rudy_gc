package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/media"
	"rudy_gc/internal/service/sc"
	"rudy_gc/internal/types"
)

type ScPagesAPI struct {
	deps     *dep.Dep
	scSvc    *sc.ScService
	mediaSvc *media.Service
}

func NewScPagesAPI(deps *dep.Dep) *ScPagesAPI {
	return &ScPagesAPI{
		deps:     deps,
		scSvc:    sc.NewService(deps),
		mediaSvc: media.NewService(deps),
	}
}

type scEventListQuery struct {
	Page      int64  `form:"p"`
	PageSize  int64  `form:"ps"`
	Sort      string `form:"sort"`
	Order     string `form:"order"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

func (h *ScPagesAPI) ListScEvents(c *gin.Context) {
	defer withFrontendPath(c, "/sc-events")()

	var q scEventListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		jsonBadRequest(c, "参数解析错误", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 200 {
		q.PageSize = 24
	}
	q.Sort = normalizeScEventSortField(q.Sort)
	q.Order = normalizeScEventSortOrder(q.Order)
	startTime, endTime, err := parseScEventDateRange(q.StartDate, q.EndDate)
	if err != nil {
		jsonBadRequest(c, "时间筛选参数错误", err)
		return
	}

	resp, err := h.scSvc.ListEventPage(c.Request.Context(), int(q.Page), int(q.PageSize), q.Sort, q.Order, startTime, endTime)
	if err != nil {
		jsonServerError(c, "加载 SC 事件失败", err)
		return
	}

	pi := BuildPageInfo(c, resp.Total, q.Page, q.PageSize, pageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":       "SC Events",
		"ViewType":    "sc-events",
		"Items":       resp.Items,
		"Total":       resp.Total,
		"PageInfo":    pi,
		"pageInfo":    pi,
		"ScSortQuery": buildScEventSortQuery(c, q.Sort, q.Order),
		"ScViewQuery": buildScEventViewQuery(c),
	})
}

func (h *ScPagesAPI) ListScEventCards(c *gin.Context) {
	defer withFrontendPath(c, "/sc-events-cards")()

	var q scEventListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		jsonBadRequest(c, "参数解析错误", err)
		return
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 18
	}
	q.Sort = normalizeScEventSortField(q.Sort)
	q.Order = normalizeScEventSortOrder(q.Order)
	startTime, endTime, err := parseScEventDateRange(q.StartDate, q.EndDate)
	if err != nil {
		jsonBadRequest(c, "时间筛选参数错误", err)
		return
	}

	resp, err := h.scSvc.ListEventCardPage(c.Request.Context(), int(q.Page), int(q.PageSize), q.Sort, q.Order, startTime, endTime)
	if err != nil {
		jsonServerError(c, "加载 SC 事件失败", err)
		return
	}

	pi := BuildPageInfo(c, resp.Total, q.Page, q.PageSize, pageWindow)
	c.JSON(http.StatusOK, gin.H{
		"Title":       "SC Event Cards",
		"ViewType":    "sc-event-cards",
		"Items":       resp.Items,
		"Total":       resp.Total,
		"PageInfo":    pi,
		"pageInfo":    pi,
		"ScSortQuery": buildScEventSortQuery(c, q.Sort, q.Order),
		"ScViewQuery": buildScEventViewQuery(c),
	})
}

func parseScEventDateRange(startDate, endDate string) (int64, int64, error) {
	const layout = "2006-01-02"
	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)

	var startTime int64
	if startDate != "" {
		value, err := time.ParseInLocation(layout, startDate, time.Local)
		if err != nil {
			return 0, 0, fmt.Errorf("起始日期格式错误: %w", err)
		}
		startTime = value.Unix()
	}

	var endTime int64
	if endDate != "" {
		value, err := time.ParseInLocation(layout, endDate, time.Local)
		if err != nil {
			return 0, 0, fmt.Errorf("结束日期格式错误: %w", err)
		}
		endTime = value.AddDate(0, 0, 1).Unix()
	}

	if startTime > 0 && endTime > 0 && startTime >= endTime {
		return 0, 0, errors.New("起始日期不能晚于结束日期")
	}
	return startTime, endTime, nil
}

func (h *ScPagesAPI) ScEventDetail(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数: name"})
		return
	}
	defer withFrontendPath(c, "/sc-events/"+url.PathEscape(name))()

	movieFilter := normalizeScEventDetailMovieFilter(c.Query("mf"))
	detail, err := h.scSvc.GetEventDetail(c.Request.Context(), name, movieFilter)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到 SC 事件: " + name})
			return
		}
		jsonServerError(c, "加载 SC 事件失败", err)
		return
	}
	if detail == nil || detail.Event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到 SC 事件: " + name})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":                     detail.Event.Name,
		"ViewType":                  "sc-event-detail",
		"Event":                     detail.Event,
		"Items":                     detail.Items,
		"FailedItems":               detail.FailedItems,
		"ComeCount":                 detail.ComeCount,
		"EditImageName":             detail.EditImageName,
		"EditMovieCount":            detail.EditMovieCount,
		"EditScMovieCount":          detail.EditScMovieCount,
		"EditComeMovieJavId":        detail.EditComeMovieJavId,
		"EditComeMovieOptions":      detail.EditComeMovieOptions,
		"EditCurrentMovieCastNames": detail.EditCurrentMovieCastNames,
		"MovieFilter":               movieFilter,
		"MovieFilterAllHref":        buildScEventDetailMovieFilterHref(c, "all"),
		"MovieFilterScHref":         buildScEventDetailMovieFilterHref(c, "sc"),
		"MovieFilterNoScHref":       buildScEventDetailMovieFilterHref(c, "nosc"),
	})
}

func normalizeScEventDetailMovieFilter(v string) string {
	switch strings.TrimSpace(v) {
	case "all":
		return "all"
	case "nosc":
		return "nosc"
	default:
		return "sc"
	}
}

func buildScEventDetailMovieFilterHref(c *gin.Context, movieFilter string) string {
	path := c.Request.URL.Path
	query := c.Request.URL.Query()
	query.Set("mf", normalizeScEventDetailMovieFilter(movieFilter))
	encoded := query.Encode()
	if encoded == "" {
		return path
	}
	return path + "?" + encoded
}

type scEventSortLink struct {
	Href   string
	Active bool
	Desc   bool
}

type scEventSortQuery struct {
	ByTime          scEventSortLink
	ByTotalMovieCnt scEventSortLink
	ByMovieCnt      scEventSortLink
	ByComeMovie     scEventSortLink
	ByCooldown      scEventSortLink
	ByDuration      scEventSortLink
	ByKind          scEventSortLink
	ByMovieCast     scEventSortLink
	ByVessel        scEventSortLink
	ByFg            scEventSortLink
}

func buildScEventSortQuery(c *gin.Context, currentField, currentOrder string) *scEventSortQuery {
	makeHref := func(field string) scEventSortLink {
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

		return scEventSortLink{
			Href:   href,
			Active: currentField == field,
			Desc:   currentField == field && currentOrder == "desc",
		}
	}

	return &scEventSortQuery{
		ByTime:          makeHref("sc_time"),
		ByTotalMovieCnt: makeHref("total_movie_count"),
		ByMovieCnt:      makeHref("sc_movie_count"),
		ByComeMovie:     makeHref("come_movie_name"),
		ByCooldown:      makeHref("cooldown"),
		ByDuration:      makeHref("duration"),
		ByKind:          makeHref("kind"),
		ByMovieCast:     makeHref("movie_cast"),
		ByVessel:        makeHref("vessel"),
		ByFg:            makeHref("fg"),
	}
}

func normalizeScEventSortField(v string) string {
	switch v {
	case "total_movie_count", "sc_movie_count", "movie_number", "come_movie_name", "cooldown", "duration", "kind", "movie_cast", "vessel", "fg", "sc_time":
		return v
	default:
		return "sc_time"
	}
}

func normalizeScEventSortOrder(v string) string {
	if v == "asc" {
		return "asc"
	}
	return "desc"
}

type scEventViewQuery struct {
	ListHref string
	CardHref string
	IsList   bool
	IsCard   bool
}

func buildScEventViewQuery(c *gin.Context) *scEventViewQuery {
	q := c.Request.URL.Query()

	listHref := "/sc-events"
	cardHref := "/sc-events-cards"
	if enc := q.Encode(); enc != "" {
		listHref += "?" + enc
		cardHref += "?" + enc
	}

	return &scEventViewQuery{
		ListHref: listHref,
		CardHref: cardHref,
		IsList:   c.Request.URL.Path == "/sc-events",
		IsCard:   c.Request.URL.Path == "/sc-events-cards",
	}
}

func (h *ScPagesAPI) ScPickSmartMediaPage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Title":           "SC Pick Smart Media",
		"ViewType":        "sc-pick-smart-media",
		"PageTitle":       "SC Pick Smart Media",
		"PageSource":      "wmedia",
		"FixedHint":       "固定值：MediaOwned=3，Page=1，PageSize=100000。Rank 配额口径为累计最小值：20 以内 / 100 以内 / 500 以内。",
		"BirthStartLabel": "下载日期起 (mbs)",
		"BirthEndLabel":   "下载日期止 (mbe)",
		"BirthStartField": "MediaBirthTimeStart",
		"BirthEndField":   "MediaBirthTimeEnd",
		"Dir1Label":       "MediaDir1 (md1)",
		"Dir2Label":       "MediaDir2 (md2)",
		"Dir3Label":       "MediaDir3 (md3)",
		"Dir4Label":       "MediaDir4 (md4)",
		"Dir1Field":       "MediaDir1",
		"Dir2Field":       "MediaDir2",
		"Dir3Field":       "MediaDir3",
		"Dir4Field":       "MediaDir4",
	})
}

func (h *ScPagesAPI) ScTriggersPage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Title":     "SC Triggers",
		"ViewType":  "sc-triggers",
		"ScRootDir": h.deps.Config.Film.ScRootDir,
	})
}

func (h *ScPagesAPI) MediaPage(c *gin.Context) {
	favoriteRows, _ := h.loadMediaPageFavoriteAlbumRows(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{
		"Title":        "媒体入库",
		"ViewType":     "media-ingest",
		"PageTitle":    "媒体入库（两段执行）",
		"PageNote":     "第一段只做预处理校验与插入前预览；第二段仅插入通过项。",
		"RootDirs":     h.deps.Config.Media.RootDirs,
		"FavoriteRows": favoriteRows,
	})
}

func (h *ScPagesAPI) MediaRescanPage(c *gin.Context) {
	rootOptions, err := h.mediaSvc.ListLibraryRescanRootOptions(c.Request.Context())
	if err != nil {
		jsonServerError(c, "加载媒体重扫页面失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Title":       "媒体重扫",
		"ViewType":    "media-rescan",
		"PageTitle":   "媒体重扫（位置与删除状态同步）",
		"PageNote":    "同一个 root 下可分别勾选 media 与 watched；若勾选二级目录，则只扫描该子树。未挂载或不存在的范围会直接跳过，并保持现有 w_media 不变。",
		"RootOptions": rootOptions,
	})
}

func (h *ScPagesAPI) MediaRollbackPage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Title":     "媒体回滚",
		"ViewType":  "media-rollback",
		"PageTitle": "媒体回滚（名称还原）",
		"PageNote":  "将 004_rollback 中的文件名还原后移回 001_ingest_new。",
		"RootDirs":  h.deps.Config.Media.RootDirs,
	})
}

func (h *ScPagesAPI) ScMediaMovePage(c *gin.Context) {
	mode := "single"
	if strings.TrimSpace(c.Query("mode")) == "all" {
		mode = "all"
	}
	scName := strings.TrimSpace(c.Query("sc_name"))
	scDetailHref := ""
	if mode == "single" && scName != "" {
		scDetailHref = "/sc-events/" + url.PathEscape(scName)
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":        "SC Media 移动",
		"ViewType":     "sc-media-move",
		"PageTitle":    "SC Media 移动（两段执行）",
		"PageNote":     "可按单个 SC 或全部符合条件执行；第一段只做预处理预览，第二段仅移动预处理通过的 WMedia 到 watched 目录。",
		"Mode":         mode,
		"RootDirs":     h.deps.Config.Media.RootDirs,
		"ScName":       scName,
		"ScDetailHref": scDetailHref,
	})
}

type mediaPageFavoriteAlbumItemRow struct {
	MovieName   string
	MovieJavID  string
	SourceType  string
	SourceLabel string
	Size        string
}

const (
	mediaPageFavoriteAlbumName  = "下载中"
	mediaPageFavoriteCardMaxRow = 50
)

func (h *ScPagesAPI) loadMediaPageFavoriteAlbumRows(ctx context.Context) ([]*mediaPageFavoriteAlbumItemRow, error) {
	album, err := h.deps.AlbumModel.FindOneByName(ctx, mediaPageFavoriteAlbumName)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return []*mediaPageFavoriteAlbumItemRow{}, nil
		}
		return nil, err
	}

	rows, err := h.deps.AlbumItemModel.ListPageRows(ctx, album.Id, 0, mediaPageFavoriteCardMaxRow, "`created_on` DESC, `id` DESC", moviex.AlbumItemPageFilter{})
	if err != nil {
		return nil, err
	}

	out := make([]*mediaPageFavoriteAlbumItemRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		movieName := strings.TrimSpace(row.MovieName)
		out = append(out, &mediaPageFavoriteAlbumItemRow{
			MovieName:   movieName,
			MovieJavID:  strings.TrimSpace(row.MovieJavId),
			SourceType:  strings.TrimSpace(row.SourceType),
			SourceLabel: mediaPageFavoriteSourceTypeLabel(row.SourceType),
			Size:        strings.TrimSpace(row.Size),
		})
	}
	return out, nil
}

func mediaPageFavoriteSourceTypeLabel(sourceType string) string {
	switch strings.TrimSpace(sourceType) {
	case "javbus_magnet":
		return "JavBus"
	case "sukebei_torrent":
		return "Sukebei"
	case "sehuatang_magnet":
		return "Sehuatang"
	default:
		return "未知"
	}
}
