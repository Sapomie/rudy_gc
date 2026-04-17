package sc

import (
	"context"
	"errors"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

func (s *Service) scFindAll(ctx context.Context) ([]*types.GSc, error) {
	rows, err := s.deps.ScModel.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*types.GSc, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapScModelToTypes(row))
	}
	return out, nil
}

func (s *Service) scFindByNames(ctx context.Context, names []string) ([]*types.GSc, error) {
	rows, err := s.deps.ScModel.ListByNames(ctx, names)
	if err != nil {
		return nil, err
	}
	out := make([]*types.GSc, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapScModelToTypes(row))
	}
	return out, nil
}

func (s *Service) scFindNearest(ctx context.Context, t int64) (*types.GSc, error) {
	row, err := s.deps.ScModel.FindNearest(ctx, t)
	if err != nil {
		return nil, err
	}
	return mapScModelToTypes(row), nil
}

func (s *Service) scFindOneByName(ctx context.Context, name string) (*types.GSc, error) {
	row, err := s.deps.ScModel.FindOneByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return mapScModelToTypes(row), nil
}

func (s *Service) scListPage(ctx context.Context, page, pageSize int, sortField, sortOrder string) ([]*types.GSc, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}

	total, err := s.deps.ScModel.CountAll(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*types.GSc{}, 0, nil
	}

	offset := int64((page - 1) * pageSize)
	rows, err := s.deps.ScModel.ListPage(ctx, offset, int64(pageSize), buildScOrderBy(sortField, sortOrder))
	if err != nil {
		return nil, 0, err
	}

	out := make([]*types.GSc, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapScModelToTypes(row))
	}
	return out, total, nil
}

func (s *Service) scUpsert(ctx context.Context, in *types.GSc) (*types.GSc, error) {
	if in == nil {
		return nil, errors.New("nil input")
	}
	now := time.Now().Unix()

	old, err := s.deps.ScModel.FindOneByName(ctx, in.Name)
	if err == nil && old != nil {
		changed := false
		if old.MovieNumber != in.MovieNumber {
			old.MovieNumber = in.MovieNumber
			changed = true
		}
		if old.ScTime != in.ScTime {
			old.ScTime = in.ScTime
			changed = true
		}
		if old.ComeMovieName != in.ComeMovieName {
			old.ComeMovieName = in.ComeMovieName
			changed = true
		}
		if old.Cooldown != in.Cooldown {
			old.Cooldown = in.Cooldown
			changed = true
		}
		if old.Duration != in.DurationMinutes {
			old.Duration = in.DurationMinutes
			changed = true
		}
		if old.Fg != in.Fg {
			old.Fg = in.Fg
			changed = true
		}
		if old.Vessel != in.Vessel {
			old.Vessel = in.Vessel
			changed = true
		}
		if old.MovieCast != in.MovieCast {
			old.MovieCast = in.MovieCast
			changed = true
		}
		if old.Remarks != in.Remarks {
			old.Remarks = in.Remarks
			changed = true
		}
		if old.ImagePath != in.ImagePath {
			old.ImagePath = in.ImagePath
			changed = true
		}
		if changed {
			old.UpdatedOn = now
			if err := s.deps.ScModel.Update(ctx, old); err != nil {
				return nil, err
			}
			if err := s.refreshPersonScSnapshotsByScNames(ctx, now, in.Name); err != nil {
				return nil, err
			}
		}
		return mapScModelToTypes(old), nil
	}

	row := &moviex.GSc{
		Name:          in.Name,
		MovieNumber:   in.MovieNumber,
		ScTime:        in.ScTime,
		ComeMovieName: in.ComeMovieName,
		Cooldown:      in.Cooldown,
		Duration:      in.DurationMinutes,
		Fg:            in.Fg,
		Vessel:        in.Vessel,
		MovieCast:     in.MovieCast,
		Remarks:       in.Remarks,
		ImagePath:     in.ImagePath,
		CreatedOn:     now,
		UpdatedOn:     now,
	}
	if _, err := s.deps.ScModel.Insert(ctx, row); err != nil {
		if again, e2 := s.deps.ScModel.FindOneByName(ctx, in.Name); e2 == nil && again != nil {
			return mapScModelToTypes(again), nil
		}
		return nil, err
	}
	ins, err := s.deps.ScModel.FindOneByName(ctx, in.Name)
	if err != nil {
		return nil, err
	}
	if err := s.refreshPersonScSnapshotsByScNames(ctx, now, in.Name); err != nil {
		return nil, err
	}
	return mapScModelToTypes(ins), nil
}

func (s *Service) glFindAll(ctx context.Context) ([]*types.GList, error) {
	rows, err := s.deps.GListModel.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*types.GList, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapGListModelToTypes(row))
	}
	return out, nil
}

func (s *Service) glUpsert(ctx context.Context, in *types.GList) (*types.GList, error) {
	if in == nil {
		return nil, errors.New("nil input")
	}
	now := time.Now().Unix()

	old, err := s.deps.GListModel.FindOneByName(ctx, in.Name)
	if err == nil && old != nil {
		oldMovieJavID := old.MovieJavId
		changed := false
		if old.ScName != in.ScName {
			old.ScName = in.ScName
			changed = true
		}
		if old.MovieJavId != in.MovieJavId {
			old.MovieJavId = in.MovieJavId
			changed = true
		}
		if old.IsCome != in.IsCome {
			old.IsCome = in.IsCome
			changed = true
		}
		if changed {
			old.UpdatedOn = now
			if err := s.deps.GListModel.Update(ctx, old); err != nil {
				return nil, err
			}
			if err := s.refreshPersonScSnapshotsByMovieJavIDs(ctx, now, oldMovieJavID, in.MovieJavId); err != nil {
				return nil, err
			}
		}
		return mapGListModelToTypes(old), nil
	}

	row := &moviex.GList{
		Name:       in.Name,
		ScName:     in.ScName,
		MovieJavId: in.MovieJavId,
		IsCome:     in.IsCome,
		CreatedOn:  now,
		UpdatedOn:  now,
	}
	if _, err := s.deps.GListModel.Insert(ctx, row); err != nil {
		if again, e2 := s.deps.GListModel.FindOneByName(ctx, in.Name); e2 == nil && again != nil {
			return mapGListModelToTypes(again), nil
		}
		return nil, err
	}
	ins, err := s.deps.GListModel.FindOneByName(ctx, in.Name)
	if err != nil {
		return nil, err
	}
	if err := s.refreshPersonScSnapshotsByMovieJavIDs(ctx, now, in.MovieJavId); err != nil {
		return nil, err
	}
	return mapGListModelToTypes(ins), nil
}

func (s *Service) glFindByScName(ctx context.Context, scName string) ([]*types.GList, error) {
	rows, err := s.deps.GListModel.ListByScName(ctx, scName)
	if err != nil {
		return nil, err
	}
	out := make([]*types.GList, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapGListModelToTypes(row))
	}
	return out, nil
}

func (s *Service) glFindByFilters(ctx context.Context, scName string, isCome *int64, page, pageSize int) ([]*types.GList, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 20
	}
	offset := int64((page - 1) * pageSize)
	rows, err := s.deps.GListModel.ListByFilters(ctx, scName, isCome, offset, int64(pageSize))
	if err != nil {
		return nil, err
	}
	out := make([]*types.GList, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapGListModelToTypes(row))
	}
	return out, nil
}

func (s *Service) glFindByMovieJavIDs(ctx context.Context, javIDs []string) ([]*types.GList, error) {
	rows, err := s.deps.GListModel.ListByMovieJavIds(ctx, javIDs)
	if err != nil {
		return nil, err
	}
	out := make([]*types.GList, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapGListModelToTypes(row))
	}
	return out, nil
}

func (s *Service) glFindByMovieJavID(ctx context.Context, javID string) ([]*types.GList, error) {
	rows, err := s.deps.GListModel.ListByMovieJavId(ctx, javID)
	if err != nil {
		return nil, err
	}
	out := make([]*types.GList, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapGListModelToTypes(row))
	}
	return out, nil
}

func (s *Service) glListDistinctMovieJavIDs(ctx context.Context) ([]string, error) {
	return s.deps.GListModel.ListDistinctMovieJavIds(ctx)
}

func (s *Service) movieFindOneByJavID(ctx context.Context, javID string) (*moviex.AMovie, error) {
	row, err := s.deps.MovieModel.FindOneByJavId(ctx, javID)
	if err != nil {
		return nil, err
	}
	return row, nil
}

func (s *Service) gScStatFindOneByMovieJavID(ctx context.Context, javID string) (*moviex.GScStat, error) {
	row, err := s.deps.GScStatModel.FindOneByMovieJavId(ctx, javID)
	if err != nil {
		return nil, err
	}
	return row, nil
}

func (s *Service) wMediaFindOneByMovieJavID(ctx context.Context, javID string) (*moviex.WMedia, error) {
	row, err := s.deps.WMediaModel.FindOneByMovieJavId(ctx, javID)
	if err != nil {
		return nil, err
	}
	return row, nil
}

func (s *Service) wMediaFindOneByMovieName(ctx context.Context, name string) (*moviex.WMedia, error) {
	row, err := s.deps.WMediaModel.FindOneByMovieName(ctx, name)
	if err != nil {
		return nil, err
	}
	return row, nil
}

func (s *Service) gScStatUpsert(ctx context.Context, movieJavID, movieName string, releasingDate, mediaBirthTime int64, info movieScInfo) (*moviex.GScStat, consts.UpsertStatus, error) {
	movieJavID = strings.TrimSpace(movieJavID)
	if movieJavID == "" {
		return nil, 0, errors.New("empty movie jav id")
	}
	now := time.Now().Unix()

	old, err := s.gScStatFindOneByMovieJavID(ctx, movieJavID)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, 0, err
	}
	if old != nil {
		changed := false
		if old.MovieName != movieName {
			old.MovieName = movieName
			changed = true
		}
		if old.ScTimes != info.ScTimes {
			old.ScTimes = info.ScTimes
			changed = true
		}
		if old.ComeTimes != info.ComeTimes {
			old.ComeTimes = info.ComeTimes
			changed = true
		}
		if old.LastScTime != info.LastScTime {
			old.LastScTime = info.LastScTime
			changed = true
		}
		if old.ReleasingDate != releasingDate {
			old.ReleasingDate = releasingDate
			changed = true
		}
		if old.MediaBirthTime != mediaBirthTime {
			old.MediaBirthTime = mediaBirthTime
			changed = true
		}
		if changed {
			old.UpdatedOn = now
			if err := s.deps.GScStatModel.Update(ctx, old); err != nil {
				return nil, 0, err
			}
			return old, consts.UpsertUpdated, nil
		}
		return old, consts.UpsertUnchanged, nil
	}

	row := &moviex.GScStat{
		MovieJavId:     movieJavID,
		MovieName:      movieName,
		ScTimes:        info.ScTimes,
		ComeTimes:      info.ComeTimes,
		LastScTime:     info.LastScTime,
		ReleasingDate:  releasingDate,
		MediaBirthTime: mediaBirthTime,
		CreatedOn:      now,
		UpdatedOn:      now,
	}
	if _, err := s.deps.GScStatModel.Insert(ctx, row); err != nil {
		if again, e2 := s.gScStatFindOneByMovieJavID(ctx, movieJavID); e2 == nil && again != nil {
			return again, consts.UpsertUpdated, nil
		}
		return nil, 0, err
	}
	ins, err := s.gScStatFindOneByMovieJavID(ctx, movieJavID)
	if err != nil {
		return nil, 0, err
	}
	return ins, consts.UpsertInserted, nil
}

func (s *Service) castFindOne(ctx context.Context, id int64) (*types.Cast, error) {
	row, err := s.deps.CastModel.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapCastModelToTypes(row), nil
}

func (s *Service) castFindOneByName(ctx context.Context, name string) (*types.Cast, error) {
	row, err := s.deps.CastModel.FindOneByName(ctx, name)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	return mapCastModelToTypes(row), nil
}

func (s *Service) castFindByNames(ctx context.Context, names []string) ([]*types.Cast, error) {
	return s.deps.CastModel.FindByNames(ctx, names)
}

func (s *Service) castCountOwnedScMovieNumbersByNames(ctx context.Context, names []string) (map[string]int64, error) {
	return s.deps.CastModel.CountOwnedScMovieNumbersByNames(ctx, names)
}

func (s *Service) personFindOne(ctx context.Context, id int64) (*types.Person, error) {
	row, err := s.deps.PersonModel.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	return mapPersonModelToTypes(row), nil
}

func (s *Service) personFindByIDs(ctx context.Context, ids []int64) ([]*types.Person, error) {
	return s.deps.PersonModel.FindByIDs(ctx, ids)
}

func (s *Service) personCountOwnedScMovieNumbersByIDs(ctx context.Context, ids []int64) (map[int64]int64, error) {
	return s.deps.PersonModel.CountOwnedScMovieNumbersByIDs(ctx, ids)
}

func (s *Service) castUpsert(ctx context.Context, in *types.Cast) (*types.Cast, error) {
	if in == nil {
		return nil, errors.New("nil input")
	}
	now := time.Now().Unix()
	old, err := s.deps.CastModel.FindOneByName(ctx, in.Name)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, err
	}
	if old != nil {
		if _, err := s.ensureCastPersonID(ctx, old, now); err != nil {
			return nil, err
		}
		old.JavId = in.JavId
		old.MovieNumber = in.MovieNumber
		old.OwnedMovieNumber = in.OwnedWMediaNumber
		old.OwnedWMediaNumber = in.OwnedWMediaNumber
		old.ScTimes = in.ScTimes
		old.ComeTimes = in.ComeTimes
		old.LastScTime = in.LastScTime
		old.Rank500MovieNumber = in.Rank500MovieNumber
		old.Rank20MovieNumber = in.Rank20MovieNumber
		old.Rank1MovieNumber = in.Rank1MovieNumber
		old.HighestRank = in.HighestRank
		old.RankTimes = in.RankTimes
		if in.CreatedOn > 0 {
			old.CreatedOn = in.CreatedOn
		}
		old.UpdatedOn = now
		if err := s.deps.CastModel.Update(ctx, old); err != nil {
			return nil, err
		}
		if err := s.syncPersonStatsByIDs(ctx, now, old.PersonId); err != nil {
			return nil, err
		}
		return mapCastModelToTypes(old), nil
	}

	row := &moviex.AmCast{
		PersonId:           in.PersonId,
		Name:               in.Name,
		JavId:              in.JavId,
		MovieNumber:        in.MovieNumber,
		OwnedMovieNumber:   in.OwnedWMediaNumber,
		OwnedWMediaNumber:  in.OwnedWMediaNumber,
		ScTimes:            in.ScTimes,
		ComeTimes:          in.ComeTimes,
		LastScTime:         in.LastScTime,
		Rank500MovieNumber: in.Rank500MovieNumber,
		Rank20MovieNumber:  in.Rank20MovieNumber,
		Rank1MovieNumber:   in.Rank1MovieNumber,
		HighestRank:        in.HighestRank,
		RankTimes:          in.RankTimes,
		CreatedOn:          ifElseInt64(in.CreatedOn > 0, in.CreatedOn, now),
		UpdatedOn:          now,
	}
	if row.PersonId <= 0 {
		personID, err := s.insertPersonForCast(ctx, in, now)
		if err != nil {
			return nil, err
		}
		row.PersonId = personID
	}
	if _, err := s.deps.CastModel.Insert(ctx, row); err != nil {
		if again, e2 := s.deps.CastModel.FindOneByName(ctx, in.Name); e2 == nil && again != nil {
			if _, err := s.ensureCastPersonID(ctx, again, now); err != nil {
				return nil, err
			}
			return mapCastModelToTypes(again), nil
		}
		return nil, err
	}
	ins, err := s.deps.CastModel.FindOneByName(ctx, in.Name)
	if err != nil {
		return nil, err
	}
	if err := s.syncPersonStatsByIDs(ctx, now, row.PersonId); err != nil {
		return nil, err
	}
	return mapCastModelToTypes(ins), nil
}

func (s *Service) castUpdateMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64, now int64) error {
	movieNumber, ownedMovieNumber, ownedWMediaNumber, err := s.deps.CastModel.GetMovieNumbersWithWMediaByID(ctx, id, ownedRemovedStatus)
	if err != nil {
		return err
	}
	row, err := s.deps.CastModel.FindOne(ctx, id)
	if err != nil {
		return err
	}
	if row.MovieNumber == movieNumber && row.OwnedMovieNumber == ownedMovieNumber && row.OwnedWMediaNumber == ownedWMediaNumber {
		return nil
	}
	row.MovieNumber = movieNumber
	row.OwnedMovieNumber = ownedWMediaNumber
	row.OwnedWMediaNumber = ownedWMediaNumber
	row.UpdatedOn = now
	if err := s.deps.CastModel.Update(ctx, row); err != nil {
		return err
	}
	return s.syncPersonStatsByIDs(ctx, now, row.PersonId)
}

func (s *Service) syncPersonStatsByIDs(ctx context.Context, now int64, ids ...int64) error {
	if s == nil || s.deps == nil {
		return nil
	}
	filtered := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return s.deps.SyncPersonStatsByIDs(ctx, filtered, now)
}

func (s *Service) refreshPersonScSnapshotsByMovieJavIDs(ctx context.Context, now int64, movieJavIDs ...string) error {
	if s == nil || s.deps == nil || s.deps.CPersonScModel == nil {
		return nil
	}
	return s.deps.CPersonScModel.RebuildByMovieJavIDs(ctx, movieJavIDs, now)
}

func (s *Service) refreshPersonScSnapshotsByScNames(ctx context.Context, now int64, scNames ...string) error {
	if s == nil || s.deps == nil || s.deps.CPersonScModel == nil {
		return nil
	}
	return s.deps.CPersonScModel.RebuildByScNames(ctx, scNames, now)
}

func (s *Service) castListAllIDs(ctx context.Context) ([]int64, error) {
	return s.deps.CastModel.ListAllIDs(ctx)
}

func (s *Service) movieCastListCastIDsByMovieJavID(ctx context.Context, movieJavID string) ([]int64, error) {
	return s.deps.MovieCastModel.ListCastIDsByMovieJavId(ctx, movieJavID)
}

func (s *Service) movieCastListMovieJavIDsByCastID(ctx context.Context, castID int64) ([]string, error) {
	return s.deps.MovieCastModel.ListMovieJavIDsByCastID(ctx, castID)
}

func mapScModelToTypes(v *moviex.GSc) *types.GSc {
	if v == nil {
		return nil
	}
	return &types.GSc{
		Id:              v.Id,
		Name:            v.Name,
		MovieNumber:     v.MovieNumber,
		ScTime:          v.ScTime,
		ComeMovieName:   v.ComeMovieName,
		Cooldown:        v.Cooldown,
		DurationMinutes: v.Duration,
		Fg:              v.Fg,
		Vessel:          v.Vessel,
		MovieCast:       v.MovieCast,
		Remarks:         v.Remarks,
		ImagePath:       v.ImagePath,
		CreatedOn:       v.CreatedOn,
		UpdatedOn:       v.UpdatedOn,
	}
}

func mapGListModelToTypes(v *moviex.GList) *types.GList {
	if v == nil {
		return nil
	}
	return &types.GList{
		Id:         v.Id,
		Name:       v.Name,
		ScName:     v.ScName,
		MovieJavId: v.MovieJavId,
		IsCome:     v.IsCome,
		CreatedOn:  v.CreatedOn,
		UpdatedOn:  v.UpdatedOn,
	}
}

func mapCastModelToTypes(v *moviex.AmCast) *types.Cast {
	if v == nil {
		return nil
	}
	return &types.Cast{
		Id:                 v.Id,
		PersonId:           v.PersonId,
		Name:               v.Name,
		JavId:              v.JavId,
		Chinese:            "",
		BirthDay:           0,
		Height:             0,
		MovieNumber:        v.MovieNumber,
		OwnedMovieNumber:   v.OwnedWMediaNumber,
		OwnedWMediaNumber:  v.OwnedWMediaNumber,
		ScTimes:            v.ScTimes,
		ComeTimes:          v.ComeTimes,
		LastScTime:         v.LastScTime,
		Rank500MovieNumber: v.Rank500MovieNumber,
		Rank20MovieNumber:  v.Rank20MovieNumber,
		Rank1MovieNumber:   v.Rank1MovieNumber,
		HighestRank:        v.HighestRank,
		RankTimes:          v.RankTimes,
		CreatedOn:          v.CreatedOn,
		UpdatedOn:          v.UpdatedOn,
	}
}

func mapPersonModelToTypes(v *moviex.CPerson) *types.Person {
	if v == nil {
		return nil
	}
	return &types.Person{
		Id:                v.Id,
		Name:              v.Name,
		Alias:             v.Alias,
		Chinese:           v.Chinese,
		BirthDay:          v.BirthDay,
		Height:            v.Height,
		Cup:               v.Cup,
		Bwh:               v.Bwh,
		Avatar:            v.Avatar,
		MovieNumber:       v.MovieNumber,
		OwnedMovieNumber:  v.OwnedWMediaNumber,
		OwnedWMediaNumber: v.OwnedWMediaNumber,
		ScTimes:           v.ScTimes,
		ComeTimes:         v.ComeTimes,
		LastScTime:        v.LastScTime,
		HighestRank:       v.HighestRank,
		RankTimes:         v.RankTimes,
		CreatedOn:         v.CreatedOn,
		UpdatedOn:         v.UpdatedOn,
	}
}

func (s *Service) insertPersonForCast(ctx context.Context, cast *types.Cast, now int64) (int64, error) {
	if cast == nil {
		return 0, errors.New("nil cast")
	}
	name := strings.TrimSpace(cast.Name)
	if name == "" {
		return 0, errors.New("empty cast name")
	}
	personRow, err := s.deps.PersonModel.FindOneByNameOrAliasToken(ctx, name)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return 0, err
	}
	if personRow != nil && personRow.Id > 0 {
		return personRow.Id, nil
	}
	row := &moviex.CPerson{
		Name:              name,
		Alias:             "",
		Chinese:           "",
		BirthDay:          0,
		Height:            0,
		Cup:               "",
		Bwh:               "",
		Avatar:            "",
		MovieNumber:       cast.MovieNumber,
		OwnedMovieNumber:  cast.OwnedWMediaNumber,
		OwnedWMediaNumber: cast.OwnedWMediaNumber,
		ScTimes:           cast.ScTimes,
		ComeTimes:         cast.ComeTimes,
		LastScTime:        cast.LastScTime,
		HighestRank:       cast.HighestRank,
		RankTimes:         cast.RankTimes,
		CreatedOn:         ifElseInt64(cast.CreatedOn > 0, cast.CreatedOn, now),
		UpdatedOn:         now,
	}
	res, err := s.deps.PersonModel.Insert(ctx, row)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Service) ensureCastPersonID(ctx context.Context, row *moviex.AmCast, now int64) (int64, error) {
	if row == nil {
		return 0, nil
	}
	if row.PersonId > 0 {
		if _, err := s.deps.PersonModel.FindOne(ctx, row.PersonId); err == nil {
			return row.PersonId, nil
		} else if !errors.Is(err, moviex.ErrNotFound) {
			return 0, err
		}
	}
	personID, err := s.insertPersonForCast(ctx, mapCastModelToTypes(row), now)
	if err != nil {
		return 0, err
	}
	row.PersonId = personID
	row.UpdatedOn = now
	if err := s.deps.CastModel.Update(ctx, row); err != nil {
		return 0, err
	}
	return personID, nil
}

func ifElseInt64(cond bool, a, b int64) int64 {
	if cond {
		return a
	}
	return b
}

func buildScOrderBy(sortField, sortOrder string) string {
	field := normalizeScSortField(sortField)
	order := normalizeScSortOrder(sortOrder)

	column := "sc_time"
	switch field {
	case "movie_number":
		column = "movie_number"
	case "come_movie_name":
		column = "come_movie_name"
	case "cooldown":
		column = "cooldown"
	case "duration":
		column = "duration"
	case "movie_cast":
		column = "movie_cast"
	case "vessel":
		column = "vessel"
	case "fg":
		column = "fg"
	default:
		column = "sc_time"
	}

	if column == "sc_time" {
		return column + " " + order + ", id DESC"
	}
	return column + " " + order + ", sc_time DESC, id DESC"
}

func normalizeScSortField(sortField string) string {
	switch strings.ToLower(strings.TrimSpace(sortField)) {
	case "movie_number", "come_movie_name", "cooldown", "duration", "movie_cast", "vessel", "fg", "sc_time":
		return strings.ToLower(strings.TrimSpace(sortField))
	default:
		return "sc_time"
	}
}

func normalizeScSortOrder(sortOrder string) string {
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		return "ASC"
	}
	return "DESC"
}
