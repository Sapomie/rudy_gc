package moviex

import (
	"context"
	"errors"
	"rudy_gc/internal/consts"
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

func (m *customVFilmModel) TableName() string { return m.table }
func (m *customVFilmModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}
func (m *customVFilmModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
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
