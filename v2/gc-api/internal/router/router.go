package router

import (
	"rudy-gc-api/internal/dep"
	"rudy-gc-api/internal/router/handler/cards"
	"rudy-gc-api/internal/router/handler/movie"
	"rudy-gc-api/internal/router/handler/page"
	"rudy-gc-api/internal/router/handler/rank"
	"rudy-gc-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func New(d *dep.Dep) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	registerStaticRoutes(engine)
	registerRoutes(engine, d)

	return engine
}

func registerStaticRoutes(engine *gin.Engine) {
	engine.Static("/Volumes/Expansion", "/Volumes/Expansion")
	engine.Static("/Volumes/Getea", "/Volumes/Getea")
	engine.Static("/Volumes/movie-un", "/Volumes/movie-un")
	engine.Static("/Volumes/T7/data", "/Volumes/T7/data")
}

func registerRoutes(engine *gin.Engine, d *dep.Dep) {
	api := engine.Group("/api/gc/v2")
	{
		api.GET("/healthz", func(c *gin.Context) {
			response.JSON(c, 200, gin.H{"status": "ok"})
		})

		cards.Register(api, d)
		movie.Register(api, d)
		page.Register(api, d)
		rank.Register(api, d)
	}
}
