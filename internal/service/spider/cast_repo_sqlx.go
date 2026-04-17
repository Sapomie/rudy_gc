package spider

import (
	"context"
	"errors"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

type CastRepoSqlx struct {
	m               moviex.AmCastModel
	pm              moviex.CPersonModel
	syncPersonStats func(ctx context.Context, ids []int64, now int64) error
}

// ====== 已有：保持不变（返回 id）
func (r *CastRepoSqlx) GetOrCreateByName(ctx context.Context, name, javId string) (int64, error) {
	if name == "" {
		return 0, nil
	}
	row, err := r.m.FindOneByName(ctx, name)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return 0, err
	}
	if row != nil {
		now := time.Now().Unix()
		if _, err := r.ensurePersonID(ctx, row, now); err != nil {
			return 0, err
		}
		return row.Id, nil
	}
	now := time.Now().Unix()
	personID, err := r.insertPerson(ctx, name, nil, now)
	if err != nil {
		return 0, err
	}
	res, err := r.m.Insert(ctx, &moviex.AmCast{
		PersonId:  personID,
		Name:      name,
		JavId:     javId,
		CreatedOn: now,
		UpdatedOn: now,
	})
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

// ====== 新版：FindOne / FindOneByName 返回 types.Cast
func (r *CastRepoSqlx) FindOne(ctx context.Context, id int64) (*types.Cast, error) {
	row, err := r.m.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapAmCastToTypes(row), nil
}

func (r *CastRepoSqlx) FindOneByName(ctx context.Context, name string) (*types.Cast, error) {
	row, err := r.m.FindOneByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return mapAmCastToTypes(row), nil
}

func (r *CastRepoSqlx) FindByNames(ctx context.Context, names []string) ([]*types.Cast, error) {
	return r.m.FindByNames(ctx, names)
}

func (r *CastRepoSqlx) CountOwnedScMovieNumbersByNames(ctx context.Context, names []string) (map[string]int64, error) {
	return r.m.CountOwnedScMovieNumbersByNames(ctx, names)
}

// ====== 新增：Upsert（以 name 作为幂等键）
func (r *CastRepoSqlx) Upsert(ctx context.Context, in *types.Cast) (*types.Cast, error) {
	if in == nil {
		return nil, errors.New("nil input")
	}
	now := time.Now().Unix()

	// 以 name 为幂等键
	old, err := r.m.FindOneByName(ctx, in.Name)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, err
	}

	if old != nil {
		if _, err := r.ensurePersonID(ctx, old, now); err != nil {
			return nil, err
		}
		// 覆盖式更新（按你的字段一一映射）
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

		if err := r.m.Update(ctx, old); err != nil {
			return nil, err
		}
		if err := r.syncPersonStatsIfNeeded(ctx, now, old.PersonId); err != nil {
			return nil, err
		}
		return mapAmCastToTypes(old), nil
	}

	// 插入
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
		personID, err := r.insertPerson(ctx, in.Name, in, now)
		if err != nil {
			return nil, err
		}
		row.PersonId = personID
	}
	if _, err := r.m.Insert(ctx, row); err != nil {
		// 并发兜底
		if again, e2 := r.m.FindOneByName(ctx, in.Name); e2 == nil && again != nil {
			if _, err := r.ensurePersonID(ctx, again, now); err != nil {
				return nil, err
			}
			return mapAmCastToTypes(again), nil
		}
		return nil, err
	}
	ins, err := r.m.FindOneByName(ctx, in.Name)
	if err != nil {
		return nil, err
	}
	if err := r.syncPersonStatsIfNeeded(ctx, now, row.PersonId); err != nil {
		return nil, err
	}
	return mapAmCastToTypes(ins), nil
}

/******** helpers ********/

func mapAmCastToTypes(v *moviex.AmCast) *types.Cast {
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

func (r *CastRepoSqlx) ensurePersonID(ctx context.Context, row *moviex.AmCast, now int64) (int64, error) {
	if row == nil {
		return 0, nil
	}
	if row.PersonId > 0 {
		if _, err := r.pm.FindOne(ctx, row.PersonId); err == nil {
			return row.PersonId, nil
		} else if !errors.Is(err, moviex.ErrNotFound) {
			return 0, err
		}
	}
	personID, err := r.insertPerson(ctx, row.Name, mapAmCastToTypes(row), now)
	if err != nil {
		return 0, err
	}
	row.PersonId = personID
	row.UpdatedOn = now
	if err := r.m.Update(ctx, row); err != nil {
		return 0, err
	}
	if err := r.syncPersonStatsIfNeeded(ctx, now, personID); err != nil {
		return 0, err
	}
	return personID, nil
}

func (r *CastRepoSqlx) insertPerson(ctx context.Context, name string, cast *types.Cast, now int64) (int64, error) {
	if r.pm == nil {
		return 0, errors.New("person model is nil")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("person name is empty")
	}
	personRow, err := r.pm.FindOneByNameOrAliasToken(ctx, name)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return 0, err
	}
	if personRow != nil && personRow.Id > 0 {
		return personRow.Id, nil
	}
	row := &moviex.CPerson{
		Name:      name,
		Alias:     "",
		Chinese:   "",
		BirthDay:  0,
		Height:    0,
		Cup:       "",
		Bwh:       "",
		Avatar:    "",
		CreatedOn: now,
		UpdatedOn: now,
	}
	if cast != nil {
		row.MovieNumber = cast.MovieNumber
		row.OwnedMovieNumber = cast.OwnedWMediaNumber
		row.OwnedWMediaNumber = cast.OwnedWMediaNumber
		row.ScTimes = cast.ScTimes
		row.ComeTimes = cast.ComeTimes
		row.LastScTime = cast.LastScTime
		row.HighestRank = cast.HighestRank
		row.RankTimes = cast.RankTimes
	}
	res, err := r.pm.Insert(ctx, row)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ifElseInt64(cond bool, a, b int64) int64 {
	if cond {
		return a
	}
	return b
}

func (r *CastRepoSqlx) UpdateMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64, now int64) error {
	movieNumber, ownedMovieNumber, ownedWMediaNumber, err := r.m.GetMovieNumbersWithWMediaByID(ctx, id, ownedRemovedStatus)
	if err != nil {
		return err
	}

	row, err := r.m.FindOne(ctx, id)
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

	if err := r.m.Update(ctx, row); err != nil {
		return err
	}
	return r.syncPersonStatsIfNeeded(ctx, now, row.PersonId)
}

func (r *CastRepoSqlx) syncPersonStatsIfNeeded(ctx context.Context, now int64, personIDs ...int64) error {
	if r == nil || r.syncPersonStats == nil {
		return nil
	}
	ids := make([]int64, 0, len(personIDs))
	for _, id := range personIDs {
		if id > 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return r.syncPersonStats(ctx, ids, now)
}

func (r *CastRepoSqlx) ListAllIDs(ctx context.Context) ([]int64, error) {
	return r.m.ListAllIDs(ctx)
}

func (r *CastRepoSqlx) ListPage(ctx context.Context, page, pageSize int, sortField, sortOrder string, filter types.CastListFilter) ([]*types.Cast, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 24
	}

	total, err := r.m.CountAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*types.Cast{}, 0, nil
	}

	offset := int64((page - 1) * pageSize)
	rows, err := r.m.ListPage(ctx, offset, int64(pageSize), buildCastOrderBy(sortField, sortOrder), filter)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func buildCastOrderBy(sortField, sortOrder string) string {
	field := normalizeCastSortField(sortField)
	order := normalizeCastSortOrder(sortOrder)

	column := "ac.owned_w_media_number"
	switch field {
	case "name":
		column = "ac.name"
	case "chinese":
		column = "cc.chinese"
	case "age":
		if order == "ASC" {
			return "CASE WHEN cc.birth_day > 0 THEN 0 ELSE 1 END ASC, cc.birth_day DESC, ac.owned_w_media_number DESC, ac.movie_number DESC, ac.name ASC, ac.id DESC"
		}
		return "CASE WHEN cc.birth_day > 0 THEN 0 ELSE 1 END ASC, cc.birth_day ASC, ac.owned_w_media_number DESC, ac.movie_number DESC, ac.name ASC, ac.id DESC"
	case "height":
		return "CASE WHEN cc.height > 0 THEN 0 ELSE 1 END ASC, cc.height " + order + ", ac.owned_w_media_number DESC, ac.movie_number DESC, ac.name ASC, ac.id DESC"
	case "movie_number":
		column = "ac.movie_number"
	case "owned_movie_number", "owned_w_media_number":
		column = "ac.owned_w_media_number"
	case "sc_times":
		column = "ac.sc_times"
	case "come_times":
		column = "ac.come_times"
	case "last_sc_time":
		column = "ac.last_sc_time"
	case "highest_rank":
		column = "ac.highest_rank"
	}

	if column == "ac.owned_w_media_number" {
		return column + " " + order + ", ac.movie_number DESC, ac.name ASC, ac.id DESC"
	}
	if column == "ac.name" {
		return column + " " + order + ", ac.owned_w_media_number DESC, ac.movie_number DESC, ac.id DESC"
	}
	if column == "cc.chinese" {
		return "CASE WHEN cc.chinese <> '' THEN 0 ELSE 1 END ASC, " + column + " " + order + ", ac.owned_w_media_number DESC, ac.movie_number DESC, ac.name ASC, ac.id DESC"
	}
	return column + " " + order + ", ac.owned_w_media_number DESC, ac.movie_number DESC, ac.name ASC, ac.id DESC"
}

func normalizeCastSortField(sortField string) string {
	switch strings.ToLower(strings.TrimSpace(sortField)) {
	case "owned_movie_number":
		return "owned_w_media_number"
	case "name", "chinese", "age", "height", "movie_number", "owned_w_media_number", "sc_times", "come_times", "last_sc_time", "highest_rank":
		return strings.ToLower(strings.TrimSpace(sortField))
	default:
		return "owned_w_media_number"
	}
}

func normalizeCastSortOrder(sortOrder string) string {
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		return "ASC"
	}
	return "DESC"
}
