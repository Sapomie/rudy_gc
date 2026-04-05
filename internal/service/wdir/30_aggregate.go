package wdir

import "rudy_gc/internal/types"

func buildDirStatsFromAll(all []*types.Film, recursive bool) *types.DirStats {
	st := &types.DirStats{Recursive: recursive}
	if len(all) == 0 {
		return st
	}

	var (
		sumSize      int64
		maxBirthTime int64
		maxUpdatedOn int64
	)
	for _, film := range all {
		sumSize += film.Size
		if film.BirthTime > maxBirthTime {
			maxBirthTime = film.BirthTime
		}
		if film.UpdatedOn > maxUpdatedOn {
			maxUpdatedOn = film.UpdatedOn
		}
	}
	st.FilmCount = int64(len(all))
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
			ds = &types.DirStats{Recursive: false}
			res[film.DirectoryId] = ds
		}
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
			ds = &types.DirStats{Recursive: recursive}
		}
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
