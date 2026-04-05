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

var _ CPersonModel = (*customCPersonModel)(nil)

type (
	CPersonModel interface {
		cPersonModel
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		ListAll(ctx context.Context) ([]*CPerson, error)
		FindOneByNameOrAliasToken(ctx context.Context, name string) (*CPerson, error)
		SearchMergeCandidates(ctx context.Context, keyword string, excludeIDs []int64, limit int64) ([]*types.Person, error)
		ListPage(ctx context.Context, offset, limit int64, orderBy string, filter types.PersonListFilter) ([]*types.Person, error)
		CountAll(ctx context.Context, filter types.PersonListFilter) (int64, error)
		FindByIDs(ctx context.Context, ids []int64) ([]*types.Person, error)
		CountOwnedScMovieNumbersByIDs(ctx context.Context, ids []int64) (map[int64]int64, error)
	}

	customCPersonModel struct {
		*defaultCPersonModel
	}
)

func NewCPersonModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CPersonModel {
	return &customCPersonModel{
		defaultCPersonModel: newCPersonModel(conn, c, opts...),
	}
}

func (m *customCPersonModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customCPersonModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customCPersonModel) ListAll(ctx context.Context) ([]*CPerson, error) {
	var rows []*CPerson
	query := "select " + cPersonRows + " from " + m.table + " order by id asc"
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*CPerson{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customCPersonModel) FindOneByNameOrAliasToken(ctx context.Context, name string) (*CPerson, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrNotFound
	}

	query := "select " + cPersonRows + " from " + m.table + " where `name` = ? or `alias` like ? escape '\\\\' order by case when `name` = ? then 0 else 1 end, `movie_number` desc, `owned_movie_number` desc, `id` asc limit 64"
	likePattern := "%" + escapeSQLLikeValue(name) + "%"

	var rows []*CPerson
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, name, likePattern, name); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	norm := normalizePersonAliasToken(name)
	for _, row := range rows {
		if row == nil {
			continue
		}
		if normalizePersonAliasToken(row.Name) == norm {
			return row, nil
		}
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		if personAliasContainsToken(row.Alias, norm) {
			return row, nil
		}
	}

	return nil, ErrNotFound
}

func (m *customCPersonModel) SearchMergeCandidates(ctx context.Context, keyword string, excludeIDs []int64, limit int64) ([]*types.Person, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []*types.Person{}, nil
	}
	if limit <= 0 {
		limit = 12
	}

	like := "%" + keyword + "%"
	args := []any{like, like, like, like}
	query := `
SELECT DISTINCT
	p.id AS id,
	p.name AS name,
	p.alias AS alias,
	p.chinese AS chinese,
	p.birth_day AS birth_day,
	p.height AS height,
	p.cup AS cup,
	p.bwh AS bwh,
	p.avatar AS avatar,
	p.movie_number AS movie_number,
	p.owned_movie_number AS owned_movie_number,
	p.sc_times AS sc_times,
	p.come_times AS come_times,
	p.last_sc_time AS last_sc_time,
	p.highest_rank AS highest_rank,
	p.rank_times AS rank_times,
	p.created_on AS created_on,
	p.updated_on AS updated_on
FROM ` + m.table + ` p
LEFT JOIN am_cast ac ON ac.person_id = p.id
WHERE (
	p.name LIKE ?
	OR p.chinese LIKE ?
	OR p.alias LIKE ?
	OR ac.name LIKE ?
)`

	uniqExclude := uniqueInt64s(excludeIDs)
	if len(uniqExclude) > 0 {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(uniqExclude)), ",")
		query += " AND p.id NOT IN (" + placeholders + ")"
		for _, id := range uniqExclude {
			args = append(args, id)
		}
	}

	query += " ORDER BY p.movie_number DESC, p.owned_movie_number DESC, p.name ASC, p.id ASC LIMIT ?"
	args = append(args, limit)

	var rows []*CPerson
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*types.Person{}, nil
		}
		return nil, err
	}

	out := make([]*types.Person, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, mapCPersonModelToTypes(row))
	}
	return out, nil
}

func (m *customCPersonModel) ListPage(ctx context.Context, offset, limit int64, orderBy string, filter types.PersonListFilter) ([]*types.Person, error) {
	if orderBy == "" {
		orderBy = "p.owned_movie_number DESC, p.movie_number DESC, p.name ASC, p.id DESC"
	}

	sqlStr, args, err := squirrel.
		Select(
			"p.id",
			"p.name",
			"p.alias",
			"p.chinese",
			"p.birth_day",
			"p.height",
			"p.cup",
			"p.bwh",
			"p.avatar",
			"p.movie_number",
			"p.owned_movie_number",
			"p.sc_times",
			"p.come_times",
			"p.last_sc_time",
			"p.highest_rank",
			"p.rank_times",
			"p.created_on",
			"p.updated_on",
		).
		From(m.table + " p").
		Where(applyCPersonListFilter(filter)).
		OrderBy(orderBy).
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*CPerson
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*types.Person{}, nil
		}
		return nil, err
	}

	out := make([]*types.Person, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, mapCPersonModelToTypes(row))
	}
	return out, nil
}

func (m *customCPersonModel) CountAll(ctx context.Context, filter types.PersonListFilter) (int64, error) {
	sqlStr, args, err := squirrel.
		Select("COUNT(*)").
		From(m.table + " p").
		Where(applyCPersonListFilter(filter)).
		ToSql()
	if err != nil {
		return 0, err
	}

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, sqlStr, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func (m *customCPersonModel) FindByIDs(ctx context.Context, ids []int64) ([]*types.Person, error) {
	uniq := uniqueInt64s(ids)
	if len(uniq) == 0 {
		return []*types.Person{}, nil
	}

	sqlStr, args, err := squirrel.
		Select(cPersonRows).
		From(m.table).
		Where(squirrel.Eq{"id": uniq}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*CPerson
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*types.Person{}, nil
		}
		return nil, err
	}

	out := make([]*types.Person, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, mapCPersonModelToTypes(row))
	}
	return out, nil
}

func (m *customCPersonModel) CountOwnedScMovieNumbersByIDs(ctx context.Context, ids []int64) (map[int64]int64, error) {
	uniq := uniqueInt64s(ids)
	if len(uniq) == 0 {
		return map[int64]int64{}, nil
	}

	type row struct {
		Id                 int64 `db:"id"`
		OwnedScMovieNumber int64 `db:"owned_sc_movie_number"`
	}

	sqlStr, args, err := squirrel.
		Select(
			"p.id AS id",
			"COUNT(DISTINCT CASE WHEN vf.is_removed = ? AND COALESCE(gss.sc_times, 0) > 0 THEN amr.movie_jav_id END) AS owned_sc_movie_number",
		).
		From(m.table + " p").
		LeftJoin("`am_cast` ac ON ac.person_id = p.id").
		LeftJoin("`amr_movie_cast` amr ON amr.cast_id = ac.id").
		LeftJoin("`v_film` vf ON vf.movie_jav_id = amr.movie_jav_id").
		LeftJoin("`g_sc_stat` gss ON gss.movie_jav_id = amr.movie_jav_id").
		Where(squirrel.Eq{"p.id": uniq}).
		GroupBy("p.id").
		PlaceholderFormat(squirrel.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	args = append([]any{consts.FilmIsNotRemoved}, args...)

	var rows []*row
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[int64]int64{}, nil
		}
		return nil, err
	}

	out := make(map[int64]int64, len(rows))
	for _, item := range rows {
		if item == nil || item.Id <= 0 {
			continue
		}
		out[item.Id] = item.OwnedScMovieNumber
	}
	return out, nil
}

func mapCPersonModelToTypes(v *CPerson) *types.Person {
	if v == nil {
		return nil
	}
	return &types.Person{
		Id:               v.Id,
		Name:             v.Name,
		Alias:            v.Alias,
		Chinese:          v.Chinese,
		BirthDay:         v.BirthDay,
		Height:           v.Height,
		Cup:              v.Cup,
		Bwh:              v.Bwh,
		Avatar:           v.Avatar,
		MovieNumber:      v.MovieNumber,
		OwnedMovieNumber: v.OwnedMovieNumber,
		ScTimes:          v.ScTimes,
		ComeTimes:        v.ComeTimes,
		LastScTime:       v.LastScTime,
		HighestRank:      v.HighestRank,
		RankTimes:        v.RankTimes,
		CreatedOn:        v.CreatedOn,
		UpdatedOn:        v.UpdatedOn,
	}
}

func applyCPersonListFilter(filter types.PersonListFilter) squirrel.And {
	w := squirrel.And{}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + escapeSQLLikeValue(keyword) + "%"
		w = append(w, squirrel.Expr(
			"(p.name LIKE ? ESCAPE '\\\\' OR p.chinese LIKE ? ESCAPE '\\\\' OR p.alias LIKE ? ESCAPE '\\\\')",
			like,
			like,
			like,
		))
	}
	if filter.HasOwnedMin {
		w = append(w, squirrel.GtOrEq{"p.owned_movie_number": filter.OwnedMin})
	}
	if filter.HasOwnedMax {
		w = append(w, squirrel.LtOrEq{"p.owned_movie_number": filter.OwnedMax})
	}
	if filter.HasScTimesMin {
		w = append(w, squirrel.GtOrEq{"p.sc_times": filter.ScTimesMin})
	}
	if filter.HasScTimesMax {
		w = append(w, squirrel.LtOrEq{"p.sc_times": filter.ScTimesMax})
	}
	if filter.HasComeTimesMin {
		w = append(w, squirrel.GtOrEq{"p.come_times": filter.ComeTimesMin})
	}
	if filter.HasComeTimesMax {
		w = append(w, squirrel.LtOrEq{"p.come_times": filter.ComeTimesMax})
	}
	if filter.HasLastScFrom {
		w = append(w, squirrel.GtOrEq{"p.last_sc_time": filter.LastScFrom})
	}
	if filter.HasLastScTo {
		w = append(w, squirrel.LtOrEq{"p.last_sc_time": filter.LastScTo})
	}
	return w
}

func escapeSQLLikeValue(v string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_")
	return replacer.Replace(v)
}

func normalizePersonAliasToken(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func splitPersonAliasTokens(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	replacer := strings.NewReplacer("，", ",", "、", ",", "/", ",", "|", ",", "\n", ",", ";", ",", "；", ",")
	parts := strings.Split(replacer.Replace(raw), ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func personAliasContainsToken(alias string, target string) bool {
	if target == "" {
		return false
	}
	for _, token := range splitPersonAliasTokens(alias) {
		if normalizePersonAliasToken(token) == target {
			return true
		}
	}
	return false
}

func uniqueInt64s(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
