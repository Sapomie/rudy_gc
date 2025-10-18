package http

import (
	"html/template"
	"net/http"
	api2 "rudy_gc/internal/transport/http/api"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/svc"
	htmlHandlers "rudy_gc/internal/transport/http/html"
)

func NewEngine(deps *svc.Deps) *gin.Engine {
	r := gin.Default()

	// 模板（不使用 ** 通配；仅按需解析三层）
	tpl := template.New("")
	//tpl = template.Must(tpl.ParseGlob("ui/templates/layouts/*.html"))
	tpl = template.Must(tpl.ParseGlob("ui/templates/partials/*.html"))
	tpl = template.Must(tpl.ParseGlob("ui/templates/pages/*.html"))
	r.SetHTMLTemplate(tpl)

	// 静态资源
	r.Static("/static", "ui/static")

	// 单独挂载需要的外置硬盘目录
	r.Static("/Volumes/Expansion", "/Volumes/Expansion")
	r.Static("/Volumes/Getea", "/Volumes/Getea")
	r.Static("/Volumes/T7/data", "/Volumes/T7/data")

	// HTML 路由
	movieHTML := htmlHandlers.NewMovieHTMLHandler(deps)
	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/cards") })

	r.GET("/cards", movieHTML.ListMovieCardFull)
	r.GET("/cardstoday", movieHTML.ListMovieCardToday)
	r.GET("/cardshasrank", movieHTML.ListMovieCardHasRank)
	r.GET("/cardsowned", movieHTML.ListMovieCardOwned)
	r.GET("/cardsneeddownload", movieHTML.ListMovieCardNeedDownload)
	r.GET("/movie/:movie", movieHTML.MovieDetail)

	// API 路由：稍后下载
	api := r.Group("/api")
	movieDownload := api2.NewAPI(deps)
	api.POST("/movie/:movie/downloadlater", movieDownload.Add)
	api.DELETE("/movie/:movie/downloadlater", movieDownload.Remove)

	api.POST("/open-finder", movieDownload.OpenFinderHandler([]string{
		"/Volumes/Getea",
		"/Volumes/Expansion",
	}))

	return r
}
