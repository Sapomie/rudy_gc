package migrate

import (
	"rudy_gc/internal/svc"
	"rudy_gc/oldmodel/modelx"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type XModel struct {
	ItemModel       modelx.AItemModel
	MovieModel      modelx.AMovieModel
	MurlModel       modelx.BmMurlModel
	CastModel       modelx.BmCastModel
	GenreModel      modelx.BmGenreModel
	DirectorModel   modelx.BmDirectorModel
	LabelModel      modelx.BmLabelModel
	MakerModel      modelx.BmMakerModel
	PrefixModel     modelx.BmPrefixModel
	MovieCastModel  modelx.BmrMovieCastModel
	MovieGenreModel modelx.BmrMovieGenreModel
	RankModel       modelx.CRankModel
	CafoModel       modelx.CCafoModel
	InventoryModel  modelx.RawInventoryModel
	BestInvModel    modelx.RawBestinvModel
	DetailModel     modelx.RawDetailModel
	FilmModel       modelx.VFilmModel
	AlbumModel      modelx.VAlbumModel
	VideoModel      modelx.VVideoModel
	QueryModel      modelx.DQueryModel
	RecordModel     modelx.DRecordModel
	GListModel      modelx.GListModel
	GScModel        modelx.GScModel
}

type Service struct {
	deps   *svc.Deps
	xModel *XModel
}

func New(deps *svc.Deps) *Service {
	return &Service{
		deps:   deps,
		xModel: NewXModel(deps),
	}
}

func NewXModel(deps *svc.Deps) *XModel {
	oldDsn := "root:4521822123@tcp(127.0.0.1:3306)/zero_gc_v2_b?charset=utf8mb4"
	dbSqlx := sqlx.NewMysql(oldDsn)

	xm := XModel{
		ItemModel:  modelx.NewAItemModel(dbSqlx, deps.Config.Cache),
		MovieModel: modelx.NewAMovieModel(dbSqlx, deps.Config.Cache),

		MurlModel:       modelx.NewBmMurlModel(dbSqlx, deps.Config.Cache),
		CastModel:       modelx.NewBmCastModel(dbSqlx, deps.Config.Cache),
		GenreModel:      modelx.NewBmGenreModel(dbSqlx, deps.Config.Cache),
		DirectorModel:   modelx.NewBmDirectorModel(dbSqlx, deps.Config.Cache),
		LabelModel:      modelx.NewBmLabelModel(dbSqlx, deps.Config.Cache),
		MakerModel:      modelx.NewBmMakerModel(dbSqlx, deps.Config.Cache),
		PrefixModel:     modelx.NewBmPrefixModel(dbSqlx, deps.Config.Cache),
		MovieCastModel:  modelx.NewBmrMovieCastModel(dbSqlx, deps.Config.Cache),
		MovieGenreModel: modelx.NewBmrMovieGenreModel(dbSqlx, deps.Config.Cache),

		RankModel: modelx.NewCRankModel(dbSqlx, deps.Config.Cache),
		CafoModel: modelx.NewCCafoModel(dbSqlx, deps.Config.Cache),

		InventoryModel: modelx.NewRawInventoryModel(dbSqlx),
		BestInvModel:   modelx.NewRawBestinvModel(dbSqlx),
		DetailModel:    modelx.NewRawDetailModel(dbSqlx),

		FilmModel:   modelx.NewVFilmModel(dbSqlx),
		VideoModel:  modelx.NewVVideoModel(dbSqlx),
		AlbumModel:  modelx.NewVAlbumModel(dbSqlx),
		QueryModel:  modelx.NewDQueryModel(dbSqlx),
		RecordModel: modelx.NewDRecordModel(dbSqlx),

		GListModel: modelx.NewGListModel(dbSqlx),
		GScModel:   modelx.NewGScModel(dbSqlx),
	}

	return &xm
}
