package wdir

import (
	moviedomain "rudy_gc/internal/domain/movie"
	"rudy_gc/internal/types"
)

const (
	directoryMediaTypeEmpty = "empty"
	directoryMediaTypeMixed = "mixed"
)

func buildDirStatsFromAll(all []*types.Film, recursive bool) *types.DirStats {
	st := newDirStats(recursive)
	if len(all) == 0 {
		return st
	}

	var (
		sumSize      int64
		maxBirthTime int64
		maxUpdatedOn int64
	)
	for _, film := range all {
		if film == nil {
			continue
		}
		st.MediaType = mergeDirectoryMediaType(st.MediaType, film.MovieType)
		st.FilmCount++
		sumSize += film.Size
		if film.BirthTime > maxBirthTime {
			maxBirthTime = film.BirthTime
		}
		if film.UpdatedOn > maxUpdatedOn {
			maxUpdatedOn = film.UpdatedOn
		}
	}
	st.TotalSize = sumSize
	st.LastFilmBirth = maxBirthTime
	st.LastUpdatedOn = maxUpdatedOn
	return st
}

func aggregateDirStatsByLeaf(all []*types.Film) map[int64]*types.DirStats {
	res := make(map[int64]*types.DirStats, 64)
	for _, film := range all {
		if film == nil {
			continue
		}
		ds := res[film.DirectoryId]
		if ds == nil {
			ds = newDirStats(false)
			res[film.DirectoryId] = ds
		}
		ds.MediaType = mergeDirectoryMediaType(ds.MediaType, film.MovieType)
		ds.FilmCount++
		ds.TotalSize += film.Size
		if film.BirthTime > ds.LastFilmBirth {
			ds.LastFilmBirth = film.BirthTime
		}
		if film.UpdatedOn > ds.LastUpdatedOn {
			ds.LastUpdatedOn = film.UpdatedOn
		}
	}
	return res
}

func aggregateDirStatsForIDs(all []*types.Film, ids []int64, recursive bool) *types.DirStats {
	if len(ids) == 0 {
		return nil
	}

	set := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}

	var ds *types.DirStats
	for _, film := range all {
		if film == nil {
			continue
		}
		if _, ok := set[film.DirectoryId]; !ok {
			continue
		}
		if ds == nil {
			ds = newDirStats(recursive)
		}
		ds.MediaType = mergeDirectoryMediaType(ds.MediaType, film.MovieType)
		ds.FilmCount++
		ds.TotalSize += film.Size
		if film.BirthTime > ds.LastFilmBirth {
			ds.LastFilmBirth = film.BirthTime
		}
		if film.UpdatedOn > ds.LastUpdatedOn {
			ds.LastUpdatedOn = film.UpdatedOn
		}
	}
	return ds
}

func newDirStats(recursive bool) *types.DirStats {
	return &types.DirStats{
		Recursive: recursive,
		MediaType: directoryMediaTypeEmpty,
	}
}

func mergeDirectoryMediaType(current, next string) string {
	if next != moviedomain.SehuatangThreadMovieTypeJAV && next != moviedomain.SehuatangThreadMovieTypeFC2PPV {
		return normalizeDirectoryMediaType(current)
	}
	current = normalizeDirectoryMediaType(current)
	if current == directoryMediaTypeEmpty {
		return next
	}
	if current == next {
		return current
	}
	return directoryMediaTypeMixed
}

func normalizeDirectoryMediaType(value string) string {
	switch value {
	case moviedomain.SehuatangThreadMovieTypeJAV,
		moviedomain.SehuatangThreadMovieTypeFC2PPV,
		directoryMediaTypeMixed:
		return value
	default:
		return directoryMediaTypeEmpty
	}
}
