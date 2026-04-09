package vfilm

import (
	"context"
	"crypto/md5"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

func (s *Service) directoryGetOrCreateChainWithLevels(ctx context.Context, parts []string) ([4]int64, error) {
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		if v := strings.TrimSpace(part); v != "" {
			clean = append(clean, v)
		}
	}

	var levels [4]int64
	if len(clean) == 0 {
		return levels, nil
	}

	now := time.Now().Unix()
	var parentID int64
	fullIDs := make([]int64, 0, len(clean))
	for i, name := range clean {
		if row, err := s.deps.WFolderModel.FindOneByParentIdNameSourceType(ctx, parentID, name, consts.WFolderSourceLegacyVFilm); err == nil && row != nil {
			parentID = row.Id
			fullIDs = append(fullIDs, parentID)
			continue
		}

		depth := int64(i + 1)
		path := "/" + strings.Join(clean[:i+1], "/")
		sum := md5.Sum([]byte(path))
		vd := &moviex.WFolder{
			ParentId:   parentID,
			Name:       name,
			SourceType: consts.WFolderSourceLegacyVFilm,
			Depth:      depth,
			Path:       path,
			PathHash:   string(sum[:]),
			CreatedOn:  now,
			UpdatedOn:  now,
		}
		if _, err := s.deps.WFolderModel.Insert(ctx, vd); err != nil {
			if row2, err2 := s.deps.WFolderModel.FindOneByParentIdNameSourceType(ctx, parentID, name, consts.WFolderSourceLegacyVFilm); err2 == nil && row2 != nil {
				parentID = row2.Id
				fullIDs = append(fullIDs, parentID)
				continue
			}
			return levels, err
		}
		row3, err := s.deps.WFolderModel.FindOneByParentIdNameSourceType(ctx, parentID, name, consts.WFolderSourceLegacyVFilm)
		if err != nil {
			return levels, err
		}
		parentID = row3.Id
		fullIDs = append(fullIDs, parentID)
	}

	k := len(fullIDs)
	if k > 4 {
		k = 4
	}
	for i := 0; i < k; i++ {
		levels[3-i] = fullIDs[len(fullIDs)-1-i]
	}
	return levels, nil
}

func (s *Service) directoryFindOneByID(ctx context.Context, id int64) (*types.Directory, error) {
	row, err := s.deps.WFolderModel.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil || row.SourceType != consts.WFolderSourceLegacyVFilm {
		return nil, nil
	}
	return mapDirectoryModelToTypes(row), nil
}

func (s *Service) directoryFindOneByName(ctx context.Context, name string) (*types.Directory, error) {
	row, err := s.deps.WFolderModel.FindOneByNameSourceType(ctx, name, consts.WFolderSourceLegacyVFilm)
	if err != nil {
		return nil, err
	}
	return mapDirectoryModelToTypes(row), nil
}

func (s *Service) directoryListRoots(ctx context.Context, page, size int) ([]*types.DirSummary, int64, error) {
	return s.directoryListByParent(ctx, 0, page, size)
}

func (s *Service) directoryListChildren(ctx context.Context, parentID int64, page, size int) ([]*types.DirSummary, int64, error) {
	return s.directoryListByParent(ctx, parentID, page, size)
}

func (s *Service) directoryListByParent(ctx context.Context, parentID int64, page, size int) ([]*types.DirSummary, int64, error) {
	rows, total, err := s.deps.WFolderModel.ListByParentSourceType(ctx, parentID, consts.WFolderSourceLegacyVFilm, page, size)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*types.DirSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, &types.DirSummary{
			Id:        row.Id,
			ParentId:  row.ParentId,
			Name:      row.Name,
			Depth:     row.Depth,
			Path:      row.Path,
			UpdatedOn: row.UpdatedOn,
		})
	}
	return out, total, nil
}

func (s *Service) directoryBuildBreadcrumbs(ctx context.Context, id int64) ([]types.Breadcrumb, error) {
	cur, err := s.deps.WFolderModel.FindOne(ctx, id)
	if err != nil || cur == nil || cur.SourceType != consts.WFolderSourceLegacyVFilm {
		return []types.Breadcrumb{}, err
	}
	parts := stringsSplitPath(cur.Path)
	var parentID int64
	var out []types.Breadcrumb
	for _, name := range parts {
		row, err := s.deps.WFolderModel.FindOneByParentIdNameSourceType(ctx, parentID, name, consts.WFolderSourceLegacyVFilm)
		if err != nil {
			return nil, err
		}
		out = append(out, types.Breadcrumb{Id: row.Id, Name: row.Name, Path: row.Path})
		parentID = row.Id
	}
	return out, nil
}

func (s *Service) directoryListSubtreeIDs(ctx context.Context, id int64) ([]int64, error) {
	d, err := s.deps.WFolderModel.FindOne(ctx, id)
	if err != nil {
		if err == moviex.ErrNotFound {
			return []int64{}, nil
		}
		return nil, err
	}
	if d == nil {
		return []int64{}, nil
	}
	if d.SourceType != consts.WFolderSourceLegacyVFilm {
		return []int64{}, nil
	}
	ids, err := s.deps.WFolderModel.ListSubtreeIDsByPathSourceType(ctx, d.Path, consts.WFolderSourceLegacyVFilm)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []int64{id}, nil
	}
	return ids, nil
}

func (s *Service) filmFindAll(ctx context.Context, removedStatus int64) ([]*types.Film, error) {
	rows, err := s.deps.WMediaModel.FindAllLegacyFilms(ctx, removedStatus)
	if err != nil {
		return nil, err
	}
	out := make([]*types.Film, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapFilmModelToTypes(row))
	}
	return out, nil
}

func (s *Service) filmFindOne(ctx context.Context, id int64) (*types.Film, error) {
	row, err := s.deps.WMediaModel.FindOneLegacyFilmByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapFilmModelToTypes(row), nil
}

func (s *Service) filmFindOneByMovieJavID(ctx context.Context, javID string) (*types.Film, error) {
	row, err := s.deps.WMediaModel.FindOneLegacyFilmByMovieJavId(ctx, javID)
	if err != nil {
		return nil, err
	}
	return mapFilmModelToTypes(row), nil
}

func (s *Service) filmUpsert(ctx context.Context, in *types.Film) (*types.Film, consts.UpsertStatus, error) {
	if in == nil {
		return nil, 0, nil
	}
	if row, err := s.deps.WMediaModel.FindOneByMovieJavIdSourceType(ctx, in.MovieJavId, consts.WMediaSourceLegacyVFilm); err == nil && row != nil {
		changed := false
		if row.MovieName != in.MovieName {
			row.MovieName = in.MovieName
			changed = true
		}
		if row.FileName != in.FileName {
			row.FileName = in.FileName
			changed = true
		}
		if row.DirectoryId != in.DirectoryId {
			row.DirectoryId = in.DirectoryId
			changed = true
		}
		if row.RootDir != in.RootDir {
			row.RootDir = in.RootDir
			changed = true
		}
		if row.FullDir != in.FullDir {
			row.FullDir = in.FullDir
			changed = true
		}
		if row.Alias != in.Alias {
			row.Alias = in.Alias
			changed = true
		}
		if row.Size != in.Size {
			row.Size = in.Size
			changed = true
		}
		if row.Width != in.Width {
			row.Width = in.Width
			changed = true
		}
		if row.Height != in.Height {
			row.Height = in.Height
			changed = true
		}
		if row.BitRate != in.BitRate {
			row.BitRate = in.BitRate
			changed = true
		}
		if row.Duration != in.Duration {
			row.Duration = in.Duration
			changed = true
		}
		if row.FrameAverage != in.FrameAverage {
			row.FrameAverage = in.FrameAverage
			changed = true
		}
		if row.HasSub != in.HasSub {
			row.HasSub = in.HasSub
			changed = true
		}
		if row.SelfMake != in.SelfMake {
			row.SelfMake = in.SelfMake
			changed = true
		}
		if row.HasMask != in.HasMask {
			row.HasMask = in.HasMask
			changed = true
		}
		if row.NeedScanMeta != in.NeedScanMeta {
			row.NeedScanMeta = in.NeedScanMeta
			changed = true
		}
		if row.IsRemoved != in.IsRemoved {
			row.IsRemoved = in.IsRemoved
			changed = true
		}
		if row.RemoveTime != in.RemoveTime {
			row.RemoveTime = in.RemoveTime
			changed = true
		}
		if row.BirthTime != in.BirthTime {
			row.BirthTime = in.BirthTime
			changed = true
		}
		if row.ReleasingDate != in.ReleasingDate {
			row.ReleasingDate = in.ReleasingDate
			changed = true
		}
		if changed {
			row.UpdatedOn = time.Now().Unix()
			if err := s.deps.WMediaModel.Update(ctx, row); err != nil {
				return nil, 0, err
			}
			updated, err := s.filmFindOneByMovieJavID(ctx, in.MovieJavId)
			if err != nil {
				return nil, 0, err
			}
			return updated, consts.UpsertUpdated, nil
		}
		current, err := s.filmFindOneByMovieJavID(ctx, in.MovieJavId)
		if err != nil {
			return nil, 0, err
		}
		return current, consts.UpsertUnchanged, nil
	}

	row := mapFilmTypesToModel(in)
	now := time.Now().Unix()
	row.SourceType = consts.WMediaSourceLegacyVFilm
	row.SourceTorrentHash = ""
	row.CreatedOn = now
	row.UpdatedOn = now
	if _, err := s.deps.WMediaModel.Insert(ctx, row); err != nil {
		if again, err2 := s.filmFindOneByMovieJavID(ctx, in.MovieJavId); err2 == nil && again != nil {
			return again, consts.UpsertUpdated, nil
		}
		return nil, 0, err
	}
	ins, err := s.filmFindOneByMovieJavID(ctx, in.MovieJavId)
	if err != nil {
		return nil, 0, err
	}
	return ins, consts.UpsertInserted, nil
}

func (s *Service) filmListByDirectories(ctx context.Context, dirIDs []int64, page, size int, orderBy string) (all, paged []*types.Film, total int64, err error) {
	if len(dirIDs) == 0 {
		return []*types.Film{}, []*types.Film{}, 0, nil
	}
	rowsAll, rowsPaged, total, err := s.deps.WMediaModel.ListLegacyFilmsByDirectoryIDs(ctx, dirIDs, page, size, mapFilmOrderBy(orderBy))
	if err != nil {
		return nil, nil, 0, err
	}
	all = make([]*types.Film, 0, len(rowsAll))
	for _, row := range rowsAll {
		all = append(all, mapFilmModelToTypes(row))
	}
	paged = make([]*types.Film, 0, len(rowsPaged))
	for _, row := range rowsPaged {
		paged = append(paged, mapFilmModelToTypes(row))
	}
	return all, paged, total, nil
}

func (s *Service) movieFindByName(ctx context.Context, name string) ([]*types.Movie, error) {
	rows, err := s.deps.MovieModel.FindMoviesByName(ctx, name)
	if err != nil {
		return nil, err
	}
	out := make([]*types.Movie, 0, len(rows))
	for _, row := range rows {
		out = append(out, &types.Movie{
			Id:                   row.Id,
			Name:                 row.Name,
			JavId:                row.JavId,
			Title:                row.Title,
			EncodeName:           row.EncodeName,
			ReleasingDate:        row.ReleasingDate,
			Length:               row.Length,
			Score:                row.Score,
			ViewersNumberWant:    row.ViewersNumberWant,
			ViewersNumberOwned:   row.ViewersNumberOwned,
			ViewersNumberWatched: row.ViewersNumberWatched,
			PrefixId:             row.PrefixId,
			MakerId:              row.MakerId,
			LabelId:              row.LabelId,
			DirectorId:           row.DirectorId,
			CastNumber:           row.CastNumber,
			CastAverageAge:       row.CastAverageAge,
			DetailUpdateTime:     row.DetailUpdateTime,
			CreatedOn:            row.CreatedOn,
			UpdatedOn:            row.UpdatedOn,
		})
	}
	return out, nil
}

func mapDirectoryModelToTypes(v *moviex.WFolder) *types.Directory {
	if v == nil {
		return nil
	}
	return &types.Directory{
		Id:        v.Id,
		ParentId:  v.ParentId,
		Name:      v.Name,
		Depth:     v.Depth,
		Path:      v.Path,
		CreatedOn: v.CreatedOn,
		UpdatedOn: v.UpdatedOn,
	}
}

func mapFilmModelToTypes(v *moviex.LegacyFilm) *types.Film {
	if v == nil {
		return nil
	}
	return &types.Film{
		Id:            v.Id,
		MovieJavId:    v.MovieJavId,
		MovieName:     v.MovieName,
		FileName:      v.FileName,
		DirectoryId:   v.DirectoryId,
		RootDir:       v.RootDir,
		FullDir:       v.FullDir,
		Dir1Id:        v.Dir1Id,
		Dir2Id:        v.Dir2Id,
		Dir3Id:        v.Dir3Id,
		Dir4Id:        v.Dir4Id,
		Alias:         v.Alias,
		Size:          v.Size,
		Width:         v.Width,
		Height:        v.Height,
		BitRate:       v.BitRate,
		Duration:      v.Duration,
		FrameAverage:  v.FrameAverage,
		HasSub:        v.HasSub,
		SelfMake:      v.SelfMake,
		HasMask:       v.HasMask,
		NeedScanMeta:  v.NeedScanMeta,
		IsRemoved:     v.IsRemoved,
		RemoveTime:    v.RemoveTime,
		ScTimes:       v.ScTimes,
		ComeTimes:     v.ComeTimes,
		LastScTime:    v.LastScTime,
		BirthTime:     v.BirthTime,
		ReleasingDate: v.ReleasingDate,
		CreatedOn:     v.CreatedOn,
		UpdatedOn:     v.UpdatedOn,
	}
}

func mapFilmTypesToModel(in *types.Film) *moviex.WMedia {
	return &moviex.WMedia{
		Id:            in.Id,
		MovieJavId:    in.MovieJavId,
		MovieName:     in.MovieName,
		FileName:      in.FileName,
		DirectoryId:   in.DirectoryId,
		RootDir:       in.RootDir,
		FullDir:       in.FullDir,
		Alias:         in.Alias,
		Size:          in.Size,
		Width:         in.Width,
		Height:        in.Height,
		BitRate:       in.BitRate,
		Duration:      in.Duration,
		FrameAverage:  in.FrameAverage,
		HasSub:        in.HasSub,
		SelfMake:      in.SelfMake,
		HasMask:       in.HasMask,
		NeedScanMeta:  in.NeedScanMeta,
		IsRemoved:     in.IsRemoved,
		RemoveTime:    in.RemoveTime,
		BirthTime:     in.BirthTime,
		ReleasingDate: in.ReleasingDate,
		CreatedOn:     in.CreatedOn,
		UpdatedOn:     in.UpdatedOn,
	}
}

func mapFilmOrderBy(orderBy string) string {
	order := "f.birth_time DESC"
	switch orderBy {
	case consts.OrderByBirthTime:
		order = "f.birth_time DESC,f.movie_name DESC"
	case consts.OrderByScTimes:
		order = "COALESCE(gss.sc_times, 0) DESC,COALESCE(gss.last_sc_time, 0) DESC,f.movie_name DESC"
	case consts.OrderByComeTimes:
		order = "COALESCE(gss.come_times, 0) DESC,COALESCE(gss.last_sc_time, 0) DESC,f.movie_name DESC"
	case consts.OrderByLastScTime:
		order = "COALESCE(gss.last_sc_time, 0) DESC,f.movie_name DESC"
	case consts.OrderByReleasingDate:
		order = "f.releasing_date DESC,f.movie_name DESC"
	}
	return order
}

func stringsSplitPath(p string) []string {
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}
