package svc

import (
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/data/modelx/spiderx"

	"rudy_gc/internal/config"
	"rudy_gc/internal/infra/movie_infra"
	"rudy_gc/internal/infra/spider_infra"
	"rudy_gc/internal/repo/movie_repo"
	"rudy_gc/internal/repo/spider_repo"
	"rudy_gc/internal/spider/fetcher"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Deps struct {
	Config config.Config

	// spider repos (无缓存)
	SeedRepo      spider_repo.SeedRepo
	InventoryRepo spider_repo.InventoryRepo
	ItemRepo      spider_repo.ItemRepo
	DetailRepo    spider_repo.DetailRepo

	// movie repos (带缓存)
	MovieRepo      movie_repo.MovieRepo
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

	Fetcher *fetcher.Fetcher
}

func NewDeps(cfg config.Config) *Deps {
	conn := sqlx.NewMysql(cfg.DataSource)
	c := cfg.Cache

	// ========== spider (无缓存) ==========
	seedModel := spiderx.NewDSeedModel(conn)
	invModel := spiderx.NewDInventoryModel(conn)
	itemModel := spiderx.NewEItemModel(conn)
	detailModel := spiderx.NewDDetailModel(conn)

	seedRepo := spider_infra.NewSeedRepoSqlx(seedModel)
	inventoryRepo := spider_infra.NewInventoryRepoSqlx(invModel)
	itemRepo := spider_infra.NewItemRepoSqlx(itemModel)
	detailRepo := spider_infra.NewDetailRepoSqlx(detailModel)

	movieRepo := movie_infra.NewMovieRepoSqlx(moviex.NewAMovieModel(conn, c))
	castRepo := movie_infra.NewCastRepoSqlx(moviex.NewAmCastModel(conn, c))
	genreRepo := movie_infra.NewGenreRepoSqlx(moviex.NewAmGenreModel(conn, c))
	directorRepo := movie_infra.NewDirectorRepoSqlx(moviex.NewAmDirectorModel(conn, c))
	labelRepo := movie_infra.NewLabelRepoSqlx(moviex.NewAmLabelModel(conn, c))
	makerRepo := movie_infra.NewMakerRepoSqlx(moviex.NewAmMakerModel(conn, c))
	prefixRepo := movie_infra.NewPrefixRepoSqlx(moviex.NewAmPrefixModel(conn, c))
	movieCastRepo := movie_infra.NewMovieCastRepoSqlx(moviex.NewAmrMovieCastModel(conn, c))
	movieGenreRepo := movie_infra.NewMovieGenreRepoSqlx(moviex.NewAmrMovieGenreModel(conn, c))
	minfoRepo := movie_infra.NewMinfoRepoSqlx(moviex.NewBmMinfoModel(conn, c))
	murlRepo := movie_infra.NewMurlRepoSqlx(moviex.NewBmMurlModel(conn, c))

	// ========== fetcher ==========
	f := fetcher.NewFetcher(fetcher.Config{
		UserAgent: cfg.Fetcher.UserAgent,
		Cookie:    cfg.Fetcher.Cookie,
		Proxy:     cfg.Fetcher.Proxy,
		Timeout:   15 * time.Second,
	})

	return &Deps{
		Config:        cfg,
		SeedRepo:      seedRepo,
		InventoryRepo: inventoryRepo,
		ItemRepo:      itemRepo,
		DetailRepo:    detailRepo,

		MovieRepo:      movieRepo,
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

		Fetcher: f,
	}
}
