package forum

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
)

type SHForumAPI struct {
	deps *dep.Dep
}

func NewSHForumAPI(deps *dep.Dep) *SHForumAPI {
	return &SHForumAPI{deps: deps}
}

func (h *SHForumAPI) List(c *gin.Context) {
	rows, err := h.deps.ShForumModel.ListEnabled(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载论坛数据失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}
