package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *MovieHTMLHandler) ScPickSmartPage(c *gin.Context) {
	c.HTML(http.StatusOK, "page.sc_pick_smart", gin.H{
		"Title": "SC Pick Smart",
	})
}

func (h *MovieHTMLHandler) ScTriggersPage(c *gin.Context) {
	c.HTML(http.StatusOK, "page.sc_triggers", gin.H{
		"Title":     "SC Triggers",
		"ScRootDir": h.deps.Config.Film.ScRootDir,
	})
}
