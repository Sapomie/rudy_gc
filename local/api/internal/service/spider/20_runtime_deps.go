package spider

import (
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/spiderflowx"
	"rudy_gc/local/api/internal/service/personstat"
)

type runtimeDeps struct {
	*dep.Dep
	*spiderflowx.Repositories
}

func newRuntimeDeps(base *dep.Dep, personStatSvc *personstat.Service) *runtimeDeps {
	return &runtimeDeps{
		Dep: base,
		Repositories: spiderflowx.NewRepositories(spiderflowx.Dependencies{
			SeedModel:       base.SeedModel,
			InventoryModel:  base.InventoryModel,
			DetailModel:     base.DetailModel,
			BestinvModel:    base.BestinvModel,
			ItemModel:       base.ItemModel,
			MovieModel:      base.MovieModel,
			CastModel:       base.CastModel,
			PersonModel:     base.PersonModel,
			GenreModel:      base.GenreModel,
			DirectorModel:   base.DirectorModel,
			LabelModel:      base.LabelModel,
			MakerModel:      base.MakerModel,
			PrefixModel:     base.PrefixModel,
			MovieCastModel:  base.MovieCastModel,
			MovieGenreModel: base.MovieGenreModel,
			MinfoModel:      base.MinfoModel,
			MurlModel:       base.MurlModel,
			RankModel:       base.RankModel,
			RecordModel:     base.RecordModel,
			SyncPersonStats: personStatSvc.SyncByIDs,
		}),
	}
}
