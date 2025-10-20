package html

import (
	"net/http"
	"rudy_gc/internal/domain/sc"

	"github.com/gin-gonic/gin"
)

func (h *MovieHTMLHandler) ListMovieCardRandom(c *gin.Context) {

	h.scSvc.PickMovie()

	c.HTML(http.StatusOK, "page.list_movie_card", gin.H{
		"Title":       title,
		"movies":      resp.List,
		"Total":       resp.Total,
		"PageInfo":    pi,
		"pageInfo":    pi,
		"ownedQuery":  ownedQ,
		"sortQuery":   sortQ,
		"CurrentSort": curOD,
		"total":       resp.Total,
		"fieldName":   fieldName,
	})
}
