package dailybestremotehttp

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"rudy_gc/remote/api/internal/modules/dailybest/application/collector"
	"rudy_gc/remote/api/internal/modules/dailybest/domain"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo      dailybest.Repository
	collector *dailybestcollector.Service
}

func New(repo dailybest.Repository, collector *dailybestcollector.Service, readToken, adminToken string) *gin.Engine {
	h := &Handler{repo: repo, collector: collector}
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.GET("/healthz", h.Health)
	router.GET("/readyz", h.Ready)

	v1 := router.Group("/api/v1")
	read := v1.Group("/dailybest", bearerAuth(readToken))
	read.GET("/snapshots", h.ListSnapshots)
	read.GET("/snapshots/:revision", h.GetSnapshot)
	read.GET("/snapshots/:revision/pages", h.GetSnapshotPages)
	read.GET("/dates/:date", h.GetSnapshotByDate)

	admin := v1.Group("/admin/dailybest", bearerAuth(adminToken))
	admin.POST("/fetch", h.StartFetch)
	admin.GET("/fetch-runs", h.ListFetchRuns)
	admin.GET("/fetch-runs/:id", h.GetFetchRun)
	return router
}

func bearerAuth(expected string) gin.HandlerFunc {
	expected = strings.TrimSpace(expected)
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		provided := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if header == provided || !secureEqual(provided, expected) {
			writeError(c, http.StatusUnauthorized, "unauthorized", "认证失败")
			c.Abort()
			return
		}
		c.Next()
	}
}

func secureEqual(left, right string) bool {
	if left == "" || right == "" || len(left) != len(right) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
