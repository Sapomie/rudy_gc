package router

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/router/handler"
)

func New(d *dep.Dep) *gin.Engine {
	r := gin.Default()

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
		"formatUnix": func(sec int64) string {
			if sec <= 0 {
				return "-"
			}
			return time.Unix(sec, 0).Format("2006-01-02 15:04:05")
		},
		"formatUnixMinute": func(sec int64) string {
			if sec <= 0 {
				return "-"
			}
			return time.Unix(sec, 0).Format("2006-01-02 15:04")
		},
		"formatUnixDate": func(sec int64) string {
			if sec <= 0 {
				return "-"
			}
			return time.Unix(sec, 0).Format("2006-01-02")
		},
		"formatUnixHm": func(sec int64) string {
			if sec <= 0 {
				return "-"
			}
			return time.Unix(sec, 0).Format("15:04")
		},
		"monthParityClass": func(sec int64) string {
			if sec <= 0 {
				return "sc-month-even"
			}
			if int(time.Unix(sec, 0).Month())%2 == 1 {
				return "sc-month-odd"
			}
			return "sc-month-even"
		},
		"minus": func(a, b int64) int64 { return a - b },
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
		"humanDays": func(sec int64) string {
			if sec <= 0 {
				return "-"
			}
			return fmt.Sprintf("%.2f 天", float64(sec)/86400.0)
		},
		"humanAgeFromBirth": func(birth int64) string {
			if birth <= 0 {
				return "-"
			}
			now := time.Now().In(time.Local)
			bd := time.Unix(birth, 0).In(time.Local)
			age := now.Year() - bd.Year()
			if now.Month() < bd.Month() || (now.Month() == bd.Month() && now.Day() < bd.Day()) {
				age--
			}
			if age < 0 {
				return "-"
			}
			return fmt.Sprintf("%d", age)
		},
	}

	tpl := template.New("").Funcs(funcMap)
	tpl = template.Must(tpl.ParseGlob("ui/templates/partials/*.html"))
	tpl = template.Must(tpl.ParseGlob("ui/templates/pages/*.html"))
	r.SetHTMLTemplate(tpl)

	r.Static("/static", "ui/static")
	r.Static("/Volumes/Expansion", "/Volumes/Expansion")
	r.Static("/Volumes/Getea", "/Volumes/Getea")
	r.Static("/Volumes/movie-un", "/Volumes/movie-un")
	r.Static("/Volumes/T7/data", "/Volumes/T7/data")
	r.Static("/text", "z_text")

	movieHTML := handler.NewMovieHTMLHandler(d)
	dirHTML := handler.NewDirectoryHTMLHandler(d)
	wdirHTML := handler.NewWDirectoryHTMLHandler(d)
	crawlerPages := handler.NewCrawlerPages(d)
	crawlerAPI := handler.NewCrawlerAPI(d)
	aggHTML := handler.NewMovieAggHTMLHandler(d)
	wMediaAggHTML := handler.NewWMediaAggHTMLHandler(d)

	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/cards") })
	r.GET("/cards", movieHTML.ListMovieCardFull)
	r.GET("/cardstoday", movieHTML.ListMovieCardToday)
	r.GET("/cardshasrank", movieHTML.ListMovieCardHasRank)
	r.GET("/cardsowned", movieHTML.ListMovieCardOwned)
	r.GET("/cardsmediamowned", movieHTML.ListMovieCardMediaOwned)
	r.GET("/cardsneeddownload", movieHTML.ListMovieCardNeedDownload)
	r.GET("/moviecarddayrank", movieHTML.ListMovieCardDayRank)
	r.GET("/moviecardperiodrank", movieHTML.ListMovieCardPeriodRank)
	r.GET("/movie/:movie", movieHTML.MovieDetail)
	r.GET("/cardsrandom", movieHTML.ListMovieCardFullRandom)
	r.GET("/records", movieHTML.ListRecordsPage)
	r.GET("/sc-events", movieHTML.ListScEventsPage)
	r.GET("/sc-events-cards", movieHTML.ListScEventsCardPage)
	r.GET("/sc-events/:name", movieHTML.ScEventDetailPage)
	r.GET("/casts", movieHTML.CastListPage)
	r.GET("/cast", movieHTML.CastDetailPage)
	r.GET("/films", movieHTML.FilmListPage)
	r.GET("/medias", movieHTML.MediaListPage)
	r.GET("/e-items", movieHTML.EItemListPage)
	r.GET("/album-items", movieHTML.AlbumItemsPage)
	r.GET("/fetch-site-javbus-list", crawlerPages.FetchSiteJavbusListPageMain)
	r.GET("/fetch-site-sukebei-list", crawlerPages.FetchSiteSukebeiListPageMain)
	r.GET("/fetch-site-sehuatang-list", crawlerPages.FetchSiteSehuatangListPageMain)
	r.GET("/dir", dirHTML.RootList)
	r.GET("/wdir", wdirHTML.RootList)
	r.GET("/triggers", crawlerPages.JobsPage)
	r.GET("/triggers/dailybest", crawlerPages.DailyBestPage)
	r.GET("/triggers/seeds", crawlerPages.SeedsPage)
	r.GET("/triggers/film", crawlerPages.FilmPage)
	r.GET("/triggers/film-move", crawlerPages.FilmMovePage)
	r.GET("/triggers/post-process", crawlerPages.PostProcessPage)
	r.GET("/triggers/media", crawlerPages.MediaPage)
	r.GET("/triggers/sc-media", crawlerPages.ScMediaMovePage)
	r.GET("/triggers/media-rollback", crawlerPages.MediaRollbackPage)
	r.GET("/triggers/media-rescan", crawlerPages.MediaRescanPage)
	r.GET("/triggers/fetch-site", crawlerPages.FetchSitePage)
	r.GET("/triggers/fetch-site-javbus-filtered", crawlerPages.FetchSiteJavbusFilteredPage)
	r.GET("/triggers/fetch-site-sukebei-filtered", crawlerPages.FetchSiteSukebeiFilteredPage)
	r.GET("/triggers/fetch-sehuatang", crawlerPages.FetchSehuatangPage)
	r.GET("/triggers/backfill", crawlerPages.BackfillPage)
	r.GET("/crawler/tasks", crawlerPages.TasksPage)
	r.GET("/crawler/detail-loop", crawlerPages.DetailLoopPage)
	r.GET("/triggers/sc", movieHTML.ScTriggersPage)
	r.GET("/sc-pick-smart", movieHTML.ScPickSmartPage)
	r.GET("/sc-pick-smart-media", movieHTML.ScPickSmartMediaPage)
	r.GET("/movie-agg-owned/birth", aggHTML.MovieAggOwnedBirthYears)
	r.GET("/movie-agg-owned/birth/:year", aggHTML.MovieAggOwnedBirthMonths)
	r.GET("/movie-agg-owned/birth/:year/q/:q", aggHTML.MovieAggOwnedBirthQuarter)
	r.GET("/movie-agg-owned/birth/:year/:month", aggHTML.MovieAggOwnedBirthMonth)
	r.GET("/w-media-agg/birth", wMediaAggHTML.BirthYears)
	r.GET("/w-media-agg/bucket-list", wMediaAggHTML.BucketList)
	r.GET("/movie-agg-all/release", aggHTML.MovieAggAllReleaseYears)
	r.GET("/movie-agg-all/release-bucket-list", aggHTML.MovieAggAllReleaseBucketList)

	api := r.Group("/api")
	{
		movieAPI := handler.NewMovieAPI(d)
		eItemAPI := handler.NewEItemAPI(d)
		personAPI := handler.NewPersonAPI(d)
		wKvAPI := handler.NewWKvAPI(d)
		api.POST("/movie/:movie/downloadlater", movieAPI.AddToDownloadLater)
		api.DELETE("/movie/:movie/downloadlater", movieAPI.RemoveFromDownloadLater)
		api.POST("/movie/:movie/download-cover", movieAPI.DownloadCoverNow)
		api.POST("/movie/:movie/move-wmedia-removed", movieAPI.MoveWMediaToRemoved)
		api.POST("/movie/:movie/add-cast", movieAPI.AddCast)
		api.POST("/movie/:movie/album-item", movieAPI.AddFetchResourceToAlbum)
		api.DELETE("/movie/:movie/album-item", movieAPI.RemoveFetchResourceFromAlbum)
		api.POST("/albums", movieAPI.CreateAlbum)
		api.POST("/albums/:albumID/items/remove", movieAPI.RemoveAlbumItem)
		api.POST("/albums/:albumID/items/batch-remove", movieAPI.BatchRemoveAlbumItems)
		api.POST("/albums/:albumID/items/batch-move", movieAPI.BatchMoveAlbumItems)
		api.POST("/films/:id/probe-meta", movieHTML.ProbeFilmMeta)
		api.POST("/medias/:id/probe-meta", movieHTML.ProbeMediaMeta)
		api.GET("/persons/:id/merge-candidates", personAPI.MergeCandidates)
		api.POST("/persons/merge-preview", personAPI.MergePreview)
		api.POST("/persons/merge", personAPI.Merge)
		api.POST("/e-items/status/batch", eItemAPI.UpdateStatusBatch)
		api.POST("/e-items/:id/status", eItemAPI.UpdateStatus)
		api.DELETE("/e-items/:id", eItemAPI.Delete)
		api.POST("/open-finder", movieAPI.OpenFinderHandler([]string{
			"/Volumes/Getea",
			"/Volumes/Expansion",
			"/Volumes/movie-un",
			"/Volumes/T7/data",
		}))
		api.POST("/w-kv/date", wKvAPI.UpsertDate)

		triggerAPI := handler.NewTriggerAPI(d)
		api.POST("/triggers/daily-best", triggerAPI.DailyBest)
		api.POST("/triggers/daily-best-sync", triggerAPI.DailyBestSync)
		api.POST("/triggers/rebuild-cast-rank", triggerAPI.RebuildCastRank)
		api.POST("/triggers/cast/rebuild-rank", triggerAPI.RebuildActorRank)
		api.POST("/triggers/seeds", triggerAPI.Seeds)
		api.POST("/triggers/seed-by-name", triggerAPI.SeedByName)
		api.POST("/triggers/refresh-oldest-detail", triggerAPI.RefreshOldestDetail)

		filmTriggerAPI := handler.NewFilmTriggerAPI(d)
		api.POST("/triggers/film/rename", filmTriggerAPI.Rename)
		api.POST("/triggers/film/process", filmTriggerAPI.Process)

		filmMoveAPI := handler.NewFilmMoveAPI(d)
		api.GET("/triggers/film-move/preview", filmMoveAPI.Preview)
		api.POST("/triggers/film-move/commit", filmMoveAPI.Commit)

		mediaTriggerAPI := handler.NewMediaTriggerAPI(d)
		api.GET("/triggers/media/plan", mediaTriggerAPI.Plan)
		api.POST("/triggers/media/precheck", mediaTriggerAPI.Precheck)
		api.POST("/triggers/media/commit", mediaTriggerAPI.Commit)
		api.POST("/triggers/media/return", mediaTriggerAPI.Return)
		api.POST("/triggers/media/rollback", mediaTriggerAPI.Rollback)
		api.POST("/triggers/media/rescan", mediaTriggerAPI.Rescan)
		api.POST("/triggers/media/agg-backfill", mediaTriggerAPI.BackfillWMediaAgg)
		api.POST("/triggers/movie/release-agg-backfill", mediaTriggerAPI.BackfillMovieReleaseAgg)

		scMediaTriggerAPI := handler.NewScMediaTriggerAPI(d)
		api.GET("/triggers/sc-media/plan", scMediaTriggerAPI.Plan)
		api.POST("/triggers/sc-media/precheck", scMediaTriggerAPI.Precheck)
		api.POST("/triggers/sc-media/commit", scMediaTriggerAPI.Commit)
		api.POST("/triggers/sc-media/return", scMediaTriggerAPI.Return)

		scTriggerAPI := handler.NewScTriggerAPI(d)
		api.POST("/triggers/sc/move", scTriggerAPI.Move)
		api.POST("/triggers/sc/add-preview", scTriggerAPI.AddPreview)
		api.POST("/triggers/sc/add", scTriggerAPI.Add)
		api.POST("/triggers/sc/rebuild-stats", scTriggerAPI.RebuildStats)
		api.POST("/triggers/sc/pick-smart-copy", scTriggerAPI.PickSmartCopy)
		api.POST("/triggers/sc/pick-smart-only", scTriggerAPI.PickSmartOnly)
		api.GET("/triggers/sc/copy-status", scTriggerAPI.CopyStatus)
		api.POST("/triggers/sc/copy-stop", scTriggerAPI.CopyStop)
		api.POST("/crawler/jobs/start", crawlerAPI.Start)
		api.POST("/crawler/jobs/fetch-site/preview", crawlerAPI.PreviewFetchSite)
		api.GET("/crawler/jobs", crawlerAPI.ListJobs)
		api.GET("/crawler/jobs/stream", crawlerAPI.StreamAll)
		api.GET("/crawler/detail-loop", crawlerAPI.GetDetailLoop)
		api.POST("/crawler/detail-loop/start", crawlerAPI.StartDetailLoop)
		api.POST("/crawler/detail-loop/stop", crawlerAPI.StopDetailLoop)
		api.GET("/crawler/detail-loop/stream", crawlerAPI.StreamDetailLoop)
		api.GET("/crawler/jobs/:jobID", crawlerAPI.GetJob)
		api.GET("/crawler/jobs/:jobID/events", crawlerAPI.GetJobEvents)
		api.GET("/crawler/jobs/:jobID/stream", crawlerAPI.Stream)
		api.POST("/crawler/jobs/:jobID/pause", crawlerAPI.Pause)
		api.POST("/crawler/jobs/:jobID/resume", crawlerAPI.Resume)
		api.POST("/crawler/jobs/:jobID/stop", crawlerAPI.Stop)
	}

	return r
}
