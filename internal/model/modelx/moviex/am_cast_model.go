package moviex

import (
	"context"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

var _ AmCastModel = (*customAmCastModel)(nil)

type (
	// AmCastModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmCastModel.
	AmCastModel interface {
		amCastModel
		GetMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, error)
		GetMovieNumbersWithWMediaByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, int64, error)
		AggregatePersonStatsByIDs(ctx context.Context, ids []int64) (map[int64]*types.Person, error)
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		ListAllIDs(ctx context.Context) ([]int64, error)
		ListAll(ctx context.Context) ([]*AmCast, error)
		ListRowsByPersonIDs(ctx context.Context, personIDs []int64) ([]*AmCast, error)
		ListPage(ctx context.Context, offset, limit int64, orderBy string, filter types.CastListFilter) ([]*types.Cast, error)
		CountAll(ctx context.Context, filter types.CastListFilter) (int64, error)
		FindByNames(ctx context.Context, names []string) ([]*types.Cast, error)
		CountOwnedScMovieNumbersByNames(ctx context.Context, names []string) (map[string]int64, error)
	}

	customAmCastModel struct {
		*defaultAmCastModel
	}
)

// NewAmCastModel returns a model for the database table.
func NewAmCastModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmCastModel {
	return &customAmCastModel{
		defaultAmCastModel: newAmCastModel(conn, c, opts...),
	}
}

func (m *customAmCastModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customAmCastModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customAmCastModel) GetMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, error) {
	const query = `
SELECT
	(SELECT COUNT(DISTINCT movie_jav_id) FROM amr_movie_cast WHERE cast_id = ?) AS movie_number,
	(SELECT COUNT(DISTINCT amr.movie_jav_id)
		FROM amr_movie_cast amr
		JOIN w_media vf ON vf.movie_jav_id = amr.movie_jav_id AND vf.source_type = ? AND vf.is_removed = ?
		WHERE amr.cast_id = ?) AS owned_movie_number
`
	var resp struct {
		MovieNumber      int64 `db:"movie_number"`
		OwnedMovieNumber int64 `db:"owned_movie_number"`
	}
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, id, consts.WMediaSourceLegacyVFilm, ownedRemovedStatus, id); err != nil {
		return 0, 0, err
	}
	return resp.MovieNumber, resp.OwnedMovieNumber, nil
}

func (m *customAmCastModel) GetMovieNumbersWithWMediaByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, int64, error) {
	const query = `
SELECT
	(SELECT COUNT(DISTINCT movie_jav_id) FROM amr_movie_cast WHERE cast_id = ?) AS movie_number,
	(SELECT COUNT(DISTINCT amr.movie_jav_id)
		FROM amr_movie_cast amr
		JOIN w_media vf ON vf.movie_jav_id = amr.movie_jav_id AND vf.source_type = ? AND vf.is_removed = ?
		WHERE amr.cast_id = ?) AS owned_movie_number,
	(SELECT COUNT(DISTINCT amr.movie_jav_id)
		FROM amr_movie_cast amr
		JOIN w_media wm ON wm.movie_jav_id = amr.movie_jav_id AND wm.source_type = ? AND wm.is_removed = ?
		WHERE amr.cast_id = ?) AS owned_w_media_number
`
	var resp struct {
		MovieNumber       int64 `db:"movie_number"`
		OwnedMovieNumber  int64 `db:"owned_movie_number"`
		OwnedWMediaNumber int64 `db:"owned_w_media_number"`
	}
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query,
		id,
		consts.WMediaSourceLegacyVFilm, ownedRemovedStatus, id,
		consts.WMediaSourceNative, ownedRemovedStatus, id,
	); err != nil {
		return 0, 0, 0, err
	}
	return resp.MovieNumber, resp.OwnedMovieNumber, resp.OwnedWMediaNumber, nil
}

func (m *customAmCastModel) AggregatePersonStatsByIDs(ctx context.Context, ids []int64) (map[int64]*types.Person, error) {
	uniq := uniqueInt64s(ids)
	if len(uniq) == 0 {
		return map[int64]*types.Person{}, nil
	}

	type row struct {
		PersonId          int64 `db:"person_id"`
		MovieNumber       int64 `db:"movie_number"`
		OwnedMovieNumber  int64 `db:"owned_movie_number"`
		OwnedWMediaNumber int64 `db:"owned_w_media_number"`
		ScTimes           int64 `db:"sc_times"`
		ComeTimes         int64 `db:"come_times"`
		LastScTime        int64 `db:"last_sc_time"`
		HighestRank       int64 `db:"highest_rank"`
		RankTimes         int64 `db:"rank_times"`
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(uniq)), ",")
	query := `
SELECT
	pm.person_id AS person_id,
	COUNT(*) AS movie_number,
	COALESCE(SUM(CASE WHEN vf.is_removed = ? THEN 1 ELSE 0 END), 0) AS owned_movie_number,
	COALESCE(SUM(CASE WHEN wm.is_removed = ? THEN 1 ELSE 0 END), 0) AS owned_w_media_number,
	COALESCE(SUM(COALESCE(gss.sc_times, 0)), 0) AS sc_times,
	COALESCE(SUM(COALESCE(gss.come_times, 0)), 0) AS come_times,
	COALESCE(MAX(COALESCE(gss.last_sc_time, 0)), 0) AS last_sc_time,
	COALESCE(MIN(CASE WHEN mi.highest_rank > 0 AND mi.highest_rank < 1000 THEN mi.highest_rank END), 0) AS highest_rank,
	COALESCE(SUM(CASE WHEN mi.days_in_rank > 0 THEN mi.days_in_rank ELSE 0 END), 0) AS rank_times
FROM (
	SELECT DISTINCT ac.person_id AS person_id, amr.movie_jav_id AS movie_jav_id
	FROM am_cast ac
	JOIN amr_movie_cast amr ON amr.cast_id = ac.id
	WHERE ac.person_id IN (` + placeholders + `)
) pm
LEFT JOIN w_media vf ON vf.movie_jav_id = pm.movie_jav_id AND vf.source_type = ?
LEFT JOIN w_media wm ON wm.movie_jav_id = pm.movie_jav_id AND wm.source_type = ?
LEFT JOIN g_sc_stat gss ON gss.movie_jav_id = pm.movie_jav_id
LEFT JOIN bm_minfo mi ON mi.jav_id = pm.movie_jav_id
GROUP BY pm.person_id
`

	args := make([]any, 0, len(uniq)+4)
	args = append(args, consts.FilmIsNotRemoved, consts.FilmIsNotRemoved)
	for _, id := range uniq {
		args = append(args, id)
	}
	args = append(args, consts.WMediaSourceLegacyVFilm, consts.WMediaSourceNative)

	var rows []*row
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[int64]*types.Person{}, nil
		}
		return nil, err
	}

	out := make(map[int64]*types.Person, len(rows))
	for _, item := range rows {
		if item == nil || item.PersonId <= 0 {
			continue
		}
		out[item.PersonId] = &types.Person{
			Id:                item.PersonId,
			MovieNumber:       item.MovieNumber,
			OwnedMovieNumber:  item.OwnedMovieNumber,
			OwnedWMediaNumber: item.OwnedWMediaNumber,
			ScTimes:           item.ScTimes,
			ComeTimes:         item.ComeTimes,
			LastScTime:        item.LastScTime,
			HighestRank:       item.HighestRank,
			RankTimes:         item.RankTimes,
		}
	}
	return out, nil
}

func (m *customAmCastModel) ListAllIDs(ctx context.Context) ([]int64, error) {
	const query = "SELECT id FROM am_cast"
	var ids []int64
	if err := m.QueryRowsNoCacheCtx(ctx, &ids, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ids, nil
}

func (m *customAmCastModel) ListAll(ctx context.Context) ([]*AmCast, error) {
	var rows []*AmCast
	query := "select " + amCastRows + " from " + m.table + " order by id asc"
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*AmCast{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customAmCastModel) ListRowsByPersonIDs(ctx context.Context, personIDs []int64) ([]*AmCast, error) {
	uniq := uniqueInt64s(personIDs)
	if len(uniq) == 0 {
		return []*AmCast{}, nil
	}

	sqlStr, args, err := squirrel.
		Select(amCastRows).
		From(m.table).
		Where(squirrel.Eq{"person_id": uniq}).
		OrderBy("person_id ASC", "name ASC", "id ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*AmCast
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*AmCast{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customAmCastModel) ListPage(ctx context.Context, offset, limit int64, orderBy string, filter types.CastListFilter) ([]*types.Cast, error) {
	if orderBy == "" {
		orderBy = "ac.owned_movie_number DESC, ac.movie_number DESC, ac.name ASC, ac.id DESC"
	}

	type castListRow struct {
		Id                 int64  `db:"id"`
		PersonId           int64  `db:"person_id"`
		Name               string `db:"name"`
		JavId              string `db:"jav_id"`
		Chinese            string `db:"chinese"`
		BirthDay           int64  `db:"birth_day"`
		Height             int64  `db:"height"`
		MovieNumber        int64  `db:"movie_number"`
		OwnedMovieNumber   int64  `db:"owned_movie_number"`
		OwnedWMediaNumber  int64  `db:"owned_w_media_number"`
		ScTimes            int64  `db:"sc_times"`
		ComeTimes          int64  `db:"come_times"`
		LastScTime         int64  `db:"last_sc_time"`
		Rank500MovieNumber int64  `db:"rank500_movie_number"`
		Rank20MovieNumber  int64  `db:"rank20_movie_number"`
		Rank1MovieNumber   int64  `db:"rank1_movie_number"`
		HighestRank        int64  `db:"highest_rank"`
		RankTimes          int64  `db:"rank_times"`
		CreatedOn          int64  `db:"created_on"`
		UpdatedOn          int64  `db:"updated_on"`
	}

	builder := squirrel.
		Select(
			"ac.id AS id",
			"ac.person_id AS person_id",
			"ac.name AS name",
			"ac.jav_id AS jav_id",
			"COALESCE(p.chinese, '') AS chinese",
			"COALESCE(p.birth_day, 0) AS birth_day",
			"COALESCE(p.height, 0) AS height",
			"ac.movie_number AS movie_number",
			"ac.owned_movie_number AS owned_movie_number",
			"ac.owned_w_media_number AS owned_w_media_number",
			"ac.sc_times AS sc_times",
			"ac.come_times AS come_times",
			"ac.last_sc_time AS last_sc_time",
			"ac.rank500_movie_number AS rank500_movie_number",
			"ac.rank20_movie_number AS rank20_movie_number",
			"ac.rank1_movie_number AS rank1_movie_number",
			"ac.highest_rank AS highest_rank",
			"ac.rank_times AS rank_times",
			"ac.created_on AS created_on",
			"ac.updated_on AS updated_on",
		).
		From(m.table + " ac").
		LeftJoin("`c_person` p ON p.id = ac.person_id")

	builder = applyCastListFilter(builder, filter)

	sqlStr, args, err := builder.
		OrderBy(orderBy).
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*castListRow
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*types.Cast{}, nil
		}
		return nil, err
	}

	out := make([]*types.Cast, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, &types.Cast{
			Id:                 row.Id,
			PersonId:           row.PersonId,
			Name:               row.Name,
			JavId:              row.JavId,
			Chinese:            row.Chinese,
			BirthDay:           row.BirthDay,
			Height:             row.Height,
			MovieNumber:        row.MovieNumber,
			OwnedMovieNumber:   row.OwnedMovieNumber,
			OwnedWMediaNumber:  row.OwnedWMediaNumber,
			ScTimes:            row.ScTimes,
			ComeTimes:          row.ComeTimes,
			LastScTime:         row.LastScTime,
			Rank500MovieNumber: row.Rank500MovieNumber,
			Rank20MovieNumber:  row.Rank20MovieNumber,
			Rank1MovieNumber:   row.Rank1MovieNumber,
			HighestRank:        row.HighestRank,
			RankTimes:          row.RankTimes,
			CreatedOn:          row.CreatedOn,
			UpdatedOn:          row.UpdatedOn,
		})
	}
	return out, nil
}

func (m *customAmCastModel) CountAll(ctx context.Context, filter types.CastListFilter) (int64, error) {
	builder := squirrel.
		Select("COUNT(*)").
		From(m.table + " ac")

	builder = applyCastListFilter(builder, filter)

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, sqlStr, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func (m *customAmCastModel) FindByNames(ctx context.Context, names []string) ([]*types.Cast, error) {
	uniq := uniqueNonEmptyStrings(names)
	if len(uniq) == 0 {
		return []*types.Cast{}, nil
	}

	type castRow struct {
		Id                 int64  `db:"id"`
		PersonId           int64  `db:"person_id"`
		Name               string `db:"name"`
		JavId              string `db:"jav_id"`
		MovieNumber        int64  `db:"movie_number"`
		OwnedMovieNumber   int64  `db:"owned_movie_number"`
		OwnedWMediaNumber  int64  `db:"owned_w_media_number"`
		ScTimes            int64  `db:"sc_times"`
		ComeTimes          int64  `db:"come_times"`
		LastScTime         int64  `db:"last_sc_time"`
		Rank500MovieNumber int64  `db:"rank500_movie_number"`
		Rank20MovieNumber  int64  `db:"rank20_movie_number"`
		Rank1MovieNumber   int64  `db:"rank1_movie_number"`
		HighestRank        int64  `db:"highest_rank"`
		RankTimes          int64  `db:"rank_times"`
		CreatedOn          int64  `db:"created_on"`
		UpdatedOn          int64  `db:"updated_on"`
	}

	sqlStr, args, err := squirrel.
		Select(
			"id",
			"person_id",
			"name",
			"jav_id",
			"movie_number",
			"owned_movie_number",
			"owned_w_media_number",
			"sc_times",
			"come_times",
			"last_sc_time",
			"rank500_movie_number",
			"rank20_movie_number",
			"rank1_movie_number",
			"highest_rank",
			"rank_times",
			"created_on",
			"updated_on",
		).
		From(m.table).
		Where(squirrel.Eq{"name": uniq}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*castRow
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*types.Cast{}, nil
		}
		return nil, err
	}

	out := make([]*types.Cast, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, &types.Cast{
			Id:                 row.Id,
			PersonId:           row.PersonId,
			Name:               row.Name,
			JavId:              row.JavId,
			MovieNumber:        row.MovieNumber,
			OwnedMovieNumber:   row.OwnedMovieNumber,
			OwnedWMediaNumber:  row.OwnedWMediaNumber,
			ScTimes:            row.ScTimes,
			ComeTimes:          row.ComeTimes,
			LastScTime:         row.LastScTime,
			Rank500MovieNumber: row.Rank500MovieNumber,
			Rank20MovieNumber:  row.Rank20MovieNumber,
			Rank1MovieNumber:   row.Rank1MovieNumber,
			HighestRank:        row.HighestRank,
			RankTimes:          row.RankTimes,
			CreatedOn:          row.CreatedOn,
			UpdatedOn:          row.UpdatedOn,
		})
	}
	return out, nil
}

func (m *customAmCastModel) CountOwnedScMovieNumbersByNames(ctx context.Context, names []string) (map[string]int64, error) {
	uniq := uniqueNonEmptyStrings(names)
	if len(uniq) == 0 {
		return map[string]int64{}, nil
	}

	type row struct {
		Name               string `db:"name"`
		OwnedScMovieNumber int64  `db:"owned_sc_movie_number"`
	}

	sqlStr, args, err := squirrel.
		Select(
			"ac.name AS name",
			"COUNT(DISTINCT CASE WHEN vf.is_removed = ? AND COALESCE(gss.sc_times, 0) > 0 THEN amr.movie_jav_id END) AS owned_sc_movie_number",
		).
		From(m.table + " ac").
		LeftJoin("`amr_movie_cast` amr ON amr.cast_id = ac.id").
		LeftJoin(buildLegacyWMediaJoin("`w_media`", "vf", "amr.movie_jav_id")).
		LeftJoin("`g_sc_stat` gss ON gss.movie_jav_id = amr.movie_jav_id").
		Where(squirrel.Eq{"ac.name": uniq}).
		GroupBy("ac.name").
		PlaceholderFormat(squirrel.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	args = append([]any{consts.FilmIsNotRemoved}, args...)

	var rows []*row
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[string]int64{}, nil
		}
		return nil, err
	}

	out := make(map[string]int64, len(rows))
	for _, item := range rows {
		if item == nil || item.Name == "" {
			continue
		}
		out[item.Name] = item.OwnedScMovieNumber
	}
	return out, nil
}

func applyCastListFilter(builder squirrel.SelectBuilder, filter types.CastListFilter) squirrel.SelectBuilder {
	builder = builder.Where(squirrel.Gt{"ac.owned_movie_number": 0})

	if filter.HasOwnedMin {
		builder = builder.Where(squirrel.GtOrEq{"ac.owned_movie_number": filter.OwnedMin})
	}
	if filter.HasOwnedMax {
		builder = builder.Where(squirrel.LtOrEq{"ac.owned_movie_number": filter.OwnedMax})
	}
	if filter.HasScTimesMin {
		builder = builder.Where(squirrel.GtOrEq{"ac.sc_times": filter.ScTimesMin})
	}
	if filter.HasScTimesMax {
		builder = builder.Where(squirrel.LtOrEq{"ac.sc_times": filter.ScTimesMax})
	}
	if filter.HasComeTimesMin {
		builder = builder.Where(squirrel.GtOrEq{"ac.come_times": filter.ComeTimesMin})
	}
	if filter.HasComeTimesMax {
		builder = builder.Where(squirrel.LtOrEq{"ac.come_times": filter.ComeTimesMax})
	}
	if filter.HasLastScFrom {
		builder = builder.Where(squirrel.GtOrEq{"ac.last_sc_time": filter.LastScFrom})
	}
	if filter.HasLastScTo {
		builder = builder.Where(squirrel.LtOrEq{"ac.last_sc_time": filter.LastScTo})
	}

	return builder
}

func uniqueNonEmptyStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
