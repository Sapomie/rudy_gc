package http

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/svc"
	api2 "rudy_gc/internal/transport/http/api"
	htmlHandlers "rudy_gc/internal/transport/http/html"
)

// NewEngine 启动 HTTP 路由
func NewEngine(deps *svc.Deps) *gin.Engine {
	r := gin.Default()

	// 模板
	tpl := template.New("")
	tpl = template.Must(tpl.ParseGlob("ui/templates/partials/*.html"))
	tpl = template.Must(tpl.ParseGlob("ui/templates/pages/*.html"))
	r.SetHTMLTemplate(tpl)

	// 静态资源
	r.Static("/static", "ui/static")
	r.Static("/Volumes/Expansion", "/Volumes/Expansion")
	r.Static("/Volumes/Getea", "/Volumes/Getea")
	r.Static("/Volumes/T7/data", "/Volumes/T7/data")

	r.Static("/text", "z_text")

	// ====== 实例化服务与 handler ======
	movieHTML := htmlHandlers.NewMovieHTMLHandler(deps)
	trig := api2.NewTriggerAPI(deps) //

	// ====== HTML 页面路由 ======
	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/cards") })
	r.GET("/cards", movieHTML.ListMovieCardFull)
	r.GET("/cardstoday", movieHTML.ListMovieCardToday)
	r.GET("/cardshasrank", movieHTML.ListMovieCardHasRank)
	r.GET("/cardsowned", movieHTML.ListMovieCardOwned)
	r.GET("/cardsneeddownload", movieHTML.ListMovieCardNeedDownload)
	r.GET("/movie/:movie", movieHTML.MovieDetail)
	r.GET("/cardsrandom", movieHTML.ListMovieCardRandom)

	// trigger 页面
	r.GET("/triggers", trig.Page)

	// ====== API 路由 ======
	api := r.Group("/api")
	{
		movieDownload := api2.NewMovieAPI(deps)
		api.POST("/movie/:movie/downloadlater", movieDownload.AddToDownloadLater)
		api.DELETE("/movie/:movie/downloadlater", movieDownload.RemoveFromDownloadLater)
		api.POST("/movie/:movie/download-cover", movieDownload.DownloadCoverNow)

		api.POST("/open-finder", movieDownload.OpenFinderHandler([]string{
			"/Volumes/Getea",
			"/Volumes/Expansion",
		}))

		// === triggers ===
		api.POST("/triggers/daily-best", trig.DailyBest)
		api.POST("/triggers/seeds", trig.Seeds)
		api.POST("/triggers/seed-by-name", trig.SeedByName)

		// 新增：影片触发
		filmTrig := api2.NewFilmTriggerAPI(deps)
		api.POST("/triggers/film/rename", filmTrig.Rename)
		api.POST("/triggers/film/process", filmTrig.Process)

		// ✅ 新增 SC 触发
		scTrig := api2.NewScTriggerAPI(deps)
		api.POST("/triggers/sc/move", scTrig.Move) // 带 scName
		api.POST("/triggers/sc/add", scTrig.Add)   // 带 dir
	}

	return r
}
