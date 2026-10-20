package forum

import (
	"net/http"
	"net/url"
	"rudy_gc/local/api/internal/router/handler/common"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	movieservice "rudy_gc/local/api/internal/service/movie"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

type SHMagnetAPI struct {
	deps     *dep.Dep
	movieSvc *movieservice.Service
}

type shMagnetBatchAddAlbumRequest struct {
	AlbumName    string  `json:"album_name"`
	SourceRowIDs []int64 `json:"source_row_ids"`
}

type shMagnetListQuery struct {
	Page          int64  `form:"p"`
	PageSize      int64  `form:"ps"`
	Sort          string `form:"sort"`
	Order         string `form:"order"`
	Keyword       string `form:"keyword"`
	MovieJavID    string `form:"movie_jav_id"`
	InfoHash      string `form:"info_hash"`
	MovieType     string `form:"movie_type"`
	DownloadState string `form:"download_state"`
	SourceForumID string `form:"source_forum_id"`
	PostTimeFrom  string `form:"post_time_from"`
	PostTimeTo    string `form:"post_time_to"`
}

type shMagnetPageRow struct {
	*moviex.ShMagnet
	MovieRouteKey string
	MovieHref     string
	CanFavorite   bool
	IsInDownload  bool
	IsInPending   bool
}

func NewSHMagnetAPI(deps *dep.Dep) *SHMagnetAPI {
	return &SHMagnetAPI{deps: deps, movieSvc: movieservice.NewService(deps)}
}

func (h *SHMagnetAPI) BatchAddToAlbum(c *gin.Context) {
	var req shMagnetBatchAddAlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.JSONBadRequest(c, "批量加入相册参数无效", err)
		return
	}
	req.AlbumName = strings.TrimSpace(req.AlbumName)
	if req.AlbumName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请选择目标相册"})
		return
	}
	if len(req.SourceRowIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "请选择要加入相册的条目"})
		return
	}

	result, err := h.movieSvc.BatchAddSHMagnetsToAlbum(c.Request.Context(), req.AlbumName, req.SourceRowIDs)
	if err != nil {
		common.JSONServerError(c, "批量加入相册失败", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ok":             true,
		"album_name":     result.AlbumName,
		"added_count":    len(result.AddedIDs),
		"existing_count": len(result.ExistingIDs),
		"failed_count":   len(result.FailedItems),
		"added_ids":      result.AddedIDs,
		"existing_ids":   result.ExistingIDs,
		"failed_items":   result.FailedItems,
		"message":        "批量加入相册完成",
	})
}

func (h *SHMagnetAPI) List(c *gin.Context) {
	var query shMagnetListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		common.JSONBadRequest(c, "SH 磁力列表参数无效", err)
		return
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 50
	}
	query.Keyword = strings.TrimSpace(query.Keyword)
	query.MovieJavID = strings.TrimSpace(query.MovieJavID)
	query.InfoHash = strings.TrimSpace(query.InfoHash)
	query.SourceForumID = strings.TrimSpace(query.SourceForumID)
	query.DownloadState = normalizeSHMagnetDownloadState(query.DownloadState)
	query.Sort = normalizeSHMagnetSort(query.Sort)
	query.Order = normalizeSHMagnetOrder(query.Order)
	if movieTypes := moviedomain.NormalizeSehuatangThreadMovieTypes(query.MovieType); len(movieTypes) == 1 {
		query.MovieType = movieTypes[0]
	} else {
		query.MovieType = ""
	}

	filter := moviex.ShMagnetListFilter{
		Keyword:       query.Keyword,
		MovieJavID:    query.MovieJavID,
		InfoHash:      query.InfoHash,
		MovieType:     query.MovieType,
		DownloadState: query.DownloadState,
		SourceForumID: query.SourceForumID,
	}
	if ts, ok, err := common.ParseOptionalDateStart(query.PostTimeFrom); err != nil {
		common.JSONBadRequest(c, "发帖时间开始日期错误", err)
		return
	} else if ok {
		filter.PostTimeFrom = ts
		filter.HasPostTimeFrom = true
	}
	if ts, ok, err := common.ParseOptionalDateEnd(query.PostTimeTo); err != nil {
		common.JSONBadRequest(c, "发帖时间结束日期错误", err)
		return
	} else if ok {
		filter.PostTimeTo = ts
		filter.HasPostTimeTo = true
	}
	total, err := h.deps.ShMagnetModel.CountPage(c.Request.Context(), filter)
	if err != nil {
		common.JSONServerError(c, "SH 磁力数量加载失败", err)
		return
	}
	rows, err := h.deps.ShMagnetModel.ListPage(c.Request.Context(), (query.Page-1)*query.PageSize, query.PageSize, query.Sort, query.Order, filter)
	if err != nil {
		common.JSONServerError(c, "SH 磁力列表加载失败", err)
		return
	}
	forums, err := h.deps.ShForumModel.ListEnabled(c.Request.Context())
	if err != nil {
		common.JSONServerError(c, "SH 论坛数据加载失败", err)
		return
	}
	pageRows, albums, err := h.buildPageRows(c, rows)
	if err != nil {
		common.JSONServerError(c, "SH 磁力相册状态加载失败", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Title":     "色花堂磁力列表",
		"PageTitle": "色花堂磁力列表",
		"PageNote":  "展示由 SH 帖子详情解析并按 Info Hash 去重入库的磁力资源。",
		"Rows":      pageRows,
		"Forums":    forums,
		"Albums":    albums,
		"Total":     total,
		"Query":     query,
		"PageInfo":  common.BuildPageInfo(c, total, query.Page, query.PageSize, common.PageWindow),
	})
}

func (h *SHMagnetAPI) buildPageRows(c *gin.Context, rows []*moviex.ShMagnet) ([]*shMagnetPageRow, []string, error) {
	albums, err := h.deps.AlbumModel.ListAll(c.Request.Context())
	if err != nil {
		return nil, nil, err
	}
	albumIDs := make(map[string]int64, len(albums))
	selectable := make([]string, 0, len(albums))
	for _, album := range albums {
		if album == nil {
			continue
		}
		name := strings.TrimSpace(album.Name)
		if name == "" {
			continue
		}
		albumIDs[name] = album.Id
		if name != "下载中" && name != "待下载" {
			selectable = append(selectable, name)
		}
	}
	sort.Strings(selectable)

	rowIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row != nil && row.Id > 0 {
			rowIDs = append(rowIDs, row.Id)
		}
	}
	statusByAlbum := make(map[string]map[int64]struct{}, 2)
	for _, albumName := range []string{"下载中", "待下载"} {
		albumID := albumIDs[albumName]
		items, err := h.deps.AlbumItemModel.ListByAlbumIdSourceTypeSourceRowIDs(c.Request.Context(), albumID, "sh_magnet", rowIDs)
		if err != nil {
			return nil, nil, err
		}
		ids := make(map[int64]struct{}, len(items))
		for _, item := range items {
			if item != nil {
				ids[item.SourceRowId] = struct{}{}
			}
		}
		statusByAlbum[albumName] = ids
	}

	movieRefs := make([]movieservice.MovieDetailLinkRef, len(rows))
	for index, row := range rows {
		if row == nil || row.MovieType == moviedomain.SehuatangThreadMovieTypeFC2PPV {
			continue
		}
		movieRefs[index] = movieservice.MovieDetailLinkRef{
			MovieJavID: row.MovieJavId,
			MovieName:  row.MovieName,
		}
	}
	movieHrefs, err := h.movieSvc.ResolveMovieDetailHrefs(c.Request.Context(), movieRefs)
	if err != nil {
		return nil, nil, err
	}

	out := make([]*shMagnetPageRow, 0, len(rows))
	for index, row := range rows {
		if row == nil {
			continue
		}
		movieHref := movieHrefs[index]
		routeKey := strings.TrimSpace(row.MovieJavId)
		if row.MovieType == moviedomain.SehuatangThreadMovieTypeFC2PPV {
			routeKey = strings.TrimSpace(row.FC2Key)
			if routeKey != "" {
				movieHref = "/fcdetail/" + url.PathEscape(routeKey)
			}
		}
		if routeKey == "" {
			routeKey = strings.TrimSpace(row.MovieName)
		}
		if routeKey == "" {
			routeKey = strings.TrimSpace(row.InfoHash)
		}
		_, inDownload := statusByAlbum["下载中"][row.Id]
		_, inPending := statusByAlbum["待下载"][row.Id]
		out = append(out, &shMagnetPageRow{
			ShMagnet:      row,
			MovieRouteKey: routeKey,
			MovieHref:     movieHref,
			CanFavorite:   routeKey != "",
			IsInDownload:  inDownload,
			IsInPending:   inPending,
		})
	}
	return out, selectable, nil
}

func normalizeSHMagnetSort(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_name", "updated_on":
		return strings.TrimSpace(raw)
	default:
		return "post_time"
	}
}

func normalizeSHMagnetOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}

func normalizeSHMagnetDownloadState(raw string) string {
	switch strings.TrimSpace(raw) {
	case "0", "1", "2":
		return strings.TrimSpace(raw)
	default:
		return ""
	}
}
