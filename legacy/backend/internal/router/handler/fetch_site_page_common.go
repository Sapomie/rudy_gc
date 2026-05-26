package handler

import "github.com/gin-gonic/gin"

func buildFetchSitePageInfo(c *gin.Context, total int64, page int64, pageSize int64) *PageInfo {
	return BuildPageInfo(c, total, page, pageSize, 2)
}
