package http

import (
	"html/template"

	"rudy_gc/internal/svc"
	htmlHandlers "rudy_gc/internal/transport/http/html"

	"github.com/gin-gonic/gin"
)

func NewEngine(deps *svc.Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// 如模板里用到自定义函数可在此注册
	r.SetFuncMap(template.FuncMap{
		// 示例:
		// "formatDate": func(s string) string { return s },
	})

	// 模板与静态资源
	r.LoadHTMLGlob("ui/templates/**/*.html")
	r.Static("/static", "./ui/static")

	// ========== HTML 路由 ==========
	movieHTML := htmlHandlers.NewMovieHTMLHandler(deps)
	r.GET("/movies", movieHTML.ListMoviesPage)

	// ========== 现有 API 路由（你已有就保留原注册） ==========
	// api := r.Group("/api/v1")
	//  ... 省略 ...

	return r
}
