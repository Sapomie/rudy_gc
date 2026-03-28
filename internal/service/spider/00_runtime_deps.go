package spider

import "rudy_gc/internal/svc"

type runtimeDeps struct {
	*svc.Deps

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
	CafoRepo       *CafoRepoSqlx
	RecordRepo     *RecordRepoSqlx
}

func newRuntimeDeps(base *svc.Deps) *runtimeDeps {
	return &runtimeDeps{
		Deps:           base,
		SeedRepo:       &SeedRepoSqlx{m: base.SeedModel},
		InventoryRepo:  &InventoryRepoSqlx{m: base.InventoryModel},
		ItemRepo:       &ItemRepoSqlx{m: base.ItemModel},
		DetailRepo:     &DetailRepoSqlx{m: base.DetailModel},
		BestinvRepo:    &BestinvRepoSqlx{m: base.BestinvModel},
		MovieRepo:      &MovieRepoSqlx{m: base.MovieModel},
		CastRepo:       &CastRepoSqlx{m: base.CastModel, pm: base.PersonModel, syncPersonStats: base.SyncPersonStatsByIDs},
		GenreRepo:      &GenreRepoSqlx{m: base.GenreModel},
		DirectorRepo:   &DirectorRepoSqlx{m: base.DirectorModel},
		LabelRepo:      &LabelRepoSqlx{m: base.LabelModel},
		MakerRepo:      &MakerRepoSqlx{m: base.MakerModel},
		PrefixRepo:     &PrefixRepoSqlx{m: base.PrefixModel},
		MovieCastRepo:  &MovieCastRepoSqlx{m: base.MovieCastModel},
		MovieGenreRepo: &MovieGenreRepoSqlx{m: base.MovieGenreModel},
		MinfoRepo:      &MinfoRepoSqlx{m: base.MinfoModel},
		MurlRepo:       &MurlRepoSqlx{m: base.MurlModel},
		RankRepo:       &RankRepoSqlx{m: base.RankModel},
		CafoRepo:       &CafoRepoSqlx{m: base.CafoModel},
		RecordRepo:     &RecordRepoSqlx{m: base.RecordModel},
	}
}
