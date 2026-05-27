package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/router/handler"
)

func New(d *dep.Dep) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		cardsAPI := handler.NewMovieCardAPI(d)

		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
		api.GET("/cards", cardsAPI.ListMovieCardFull)
		api.GET("/cardstoday", cardsAPI.ListMovieCardToday)
		api.GET("/cardshasrank", cardsAPI.ListMovieCardHasRank)
		api.GET("/cardsmediamowned", cardsAPI.ListMovieCardMediaOwned)
		api.GET("/cardsneeddownload", cardsAPI.ListMovieCardNeedDownload)
		api.GET("/cardsrandom", cardsAPI.ListMovieCardFullRandom)
		api.GET("/moviecarddayrank", cardsAPI.ListMovieCardDayRank)
		api.GET("/moviecardperiodrank", cardsAPI.ListMovieCardPeriodRank)
	}

	return r
}
