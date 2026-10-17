package router

import (
	"github.com/gin-gonic/gin"

	dailybesthandler "rudy_gc/remote/api/internal/router/handler/dailybest"
	dailybestservice "rudy_gc/remote/api/internal/service/dailybest"
)

func New(dailyBestSvc *dailybestservice.Service) *gin.Engine {
	router := gin.Default()
	registerRoutes(router, dailyBestSvc)
	return router
}

func registerRoutes(router *gin.Engine, dailyBestSvc *dailybestservice.Service) {
	dailyBestAPI := dailybesthandler.NewDailyBestAPI(dailyBestSvc)
	router.GET("/healthz", dailyBestAPI.Health)
	router.GET("/readyz", dailyBestAPI.Ready)

	dailyBest := router.Group("/api/v1/dailybest")
	dailyBest.GET("/source-config", dailyBestAPI.GetSourceConfig)
	dailyBest.PUT("/source-config", dailyBestAPI.UpdateSourceConfig)
	dailyBest.GET("/rank-days", dailyBestAPI.ListRankDays)
	dailyBest.GET("/ranks", dailyBestAPI.ListRanks)
	dailyBest.GET("/rank-day-stats", dailyBestAPI.ListRankDayStats)
	dailyBest.POST("/fetch", dailyBestAPI.StartFetch)
	dailyBest.GET("/fetch-runs", dailyBestAPI.ListFetchRuns)
	dailyBest.GET("/fetch-runs/:id", dailyBestAPI.GetFetchRun)
	dailyBest.GET("/fetch-logs", dailyBestAPI.ListFetchLogs)
	dailyBest.GET("/target-probes", dailyBestAPI.ListTargetProbes)
	dailyBest.GET("/runtime-status", dailyBestAPI.RuntimeStatus)
	dailyBest.GET("/runtime-events", dailyBestAPI.ListRuntimeEvents)
}
