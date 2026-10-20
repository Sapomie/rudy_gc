package spiderflowx

import (
	"context"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/model/modelx/spiderx"
)

type Dependencies struct {
	SeedModel       spiderx.DSeedModel
	InventoryModel  spiderx.DInventoryModel
	DetailModel     spiderx.DDetailModel
	BestinvModel    spiderx.DBestinvModel
	ItemModel       moviex.EItemModel
	MovieModel      moviex.AMovieModel
	CastModel       moviex.AmCastModel
	PersonModel     moviex.CPersonModel
	GenreModel      moviex.AmGenreModel
	DirectorModel   moviex.AmDirectorModel
	LabelModel      moviex.AmLabelModel
	MakerModel      moviex.AmMakerModel
	PrefixModel     moviex.AmPrefixModel
	MovieCastModel  moviex.AmrMovieCastModel
	MovieGenreModel moviex.AmrMovieGenreModel
	MinfoModel      moviex.BmMinfoModel
	MurlModel       moviex.BmMurlModel
	RankModel       moviex.CRankModel
	RecordModel     moviex.ERecordModel

	SyncPersonStats func(ctx context.Context, ids []int64, now int64) error
}

type Repositories struct {
	SeedRepo       *SeedRepoSqlx
	InventoryRepo  *InventoryRepoSqlx
	ItemRepo       *ItemRepoSqlx
	DetailRepo     *DetailRepoSqlx
	BestinvRepo    *BestinvRepoSqlx
	MovieRepo      *MovieRepoSqlx
	CastRepo       *CastRepoSqlx
	GenreRepo      *GenreRepoSqlx
	DirectorRepo   *DirectorRepoSqlx
	LabelRepo      *LabelRepoSqlx
	MakerRepo      *MakerRepoSqlx
	PrefixRepo     *PrefixRepoSqlx
	MovieCastRepo  *MovieCastRepoSqlx
	MovieGenreRepo *MovieGenreRepoSqlx
	MinfoRepo      *MinfoRepoSqlx
	MurlRepo       *MurlRepoSqlx
	RankRepo       *RankRepoSqlx
	RecordRepo     *RecordRepoSqlx
}

func NewRepositories(dependencies Dependencies) *Repositories {
	return &Repositories{
		SeedRepo:       &SeedRepoSqlx{m: dependencies.SeedModel},
		InventoryRepo:  &InventoryRepoSqlx{m: dependencies.InventoryModel},
		ItemRepo:       &ItemRepoSqlx{m: dependencies.ItemModel},
		DetailRepo:     &DetailRepoSqlx{m: dependencies.DetailModel},
		BestinvRepo:    &BestinvRepoSqlx{m: dependencies.BestinvModel},
		MovieRepo:      &MovieRepoSqlx{m: dependencies.MovieModel},
		CastRepo:       &CastRepoSqlx{m: dependencies.CastModel, pm: dependencies.PersonModel, syncPersonStats: dependencies.SyncPersonStats},
		GenreRepo:      &GenreRepoSqlx{m: dependencies.GenreModel},
		DirectorRepo:   &DirectorRepoSqlx{m: dependencies.DirectorModel},
		LabelRepo:      &LabelRepoSqlx{m: dependencies.LabelModel},
		MakerRepo:      &MakerRepoSqlx{m: dependencies.MakerModel},
		PrefixRepo:     &PrefixRepoSqlx{m: dependencies.PrefixModel},
		MovieCastRepo:  &MovieCastRepoSqlx{m: dependencies.MovieCastModel},
		MovieGenreRepo: &MovieGenreRepoSqlx{m: dependencies.MovieGenreModel},
		MinfoRepo:      &MinfoRepoSqlx{m: dependencies.MinfoModel},
		MurlRepo:       &MurlRepoSqlx{m: dependencies.MurlModel},
		RankRepo:       &RankRepoSqlx{m: dependencies.RankModel},
		RecordRepo:     &RecordRepoSqlx{m: dependencies.RecordModel},
	}
}
