package http

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/svc"
	api2 "rudy_gc/internal/transport/http/api"
	htmlHandlers "rudy_gc/internal/transport/http/html"
)

// NewEngine 启动 HTTP 路由
func NewEngine(deps *svc.Deps) *gin.Engine {
	r := gin.Default()

	// 模板函数
	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"humanSize": func(bytes int64) string {
			if bytes < 1024 {
				return fmt.Sprintf("%d B", bytes)
			} else if bytes < 1024*1024 {
				return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
			} else if bytes < 1024*1024*1024 {
				return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
			}
			return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
		},
		"humanTime": func(ts int64) string {
			if ts == 0 {
				return "-"
			}
			return time.Unix(ts, 0).Format("2006-01-02")
		},
		// ========= Record 页面需要 =========
		"formatUnix": func(sec int64) string {
			if sec <= 0 {
				return "-"
			}
			return time.Unix(sec, 0).Format("2006-01-02 15:04:05")
		},
		"minus": func(a, b int64) int64 {
			return a - b
		},
		"humanDuration": func(sec int64) string {
			if sec <= 0 {
				return "-"
			}
			d := time.Duration(sec) * time.Second
			h := int(d.Hours())
			m := int(d.Minutes()) % 60
			s := int(d.Seconds()) % 60
			switch {
			case h > 0:
				return fmt.Sprintf("%dh %dm %ds", h, m, s)
			case m > 0:
				return fmt.Sprintf("%dm %ds", m, s)
			default:
				return fmt.Sprintf("%ds", s)
			}
		},
	}

	// 模板
	tpl := template.New("").Funcs(funcMap)
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
	dirHTML := htmlHandlers.NewDirectoryHTMLHandler(deps)
	trig := api2.NewTriggerAPI(deps) //
	aggHTML := htmlHandlers.NewMovieAggHTMLHandler(deps)

	// ====== HTML 页面路由 ======
	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/cards") })
	r.GET("/cards", movieHTML.ListMovieCardFull)
	r.GET("/cardstoday", movieHTML.ListMovieCardToday)
	r.GET("/cardshasrank", movieHTML.ListMovieCardHasRank)
	r.GET("/cardsowned", movieHTML.ListMovieCardOwned)
	r.GET("/cardsneeddownload", movieHTML.ListMovieCardNeedDownload)
	r.GET("/movie/:movie", movieHTML.MovieDetail)

	r.GET("/cardsrandom", movieHTML.ListMovieCardFullRandom)
	r.GET("/cardsrandompick", movieHTML.ListMovieCardRandomPick)
	r.GET("/records", movieHTML.ListRecordsPage)
	r.GET("/dir/:id", dirHTML.DirDetail)
	// trigger 页面
	r.GET("/triggers", trig.Page)

	r.GET("/movie-agg-owned/release", aggHTML.MovieAggOwnedReleaseYears)
	r.GET("/movie-agg-owned/release/:year", aggHTML.MovieAggOwnedReleaseMonths)
	r.GET("/movie-agg-owned/release/:year/q/:q", aggHTML.MovieAggOwnedReleaseQuarter)
	r.GET("/movie-agg-owned/release/:year/:month", aggHTML.MovieAggOwnedReleaseMonth)

	// 下载日（birth）
	r.GET("/movie-agg-owned/birth", aggHTML.MovieAggOwnedBirthYears)
	r.GET("/movie-agg-owned/birth/:year", aggHTML.MovieAggOwnedBirthMonths)
	r.GET("/movie-agg-owned/birth/:year/q/:q", aggHTML.MovieAggOwnedBirthQuarter)
	r.GET("/movie-agg-owned/birth/:year/:month", aggHTML.MovieAggOwnedBirthMonth)

	r.GET("/movie-agg-all/release", aggHTML.MovieAggAllReleaseYears)
	r.GET("/movie-agg-all/release/:year", aggHTML.MovieAggAllReleaseMonths)
	r.GET("/movie-agg-all/release/:year/q/:q", aggHTML.MovieAggAllReleaseQuarter)
	r.GET("/movie-agg-all/release/:year/:month", aggHTML.MovieAggAllReleaseMonth)

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
		trig := api2.NewTriggerAPI(deps)
		api.POST("/triggers/daily-best", trig.DailyBest)
		api.POST("/triggers/daily-best-sync", trig.DailyBestSync)
		api.POST("/triggers/seeds", trig.Seeds)
		api.POST("/triggers/seed-by-name", trig.SeedByName)

		// === 影片触发 ===
		filmTrig := api2.NewFilmTriggerAPI(deps)
		api.POST("/triggers/film/rename", filmTrig.Rename)
		api.POST("/triggers/film/process", filmTrig.Process)

		// === SC 触发 ===
		scTrig := api2.NewScTriggerAPI(deps)
		api.POST("/triggers/sc/move", scTrig.Move)
		api.POST("/triggers/sc/add", scTrig.Add)
	}

	return r
}
