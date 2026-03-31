package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *MovieHTMLHandler) ScPickSmartPage(c *gin.Context) {
	c.HTML(http.StatusOK, "page.sc_pick_smart", gin.H{
		"Title":           "SC Pick Smart",
		"PageTitle":       "SC Pick Smart",
		"PageSource":      "vfilm",
		"FixedHint":       "固定值：Owned=3，Page=1，PageSize=100000。Rank 配额口径为累计最小值：20 以内 / 100 以内 / 500 以内。",
		"BirthStartLabel": "下载日期起 (bs)",
		"BirthEndLabel":   "下载日期止 (be)",
		"BirthStartField": "FilmBirthTimeStart",
		"BirthEndField":   "FilmBirthTimeEnd",
		"Dir1Label":       "Dir1 (d1)",
		"Dir2Label":       "Dir2 (d2)",
		"Dir3Label":       "Dir3 (d3)",
		"Dir4Label":       "Dir4 (d4)",
		"Dir1Field":       "Dir1",
		"Dir2Field":       "Dir2",
		"Dir3Field":       "Dir3",
		"Dir4Field":       "Dir4",
	})
}

func (h *MovieHTMLHandler) ScPickSmartMediaPage(c *gin.Context) {
	c.HTML(http.StatusOK, "page.sc_pick_smart", gin.H{
		"Title":           "SC Pick Smart Media",
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

func (h *MovieHTMLHandler) ScTriggersPage(c *gin.Context) {
	c.HTML(http.StatusOK, "page.sc_triggers", gin.H{
		"Title":     "SC Triggers",
		"ScRootDir": h.deps.Config.Film.ScRootDir,
	})
}
