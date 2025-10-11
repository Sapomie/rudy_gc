package svc

import (
	"rudy_gc/internal/domain/spider/fetcher"
	"rudy_gc/internal/infra/bizcache"
	film_infra "rudy_gc/internal/infra/film_repo"
	"rudy_gc/internal/repo/film_repo"
	log "rudy_gc/pkg/loga"
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

	MovieTypeCache movie_repo.MovieTypeCache

	Fetcher *fetcher.Fetcher
}

func NewDeps(cfg config.Config) (*Deps, error) {
	conn := sqlx.NewMysql(cfg.DataSource)
	c := cfg.Cache

	var (
		movieModel = moviex.NewAMovieModel(conn, c)
		minfoModel = moviex.NewBmMinfoModel(conn, c)
		filmModel  = moviex.NewVFilmModel(conn, c)
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
		castRepo       = movie_infra.NewCastRepoSqlx(moviex.NewAmCastModel(conn, c))
		genreRepo      = movie_infra.NewGenreRepoSqlx(moviex.NewAmGenreModel(conn, c))
		directorRepo   = movie_infra.NewDirectorRepoSqlx(moviex.NewAmDirectorModel(conn, c))
		labelRepo      = movie_infra.NewLabelRepoSqlx(moviex.NewAmLabelModel(conn, c))
		makerRepo      = movie_infra.NewMakerRepoSqlx(moviex.NewAmMakerModel(conn, c))
		prefixRepo     = movie_infra.NewPrefixRepoSqlx(moviex.NewAmPrefixModel(conn, c))
		movieCastRepo  = movie_infra.NewMovieCastRepoSqlx(moviex.NewAmrMovieCastModel(conn, c))
		movieGenreRepo = movie_infra.NewMovieGenreRepoSqlx(moviex.NewAmrMovieGenreModel(conn, c))
		minfoRepo      = movie_infra.NewMinfoRepoSqlx(minfoModel)
		murlRepo       = movie_infra.NewMurlRepoSqlx(moviex.NewBmMurlModel(conn, c))
		itemRepo       = movie_infra.NewItemRepoSqlx(moviex.NewEItemModel(conn, c))
		rankRepo       = movie_infra.NewRankRepoSqlx(moviex.NewCRankModel(conn, c))
		cafoRepo       = movie_infra.NewCafoRepoSqlx(moviex.NewCCafoModel(conn, c))
		directoryRepo  = film_infra.NewDirectoryRepoSqlx(moviex.NewVDirectoryModel(conn, c))
		filmRepo       = film_infra.NewFilmRepoSqlx(filmModel)
		movieListRepo  = movie_infra.NewMovieListRepoSqlx(movieModel, minfoModel, filmModel)
		glistRepo      = film_infra.NewGListRepoSqlx(moviex.NewGListModel(conn, c))
		scRepo         = film_infra.NewScRepoSqlx(moviex.NewGScModel(conn, c))
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
		Log:    log.NewLogrusLogger(cfg.LogursLevel),

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

		Fetcher: f,
	}, nil
}
