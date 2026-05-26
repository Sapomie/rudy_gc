package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *MovieHTMLHandler) MovieDetail(c *gin.Context) {
	movieName := c.Param("movie")
	if movieName == "" {
		c.String(http.StatusBadRequest, "缺少参数: movie")
		return
	}

	md, err := h.movieSvc.GetMovieDetailByName(c.Request.Context(), movieName)
	if err != nil {
		c.String(http.StatusInternalServerError, "查询失败: %v", err)
		return
	}
	if md == nil || md.MovieType == nil {
		c.String(http.StatusNotFound, "未找到影片: %s", movieName)
		return
	}

	data := gin.H{
		"Title":         md.MovieType.Title,
		"movieType":     md.MovieType,
		"media":         md.MediaInfo,
		"rankInfos":     md.RankInfos,
		"sc":            md.SC,
		"javbusFetch":   md.JavbusFetch,
		"javbusMagnets": md.JavbusMagnets,
		"sukebeiFetch":  md.SukebeiFetch,
		"sukebeiRows":   md.SukebeiTorrents,
		"sehuatangRows": md.SehuatangMagnets,
		"magUpdateDate": md.MovieType.UpdateDate,
	}
	if movieAlbums, err := h.movieSvc.ListMovieAlbumsByMovieJavID(c.Request.Context(), md.MovieType.JavId); err == nil {
		data["MovieAlbums"] = movieAlbums
	}
	if markedDelete, err := h.movieSvc.IsMovieMarkedDelete(c.Request.Context(), md.MovieType.JavId); err == nil {
		data["MovieDeleteAlbumMarked"] = markedDelete
	}

	c.HTML(http.StatusOK, "page.movie_detail", data)
}
