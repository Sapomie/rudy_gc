// internal/svc/deps.go

package svc

import (
	"context"
	"rudy_gc/internal/contracts"
	"rudy_gc/pkg/mylog"
	"sync"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/model/modelx/spiderx"

	"rudy_gc/internal/config"
	"rudy_gc/internal/service/spider/fetcher"

	"github.com/sirupsen/logrus"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Deps struct {
	BestTrigger chan contracts.TriggerMsg
	ScTrigger   chan contracts.ScTriggerMsg
	// ...
	Config  config.Config
	SqlConn sqlx.SqlConn
	Cache   cache.CacheConf
	Log     *logrus.Logger

	// moviex models（新 service 直连用）
	MovieModel                           moviex.AMovieModel
	MinfoModel                           moviex.BmMinfoModel
	MurlModel                            moviex.BmMurlModel
	ItemModel                            moviex.EItemModel
	DeletedMovieModel                    moviex.EDeletedMovieModel
	CastModel                            moviex.AmCastModel
	GenreModel                           moviex.AmGenreModel
	DirectorModel                        moviex.AmDirectorModel
	LabelModel                           moviex.AmLabelModel
	MakerModel                           moviex.AmMakerModel
	PrefixModel                          moviex.AmPrefixModel
	MovieCastModel                       moviex.AmrMovieCastModel
	MovieGenreModel                      moviex.AmrMovieGenreModel
	WFolderModel                         moviex.WFolderModel
	WKvModel                             moviex.WKvModel
	WMediaModel                          moviex.WMediaModel
	WMediaBirthBucketStatModel           moviex.WMediaBirthBucketStatModel
	WMediaBirthTopStatModel              moviex.WMediaBirthTopStatModel
	MovieReleaseBucketStatModel          moviex.MovieReleaseBucketStatModel
	MovieReleaseTopStatModel             moviex.MovieReleaseTopStatModel
	WAggEventModel                       moviex.WAggEventModel
	CPersonScModel                       moviex.CPersonScModel
	GListModel                           moviex.GListModel
	ScModel                              moviex.GScModel
	ScEventWriteModel                    moviex.ScEventWriteModel
	RankModel                            moviex.CRankModel
	RankPeriodModel                      moviex.CRankPeriodModel
	RankPeriodItemModel                  moviex.CRankPeriodItemModel
	PersonModel                          moviex.CPersonModel
	GScStatModel                         moviex.GScStatModel
	CafoModel                            moviex.CCafoModel
	RecordModel                          moviex.ERecordModel
	AlbumModel                           moviex.TAlbumModel
	AlbumItemModel                       moviex.TmAlbumItemModel
	MovieAlbumModel                      moviex.CMovieAlbumModel
	MovieAlbumItemModel                  moviex.CMovieAlbumItemModel
	JavbusMagnetModel                    moviex.TJavbusMagnetModel
	ShMagnetModel                        moviex.ShMagnetModel
	ShForumModel                         moviex.ShForumModel
	FC2MovieModel                        moviex.FC2MovieModel
	FC2FolderModel                       moviex.FC2FolderModel
	FC2MediaModel                        moviex.FC2MediaModel
	FC2MediaImageModel                   moviex.FC2MediaImageModel
	FC2ScStatModel                       moviex.FC2ScStatModel
	ShSehuatangThreadModel               moviex.ShSehuatangThreadModel
	ShThreadBucketStatModel              moviex.ShThreadBucketStatModel
	ShSehuatangFetchRunModel             moviex.ShSehuatangFetchRunModel
	ShSehuatangFetchItemModel            moviex.ShSehuatangFetchItemModel
	ShSehuatangThreadListFetchRunModel   moviex.ShSehuatangThreadListFetchRunModel
	ShSehuatangThreadDetailFetchRunModel moviex.ShSehuatangThreadDetailFetchRunModel
	JavbusMagnetFetchModel               moviex.TJavbusMagnetFetchModel
	SukebeiTorrentModel                  moviex.TSukebeiTorrentModel
	SukebeiTorrentFetchModel             moviex.TSukebeiTorrentFetchModel
	SeedModel                            spiderx.DSeedModel
	InventoryModel                       spiderx.DInventoryModel
	DetailModel                          spiderx.DDetailModel
	BestinvModel                         spiderx.DBestinvModel
	FetchSiteModel                       moviex.TFetchSiteModel

	MovieTypeCache MovieTypeCache

	Fetcher      *fetcher.Fetcher
	FetchSites   map[string]FetchSiteConfig
	fetchSitesMu sync.RWMutex
	DetailJobs   chan string
}

func NewDeps(cfg config.Config) (*Deps, error) {
	conn := sqlx.NewMysql(cfg.DataSource)
	c := cfg.Cache
	logger := mylog.NewLogrusLogger(cfg.LogursLevel)
	mylog.EnsureTaskHook(logger)

	var (
		movieModel                           = moviex.NewAMovieModel(conn, c)
		minfoModel                           = moviex.NewBmMinfoModel(conn, c)
		murlModel                            = moviex.NewBmMurlModel(conn, c)
		itemModel                            = moviex.NewEItemModel(conn, c)
		deletedMovieModel                    = moviex.NewEDeletedMovieModel(conn, c)
		rankModel                            = moviex.NewCRankModel(conn, c)
		personModel                          = moviex.NewCPersonModel(conn, c)
		gScStatModel                         = moviex.NewGScStatModel(conn, c)
		cafoModel                            = moviex.NewCCafoModel(conn, c)
		glistModel                           = moviex.NewGListModel(conn, c)
		scModel                              = moviex.NewGScModel(conn, c)
		scEventWriteModel                    = moviex.NewScEventWriteModel(conn, glistModel, scModel)
		recordModel                          = moviex.NewERecordModel(conn, c)
		albumModel                           = moviex.NewTAlbumModel(conn, c)
		albumItemModel                       = moviex.NewTmAlbumItemModel(conn, c)
		movieAlbumModel                      = moviex.NewCMovieAlbumModel(conn, c)
		movieAlbumItemModel                  = moviex.NewCMovieAlbumItemModel(conn, c)
		javbusMagnetModel                    = moviex.NewTJavbusMagnetModel(conn, c)
		shMagnetModel                        = moviex.NewShMagnetModel(conn)
		shForumModel                         = moviex.NewShForumModel(conn)
		fc2MovieModel                        = moviex.NewFC2MovieModel(conn)
		fc2FolderModel                       = moviex.NewFC2FolderModel(conn)
		fc2MediaModel                        = moviex.NewFC2MediaModel(conn)
		fc2MediaImageModel                   = moviex.NewFC2MediaImageModel(conn)
		fc2ScStatModel                       = moviex.NewFC2ScStatModel(conn, c)
		shSehuatangThreadModel               = moviex.NewShSehuatangThreadModel(conn)
		shThreadBucketStatModel              = moviex.NewShThreadBucketStatModel(conn)
		shSehuatangFetchRunModel             = moviex.NewShSehuatangFetchRunModel(conn)
		shSehuatangFetchItemModel            = moviex.NewShSehuatangFetchItemModel(conn)
		shSehuatangThreadListFetchRunModel   = moviex.NewShSehuatangThreadListFetchRunModel(conn)
		shSehuatangThreadDetailFetchRunModel = moviex.NewShSehuatangThreadDetailFetchRunModel(conn)
		javbusMagnetFetchModel               = moviex.NewTJavbusMagnetFetchModel(conn, c)
		sukebeiTorrentModel                  = moviex.NewTSukebeiTorrentModel(conn, c)
		sukebeiTorrentFetchModel             = moviex.NewTSukebeiTorrentFetchModel(conn, c)
		rankPeriodModel                      = moviex.NewCRankPeriodModel(conn, c)
		rankPeriodItemModel                  = moviex.NewCRankPeriodItemModel(conn, c)
		seedModel                            = spiderx.NewDSeedModel(conn)
		invModel                             = spiderx.NewDInventoryModel(conn)
		detailModel                          = spiderx.NewDDetailModel(conn)
		bestModel                            = spiderx.NewDBestinvModel(conn)
		fetchSiteModel                       = moviex.NewTFetchSiteModel(conn, c)

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
		wFolderModel                = moviex.NewWFolderModel(conn, c)
		wKvModel                    = moviex.NewWKvModel(conn, c)
		wMediaModel                 = moviex.NewWMediaModel(conn, c)
		wMediaBirthBucketStatModel  = moviex.NewWMediaBirthBucketStatModel(conn, c)
		wMediaBirthTopStatModel     = moviex.NewWMediaBirthTopStatModel(conn, c)
		movieReleaseBucketStatModel = moviex.NewMovieReleaseBucketStatModel(conn, c)
		movieReleaseTopStatModel    = moviex.NewMovieReleaseTopStatModel(conn, c)
		wAggEventModel              = moviex.NewWAggEventModel(conn, c)
		cPersonScModel              = moviex.NewCPersonScModel(conn, c)
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
		Timeout:   time.Duration(cfg.Fetcher.Timeout) * time.Second,
	})
	fetchSites := loadFetchSiteConfigs(context.Background(), cfg, fetchSiteModel, logger)
	for siteCode, siteCfg := range fetchSites {
		f.SetSiteConfig(siteCode, fetcher.Config{
			UserAgent: siteCfg.UserAgent,
			Cookie:    siteCfg.Cookie,
			Proxy:     siteCfg.Proxy,
			Timeout:   siteCfg.Timeout,
		})
	}

	return &Deps{
		Config: cfg,
		Log:    logger,

		SqlConn: conn,
		Cache:   c,

		MovieModel:                           movieModel,
		MinfoModel:                           minfoModel,
		MurlModel:                            murlModel,
		ItemModel:                            itemModel,
		DeletedMovieModel:                    deletedMovieModel,
		CastModel:                            castModel,
		GenreModel:                           genreModel,
		DirectorModel:                        directorModel,
		LabelModel:                           labelModel,
		MakerModel:                           makerModel,
		PrefixModel:                          prefixModel,
		MovieCastModel:                       movieCastModel,
		MovieGenreModel:                      movieGenreModel,
		WFolderModel:                         wFolderModel,
		WKvModel:                             wKvModel,
		WMediaModel:                          wMediaModel,
		WMediaBirthBucketStatModel:           wMediaBirthBucketStatModel,
		WMediaBirthTopStatModel:              wMediaBirthTopStatModel,
		MovieReleaseBucketStatModel:          movieReleaseBucketStatModel,
		MovieReleaseTopStatModel:             movieReleaseTopStatModel,
		WAggEventModel:                       wAggEventModel,
		CPersonScModel:                       cPersonScModel,
		GListModel:                           glistModel,
		ScModel:                              scModel,
		ScEventWriteModel:                    scEventWriteModel,
		RankModel:                            rankModel,
		RankPeriodModel:                      rankPeriodModel,
		RankPeriodItemModel:                  rankPeriodItemModel,
		PersonModel:                          personModel,
		GScStatModel:                         gScStatModel,
		CafoModel:                            cafoModel,
		RecordModel:                          recordModel,
		AlbumModel:                           albumModel,
		AlbumItemModel:                       albumItemModel,
		MovieAlbumModel:                      movieAlbumModel,
		MovieAlbumItemModel:                  movieAlbumItemModel,
		JavbusMagnetModel:                    javbusMagnetModel,
		ShMagnetModel:                        shMagnetModel,
		ShForumModel:                         shForumModel,
		FC2MovieModel:                        fc2MovieModel,
		FC2FolderModel:                       fc2FolderModel,
		FC2MediaModel:                        fc2MediaModel,
		FC2MediaImageModel:                   fc2MediaImageModel,
		FC2ScStatModel:                       fc2ScStatModel,
		ShSehuatangThreadModel:               shSehuatangThreadModel,
		ShThreadBucketStatModel:              shThreadBucketStatModel,
		ShSehuatangFetchRunModel:             shSehuatangFetchRunModel,
		ShSehuatangFetchItemModel:            shSehuatangFetchItemModel,
		ShSehuatangThreadListFetchRunModel:   shSehuatangThreadListFetchRunModel,
		ShSehuatangThreadDetailFetchRunModel: shSehuatangThreadDetailFetchRunModel,
		JavbusMagnetFetchModel:               javbusMagnetFetchModel,
		SukebeiTorrentModel:                  sukebeiTorrentModel,
		SukebeiTorrentFetchModel:             sukebeiTorrentFetchModel,
		SeedModel:                            seedModel,
		InventoryModel:                       invModel,
		DetailModel:                          detailModel,
		BestinvModel:                         bestModel,
		FetchSiteModel:                       fetchSiteModel,
		MovieTypeCache:                       movieTypeCache,

		DetailJobs:  make(chan string, 200),
		BestTrigger: make(chan contracts.TriggerMsg, 8),
		ScTrigger:   make(chan contracts.ScTriggerMsg, 16),

		Fetcher:    f,
		FetchSites: fetchSites,
	}, nil
}
