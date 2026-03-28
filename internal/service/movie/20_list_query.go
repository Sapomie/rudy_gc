package movie

import "rudy_gc/internal/model/modelx/moviex"

func (s *Service) movieListQuery() *moviex.MovieListRepoSqlx {
	return moviex.NewMovieListRepoSqlx(
		s.deps.MovieModel,
		s.deps.MinfoModel,
		s.deps.FilmModel,
		s.deps.LabelModel,
		s.deps.MakerModel,
		s.deps.DirectorModel,
		s.deps.PrefixModel,
		s.deps.CastModel,
		s.deps.GenreModel,
		s.deps.MovieCastModel,
		s.deps.MovieGenreModel,
		s.deps.DirectoryModel,
	)
}
