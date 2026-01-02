package html

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *MovieHTMLHandler) ScPickCopyPage(c *gin.Context) {
	c.HTML(http.StatusOK, "page.sc_pick_copy", gin.H{
		"Title": "SC Pick Copy",
	})
}
