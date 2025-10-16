package html

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

	md, err := h.svc.GetMovieDetailByName(c.Request.Context(), movieName)
	if err != nil {
		c.String(http.StatusInternalServerError, "查询失败: %v", err)
		return
	}
	if md == nil || md.MovieType == nil {
		c.String(http.StatusNotFound, "未找到影片: %s", movieName)
		return
	}

	// 旧模板中用了 .magUpdateDate，这里用 MovieType.UpdateDate 填充
	data := gin.H{
		"Title":         md.MovieType.Title,
		"movieType":     md.MovieType,
		"film":          md.FilmInfo,
		"rankInfos":     md.RankInfos,
		"sc":            md.SC,
		"magUpdateDate": md.MovieType.UpdateDate,
	}

	c.HTML(http.StatusOK, "page.movie_detail", data)
}
