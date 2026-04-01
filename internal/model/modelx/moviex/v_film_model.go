package moviex

import (
	"context"
	"errors"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ VFilmModel = (*customVFilmModel)(nil)

type (
	VFilmModel interface {
		vFilmModel

		FindAll(ctx context.Context, removedStatus int64) ([]*VFilm, error)
		ListByMovieJavIds(ctx context.Context, movieJavIds []string) ([]*VFilm, error)
		ListPage(ctx context.Context, offset, limit int64, orderBy string, filter types.FilmListFilter) ([]*VFilm, error)
		CountAll(ctx context.Context, filter types.FilmListFilter) (int64, error)
		TableName() string
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error

		// ✅ 新增：按目录ID集合分页查询（仅 is_removed=0），返回 rows + total
		ListByDirectoryIDs(ctx context.Context, dirIDs []int64, page, size int, orderBy string) (all, paged []*VFilm, total int64, err error)
	}

	customVFilmModel struct {
		*defaultVFilmModel
	}
)

func NewVFilmModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) VFilmModel {
	return &customVFilmModel{
		defaultVFilmModel: newVFilmModel(conn, c, opts...),
	}
}

// ------- 你已有的 -------
func (m *customVFilmModel) FindAll(ctx context.Context, removedStatus int64) ([]*VFilm, error) {
	builder := squirrel.Select(vFilmRows).From(m.table)
	if removedStatus > 0 {
		builder = builder.Where(squirrel.Eq{"is_removed": removedStatus})
	}
	q, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var list []*VFilm
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*VFilm{}, nil
		}
		return nil, err
	}
	return list, nil
}

func (m *customVFilmModel) ListByMovieJavIds(ctx context.Context, movieJavIds []string) ([]*VFilm, error) {
	if len(movieJavIds) == 0 {
		return []*VFilm{}, nil
	}

	query, args, err := squirrel.
		Select(vFilmRows).
		From(m.table).
		Where(squirrel.Eq{"movie_jav_id": movieJavIds}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*VFilm
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*VFilm{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customVFilmModel) TableName() string { return m.table }
func (m *customVFilmModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}
func (m *customVFilmModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customVFilmModel) ListPage(ctx context.Context, offset, limit int64, orderBy string, filter types.FilmListFilter) ([]*VFilm, error) {
	orderParts := splitOrder(orderBy)
	if len(orderParts) == 0 {
		orderParts = []string{"birth_time DESC", "movie_name DESC"}
	}

	builder := applyVFilmListFilter(squirrel.
		Select(vFilmRows).
		From(m.table+" AS f"), filter).
		OrderBy(orderParts...)

	if limit > 0 {
		builder = builder.Limit(uint64(limit))
	}
	if offset > 0 {
		builder = builder.Offset(uint64(offset))
	}

	q, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var list []*VFilm
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*VFilm{}, nil
		}
		return nil, err
	}
	return list, nil
}

func (m *customVFilmModel) CountAll(ctx context.Context, filter types.FilmListFilter) (int64, error) {
	q, args, err := applyVFilmListFilter(squirrel.
		Select("COUNT(1)").
		From(m.table+" AS f"), filter).
		ToSql()
	if err != nil {
		return 0, err
	}

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, q, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return total, nil
}

func (m *customVFilmModel) ListByDirectoryIDs(ctx context.Context, dirIDs []int64, page, size int, orderBy string) (all, paged []*VFilm, total int64, err error) {

	if len(dirIDs) == 0 {
		return []*VFilm{}, []*VFilm{}, 0, nil
	}

	// ✅ 直接使用上层传入的 orderBy
	orderParts := splitOrder(orderBy)
	if len(orderParts) == 0 {
		orderParts = []string{"birth_time DESC"} // 兜底，防止为空
	}

	// 拉取全部匹配（仅 is_removed=0）
	qAll, argsAll, e := squirrel.
		Select(vFilmRows).
		From(m.table + " AS f").
		Where(squirrel.Eq{"f.is_removed": consts.FilmIsNotRemoved}).
		Where(squirrel.Eq{"f.directory_id": dirIDs}).
		OrderBy(orderParts...).
		ToSql()
	if e != nil {
		return nil, nil, 0, e
	}

	var rows []*VFilm
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, qAll, argsAll...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*VFilm{}, []*VFilm{}, 0, nil
		}
		return nil, nil, 0, err
	}

	// 内存分页
	total = int64(len(rows))
	if total == 0 {
		return []*VFilm{}, []*VFilm{}, 0, nil
	}
	start := (page - 1) * size
	if start >= len(rows) {
		return rows, []*VFilm{}, total, nil
	}
	end := start + size
	if end > len(rows) {
		end = len(rows)
	}
	paged = rows[start:end]
	return rows, paged, total, nil
}

// "a DESC,b DESC" -> []{"a DESC","b DESC"}
func splitOrder(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func applyVFilmListFilter(builder squirrel.SelectBuilder, filter types.FilmListFilter) squirrel.SelectBuilder {
	builder = builder.Where(squirrel.Eq{"f.is_removed": consts.FilmIsNotRemoved})

	if kw := strings.TrimSpace(filter.MovieNameKeyword); kw != "" {
		builder = builder.Where(squirrel.Like{"f.movie_name": "%" + kw + "%"})
	}

	if filter.HasSizeMin {
		builder = builder.Where(squirrel.GtOrEq{"f.size": filter.SizeMin})
	}
	if filter.HasSizeMax {
		builder = builder.Where(squirrel.LtOrEq{"f.size": filter.SizeMax})
	}

	if filter.HasHeightMin {
		builder = builder.Where(squirrel.GtOrEq{"f.height": filter.HeightMin})
	}
	if filter.HasHeightMax {
		builder = builder.Where(squirrel.LtOrEq{"f.height": filter.HeightMax})
	}

	if filter.HasDurationMin {
		builder = builder.Where(squirrel.GtOrEq{"f.duration": filter.DurationMin})
	}
	if filter.HasDurationMax {
		builder = builder.Where(squirrel.LtOrEq{"f.duration": filter.DurationMax})
	}

	if filter.HasBitRateMin {
		builder = builder.Where(squirrel.GtOrEq{"f.bit_rate": filter.BitRateMin})
	}
	if filter.HasBitRateMax {
		builder = builder.Where(squirrel.LtOrEq{"f.bit_rate": filter.BitRateMax})
	}

	if filter.HasFrameAverageMin {
		builder = builder.Where(squirrel.GtOrEq{"f.frame_average": filter.FrameAverageMin})
	}
	if filter.HasFrameAverageMax {
		builder = builder.Where(squirrel.LtOrEq{"f.frame_average": filter.FrameAverageMax})
	}

	if filter.HasSelfMake {
		builder = builder.Where(squirrel.Eq{"f.self_make": filter.SelfMake})
	}

	if filter.HasHasMask {
		builder = builder.Where(squirrel.Eq{"f.has_mask": filter.HasMask})
	}

	if filter.HasScTimesMin {
		builder = builder.Where(squirrel.GtOrEq{"f.sc_times": filter.ScTimesMin})
	}
	if filter.HasScTimesMax {
		builder = builder.Where(squirrel.LtOrEq{"f.sc_times": filter.ScTimesMax})
	}

	if filter.HasLastScFrom {
		builder = builder.Where(squirrel.GtOrEq{"f.last_sc_time": filter.LastScFrom})
	}
	if filter.HasLastScTo {
		builder = builder.Where(squirrel.LtOrEq{"f.last_sc_time": filter.LastScTo})
	}

	if filter.HasBirthTimeFrom {
		builder = builder.Where(squirrel.GtOrEq{"f.birth_time": filter.BirthTimeFrom})
	}
	if filter.HasBirthTimeTo {
		builder = builder.Where(squirrel.LtOrEq{"f.birth_time": filter.BirthTimeTo})
	}

	if filter.HasReleasingDateFrom {
		builder = builder.Where(squirrel.GtOrEq{"f.releasing_date": filter.ReleasingDateFrom})
	}
	if filter.HasReleasingDateTo {
		builder = builder.Where(squirrel.LtOrEq{"f.releasing_date": filter.ReleasingDateTo})
	}

	return builder
}
