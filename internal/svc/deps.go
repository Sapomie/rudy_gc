// internal/svc/deps.go

package svc

import (
	"rudy_gc/internal/contracts"
	"rudy_gc/internal/domain/spider/fetcher"
	"rudy_gc/internal/infra/bizcache"
	film_infra "rudy_gc/internal/infra/film_repo"
	"rudy_gc/internal/repo/film_repo"
	"rudy_gc/pkg/mylog"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/data/modelx/spiderx"

	"rudy_gc/internal/config"
	"rudy_gc/internal/infra/movie_infra"
	"rudy_gc/internal/infra/spider_infra"
	"rudy_gc/internal/repo/movie_repo"
	"rudy_gc/internal/repo/spider_repo"

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

	// spider repos (无缓存)
	SeedRepo      spider_repo.SeedRepo
	InventoryRepo spider_repo.InventoryRepo
	ItemRepo      spider_repo.ItemRepo
	DetailRepo    spider_repo.DetailRepo
	BestinvRepo   spider_repo.BestinvRepo

	// movie repos (带缓存)
	MovieRepo      movie_repo.MovieRepo
	MovieListRepo  movie_repo.MovieListRepo
	CastRepo       movie_repo.CastRepo
	GenreRepo      movie_repo.GenreRepo
	DirectorRepo   movie_repo.DirectorRepo
	LabelRepo      movie_repo.LabelRepo
	MakerRepo      movie_repo.MakerRepo
	PrefixRepo     movie_repo.PrefixRepo
	MovieCastRepo  movie_repo.MovieCastRepo
	MovieGenreRepo movie_repo.MovieGenreRepo
	MinfoRepo      movie_repo.MinfoRepo
	MurlRepo       movie_repo.MurlRepo
	RankRepo       movie_repo.RankRepo
	CafoRepo       movie_repo.CafoRepo
	DirectoryRepo  film_repo.DirectoryRepo
	FilmRepo       film_repo.FilmRepo
	GListRepo      film_repo.GListRepo
	ScRepo         film_repo.ScRepo

	// ✅ 新增：记录表仓库（ERecord）
	RecordRepo movie_repo.RecordRepo

	MovieTypeCache movie_repo.MovieTypeCache

	Fetcher    *fetcher.Fetcher
	DetailJobs chan string
}

func NewDeps(cfg config.Config) (*Deps, error) {
	conn := sqlx.NewMysql(cfg.DataSource)
	c := cfg.Cache

	var (
		movieModel = moviex.NewAMovieModel(conn, c)
		minfoModel = moviex.NewBmMinfoModel(conn, c)
		filmModel  = moviex.NewVFilmModel(conn, c)

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

	// ========== spider (无缓存) ==========
	var (
		seedRepo      = spider_infra.NewSeedRepoSqlx(spiderx.NewDSeedModel(conn))
		inventoryRepo = spider_infra.NewInventoryRepoSqlx(spiderx.NewDInventoryModel(conn))
		bestRepo      = spider_infra.NewBestinvRepoSqlx(spiderx.NewDBestinvModel(conn))
		detailRepo    = spider_infra.NewDetailRepoSqlx(spiderx.NewDDetailModel(conn))
	)

	// ========== movie (有缓存) ==========
	var (
		movieRepo      = movie_infra.NewMovieRepoSqlx(movieModel)
		castRepo       = movie_infra.NewCastRepoSqlx(castModel)
		genreRepo      = movie_infra.NewGenreRepoSqlx(genreModel)
		directorRepo   = movie_infra.NewDirectorRepoSqlx(directorModel)
		labelRepo      = movie_infra.NewLabelRepoSqlx(labelModel)
		makerRepo      = movie_infra.NewMakerRepoSqlx(makerModel)
		prefixRepo     = movie_infra.NewPrefixRepoSqlx(prefixModel)
		movieCastRepo  = movie_infra.NewMovieCastRepoSqlx(movieCastModel)
		movieGenreRepo = movie_infra.NewMovieGenreRepoSqlx(movieGenreModel)
		minfoRepo      = movie_infra.NewMinfoRepoSqlx(minfoModel)
		murlRepo       = movie_infra.NewMurlRepoSqlx(moviex.NewBmMurlModel(conn, c))
		itemRepo       = movie_infra.NewItemRepoSqlx(moviex.NewEItemModel(conn, c))
		rankRepo       = movie_infra.NewRankRepoSqlx(moviex.NewCRankModel(conn, c))
		cafoRepo       = movie_infra.NewCafoRepoSqlx(moviex.NewCCafoModel(conn, c))
		directoryRepo  = film_infra.NewDirectoryRepoSqlx(vdirModel)
		filmRepo       = film_infra.NewFilmRepoSqlx(filmModel)

		movieListRepo = movie_infra.NewMovieListRepoSqlx(
			movieModel, minfoModel, filmModel,
			labelModel, makerModel, directorModel, prefixModel,
			castModel, genreModel, movieCastModel, movieGenreModel,
			vdirModel,
		)

		glistRepo  = film_infra.NewGListRepoSqlx(moviex.NewGListModel(conn, c))
		scRepo     = film_infra.NewScRepoSqlx(moviex.NewGScModel(conn, c))
		recordRepo = movie_infra.NewRecordRepoSqlx(moviex.NewERecordModel(conn, c))
	)

	bizRedis, err := redis.NewRedis(redis.RedisConf{
		Host: cfg.BizRedis.Host,
		Pass: cfg.BizRedis.Pass,
		Type: cfg.BizRedis.Type,
	})
	if err != nil {
		return nil, err
	}
	movieTypeCache := bizcache.NewMovieTypeBizCache(bizRedis, cfg.MovieTypeCache.Prefix, cfg.MovieTypeCache.Version, cfg.MovieTypeCache.TTL)

	// ========== fetcher ==========
	f := fetcher.NewFetcher(fetcher.Config{
		UserAgent: cfg.Fetcher.UserAgent,
		Cookie:    cfg.Fetcher.Cookie,
		Proxy:     cfg.Fetcher.Proxy,
		Timeout:   15 * time.Second,
	})

	return &Deps{
		Config: cfg,
		Log:    mylog.NewLogrusLogger(cfg.LogursLevel),

		SqlConn: conn,
		Cache:   c,

		SeedRepo:      seedRepo,
		InventoryRepo: inventoryRepo,
		DetailRepo:    detailRepo,
		BestinvRepo:   bestRepo,

		ItemRepo:       itemRepo,
		MovieRepo:      movieRepo,
		MovieListRepo:  movieListRepo,
		CastRepo:       castRepo,
		GenreRepo:      genreRepo,
		DirectorRepo:   directorRepo,
		LabelRepo:      labelRepo,
		MakerRepo:      makerRepo,
		PrefixRepo:     prefixRepo,
		MovieCastRepo:  movieCastRepo,
		MovieGenreRepo: movieGenreRepo,
		MinfoRepo:      minfoRepo,
		MurlRepo:       murlRepo,
		RankRepo:       rankRepo,
		CafoRepo:       cafoRepo,
		MovieTypeCache: movieTypeCache,
		DirectoryRepo:  directoryRepo,
		FilmRepo:       filmRepo,
		GListRepo:      glistRepo,
		ScRepo:         scRepo,

		// ✅ 注入 RecordRepo
		RecordRepo: recordRepo,

		DetailJobs:  make(chan string, 200),
		BestTrigger: make(chan contracts.TriggerMsg, 8),
		FilmTrigger: make(chan contracts.FilmTriggerMsg, 8),
		ScTrigger:   make(chan contracts.ScTriggerMsg, 16),

		Fetcher: f,
	}, nil
}
