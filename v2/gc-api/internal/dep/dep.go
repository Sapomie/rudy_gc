package dep

import (
	"rudy-gc-api/internal/config"
	"rudy-gc-api/internal/model/modelx"

	"github.com/sirupsen/logrus"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Dep struct {
	Config config.Config
	Cache  cache.CacheConf
	Log    *logrus.Logger
	Conn   sqlx.SqlConn

	AMovieModel               modelx.AMovieModel
	AmCastModel               modelx.AmCastModel
	AmDirectorModel           modelx.AmDirectorModel
	AmGenreModel              modelx.AmGenreModel
	AmLabelModel              modelx.AmLabelModel
	AmMakerModel              modelx.AmMakerModel
	AmPrefixModel             modelx.AmPrefixModel
	AmrMovieCastModel         modelx.AmrMovieCastModel
	AmrMovieGenreModel        modelx.AmrMovieGenreModel
	BmMinfoModel              modelx.BmMinfoModel
	BmMurlModel               modelx.BmMurlModel
	CMovieAlbumModel          modelx.CMovieAlbumModel
	CMovieAlbumItemModel      modelx.CMovieAlbumItemModel
	CPersonModel              modelx.CPersonModel
	CPersonScModel            modelx.CPersonScModel
	CRankModel                modelx.CRankModel
	CRankPeriodModel          modelx.CRankPeriodModel
	CRankPeriodItemModel      modelx.CRankPeriodItemModel
	EDeletedMovieModel        modelx.EDeletedMovieModel
	EItemModel                modelx.EItemModel
	EScRecordModel            modelx.ERecordModel
	GScModel                  modelx.GScModel
	GScStatModel              modelx.GScStatModel
	TJavbusMagnetModel        modelx.TJavbusMagnetModel
	TJavbusMagnetFetchModel   modelx.TJavbusMagnetFetchModel
	TSehuatangMagnetModel     modelx.TSehuatangMagnetModel
	TSukebeiTorrentModel      modelx.TSukebeiTorrentModel
	TSukebeiTorrentFetchModel modelx.TSukebeiTorrentFetchModel
	WFolderModel              modelx.WFolderModel
	WMediaModel               modelx.WMediaModel
}

func Build(cfg config.Config) *Dep {
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)

	dp := &Dep{
		Config: cfg,
		Cache:  cfg.Cache,
		Log:    log,
	}

	if cfg.DataSource == "" {
		return dp
	}

	conn := sqlx.NewMysql(cfg.DataSource)
	dp.Conn = conn
	dp.AMovieModel = modelx.NewAMovieModel(conn, cfg.Cache)
	dp.AmCastModel = modelx.NewAmCastModel(conn, cfg.Cache)
	dp.AmDirectorModel = modelx.NewAmDirectorModel(conn, cfg.Cache)
	dp.AmGenreModel = modelx.NewAmGenreModel(conn, cfg.Cache)
	dp.AmLabelModel = modelx.NewAmLabelModel(conn, cfg.Cache)
	dp.AmMakerModel = modelx.NewAmMakerModel(conn, cfg.Cache)
	dp.AmPrefixModel = modelx.NewAmPrefixModel(conn, cfg.Cache)
	dp.AmrMovieCastModel = modelx.NewAmrMovieCastModel(conn, cfg.Cache)
	dp.AmrMovieGenreModel = modelx.NewAmrMovieGenreModel(conn, cfg.Cache)
	dp.BmMinfoModel = modelx.NewBmMinfoModel(conn, cfg.Cache)
	dp.BmMurlModel = modelx.NewBmMurlModel(conn, cfg.Cache)
	dp.CMovieAlbumModel = modelx.NewCMovieAlbumModel(conn, cfg.Cache)
	dp.CMovieAlbumItemModel = modelx.NewCMovieAlbumItemModel(conn, cfg.Cache)
	dp.CPersonModel = modelx.NewCPersonModel(conn, cfg.Cache)
	dp.CPersonScModel = modelx.NewCPersonScModel(conn, cfg.Cache)
	dp.CRankModel = modelx.NewCRankModel(conn, cfg.Cache)
	dp.CRankPeriodModel = modelx.NewCRankPeriodModel(conn, cfg.Cache)
	dp.CRankPeriodItemModel = modelx.NewCRankPeriodItemModel(conn, cfg.Cache)
	dp.EDeletedMovieModel = modelx.NewEDeletedMovieModel(conn, cfg.Cache)
	dp.EItemModel = modelx.NewEItemModel(conn, cfg.Cache)
	dp.EScRecordModel = modelx.NewERecordModel(conn, cfg.Cache)
	dp.GScModel = modelx.NewGScModel(conn, cfg.Cache)
	dp.GScStatModel = modelx.NewGScStatModel(conn, cfg.Cache)
	dp.TJavbusMagnetModel = modelx.NewTJavbusMagnetModel(conn, cfg.Cache)
	dp.TJavbusMagnetFetchModel = modelx.NewTJavbusMagnetFetchModel(conn, cfg.Cache)
	dp.TSehuatangMagnetModel = modelx.NewTSehuatangMagnetModel(conn, cfg.Cache)
	dp.TSukebeiTorrentModel = modelx.NewTSukebeiTorrentModel(conn, cfg.Cache)
	dp.TSukebeiTorrentFetchModel = modelx.NewTSukebeiTorrentFetchModel(conn, cfg.Cache)
	dp.WFolderModel = modelx.NewWFolderModel(conn, cfg.Cache)
	dp.WMediaModel = modelx.NewWMediaModel(conn, cfg.Cache)

	return dp
}
