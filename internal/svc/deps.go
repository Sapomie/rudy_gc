// internal/svc/deps.go

package svc

import (
	"rudy_gc/internal/contracts"
	"rudy_gc/pkg/mylog"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/data/modelx/spiderx"

	"rudy_gc/internal/config"
	"rudy_gc/internal/service/spider/fetcher"

	"github.com/sirupsen/logrus"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Deps struct {
	BestTrigger chan contracts.TriggerMsg
	FilmTrigger chan contracts.FilmTriggerMsg
	ScTrigger   chan contracts.ScTriggerMsg
	// ...
	Config  config.Config
	SqlConn sqlx.SqlConn
	Cache   cache.CacheConf
	Log     *logrus.Logger

	// moviex models（新 service 直连用）
	MovieModel      moviex.AMovieModel
	MinfoModel      moviex.BmMinfoModel
	MurlModel       moviex.BmMurlModel
	ItemModel       moviex.EItemModel
	CastModel       moviex.AmCastModel
	GenreModel      moviex.AmGenreModel
	DirectorModel   moviex.AmDirectorModel
	LabelModel      moviex.AmLabelModel
	MakerModel      moviex.AmMakerModel
	PrefixModel     moviex.AmPrefixModel
	MovieCastModel  moviex.AmrMovieCastModel
	MovieGenreModel moviex.AmrMovieGenreModel
	FilmModel       moviex.VFilmModel
	DirectoryModel  moviex.VDirectoryModel
	GListModel      moviex.GListModel
	ScModel         moviex.GScModel
	RankModel       moviex.CRankModel
	CafoModel       moviex.CCafoModel
	RecordModel     moviex.ERecordModel
	SeedModel       spiderx.DSeedModel
	InventoryModel  spiderx.DInventoryModel
	DetailModel     spiderx.DDetailModel
	BestinvModel    spiderx.DBestinvModel

	MovieTypeCache MovieTypeCache

	Fetcher    *fetcher.Fetcher
	DetailJobs chan string
}

func NewDeps(cfg config.Config) (*Deps, error) {
	conn := sqlx.NewMysql(cfg.DataSource)
	c := cfg.Cache

	var (
		movieModel  = moviex.NewAMovieModel(conn, c)
		minfoModel  = moviex.NewBmMinfoModel(conn, c)
		murlModel   = moviex.NewBmMurlModel(conn, c)
		itemModel   = moviex.NewEItemModel(conn, c)
		rankModel   = moviex.NewCRankModel(conn, c)
		cafoModel   = moviex.NewCCafoModel(conn, c)
		filmModel   = moviex.NewVFilmModel(conn, c)
		glistModel  = moviex.NewGListModel(conn, c)
		scModel     = moviex.NewGScModel(conn, c)
		recordModel = moviex.NewERecordModel(conn, c)
		seedModel   = spiderx.NewDSeedModel(conn)
		invModel    = spiderx.NewDInventoryModel(conn)
		detailModel = spiderx.NewDDetailModel(conn)
		bestModel   = spiderx.NewDBestinvModel(conn)

		// 新增的 8 个 model
		labelModel    = moviex.NewAmLabelModel(conn, c)
		makerModel    = moviex.NewAmMakerModel(conn, c)
		directorModel = moviex.NewAmDirectorModel(conn, c)
		prefixModel   = moviex.NewAmPrefixModel(conn, c)

		castModel       = moviex.NewAmCastModel(conn, c)
		genreModel      = moviex.NewAmGenreModel(conn, c)
		movieCastModel  = moviex.NewAmrMovieCastModel(conn, c)
		movieGenreModel = moviex.NewAmrMovieGenreModel(conn, c)

		// ★ 目录 model
		vdirModel = moviex.NewVDirectoryModel(conn, c)
	)

	bizRedis, err := redis.NewRedis(redis.RedisConf{
		Host: cfg.BizRedis.Host,
		Pass: cfg.BizRedis.Pass,
		Type: cfg.BizRedis.Type,
	})
	if err != nil {
		return nil, err
	}
	movieTypeCache := newMovieTypeBizCache(bizRedis, cfg.MovieTypeCache.Prefix, cfg.MovieTypeCache.Version, cfg.MovieTypeCache.TTL)

	// ========== fetcher ==========
	f := fetcher.NewFetcher(fetcher.Config{
		UserAgent: cfg.Fetcher.UserAgent,
		Cookie:    cfg.Fetcher.Cookie,
		Proxy:     cfg.Fetcher.Proxy,
		Timeout:   time.Duration(cfg.Fetcher.Timeout) * time.Second,
	})

	return &Deps{
		Config: cfg,
		Log:    mylog.NewLogrusLogger(cfg.LogursLevel),

		SqlConn: conn,
		Cache:   c,

		MovieModel:      movieModel,
		MinfoModel:      minfoModel,
		MurlModel:       murlModel,
		ItemModel:       itemModel,
		CastModel:       castModel,
		GenreModel:      genreModel,
		DirectorModel:   directorModel,
		LabelModel:      labelModel,
		MakerModel:      makerModel,
		PrefixModel:     prefixModel,
		MovieCastModel:  movieCastModel,
		MovieGenreModel: movieGenreModel,
		FilmModel:       filmModel,
		DirectoryModel:  vdirModel,
		GListModel:      glistModel,
		ScModel:         scModel,
		RankModel:       rankModel,
		CafoModel:       cafoModel,
		RecordModel:     recordModel,
		SeedModel:       seedModel,
		InventoryModel:  invModel,
		DetailModel:     detailModel,
		BestinvModel:    bestModel,
		MovieTypeCache:  movieTypeCache,

		DetailJobs:  make(chan string, 200),
		BestTrigger: make(chan contracts.TriggerMsg, 8),
		FilmTrigger: make(chan contracts.FilmTriggerMsg, 8),
		ScTrigger:   make(chan contracts.ScTriggerMsg, 16),

		Fetcher: f,
	}, nil
}
