package migrate

import (
	"context"
	"rudy_gc/internal/types"
	"time"
)

// removed
func (s *Service) MigrateFilm() error {
	ctx := context.Background()

	oldFilms, err := s.xModel.FilmModel.FindAll(ctx, 2)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	for _, f := range oldFilms {

		movie, err := s.deps.MovieRepo.FindOneByJavId(ctx, f.MovieJavId)
		if err != nil {
			return err
		}

		newFilm := types.Film{
			MovieJavId:    f.MovieJavId,
			MovieName:     f.Name,
			FileName:      f.Name,
			DirectoryId:   0,
			RootDir:       "-",
			FullDir:       "-",
			Dir1Id:        0,
			Dir2Id:        0,
			Dir3Id:        0,
			Dir4Id:        0,
			Alias:         f.Alias,
			Size:          0,
			Width:         0,
			Height:        0,
			BitRate:       0,
			Duration:      0,
			FrameAverage:  0,
			HasSub:        0,
			SelfMake:      0,
			HasMask:       0,
			NeedScanMeta:  0,
			IsRemoved:     f.IsRemoved,
			RemoveTime:    f.RemoveTime,
			ScTimes:       f.ScTimes,
			ComeTimes:     f.ComeTimes,
			LastScTime:    f.LastScTime,
			BirthTime:     f.BirthTime,
			ReleasingDate: movie.ReleasingDate,
			CreatedOn:     now,
			UpdatedOn:     now,
		}

		_, _, err = s.deps.FilmRepo.UpsertFilm(ctx, &newFilm)
		if err != nil {
			return err
		}
		s.movieSvc.InvalidateMovieType(ctx, f.MovieJavId)

	}

	return nil

}
