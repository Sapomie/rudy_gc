package http

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/svc"
	htmlHandlers "rudy_gc/internal/transport/http/html"
)

func NewEngine(deps *svc.Deps) *gin.Engine {
	r := gin.Default()

	// 模板（不使用 ** 通配；仅按需解析三层）
	tpl := template.New("")
	tpl = template.Must(tpl.ParseGlob("ui/templates/layouts/*.html"))
	tpl = template.Must(tpl.ParseGlob("ui/templates/partials/*.html"))
	tpl = template.Must(tpl.ParseGlob("ui/templates/pages/*.html"))
	r.SetHTMLTemplate(tpl)

	// 静态资源
	r.Static("/static", "ui/static")

	// HTML 路由
	movieHTML := htmlHandlers.NewMovieHTMLHandler(deps)
	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/cards") })
	r.GET("/cards", movieHTML.ListMovieCardFull)

	return r
}
