package moviex

import (
	"context"
	"errors"
	"fmt"
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
		ListByDirectoryIDs(ctx context.Context, dirIDs []int64, page, size int, sortField string, asc bool) (rows []*VFilm, total int64, err error)
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

// ------- ✅ 新增的分页查询 -------
func (m *customVFilmModel) ListByDirectoryIDs(ctx context.Context, dirIDs []int64, page, size int, sortField string, asc bool) ([]*VFilm, int64, error) {
	if len(dirIDs) == 0 {
		return []*VFilm{}, 0, nil
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 24
	}
	offset := (page - 1) * size

	// 只允许白名单字段参与排序，避免 SQL 注入
	col := normalizeFilmSort(sortField)
	order := "DESC"
	if asc {
		order = "ASC"
	}

	// total
	countQ, countArgs, err := squirrel.
		Select("COUNT(*)").
		From(m.table + " AS f").
		Where(squirrel.Eq{"f.is_removed": consts.FilmIsNotRemoved}).
		Where(squirrel.Eq{"f.directory_id": dirIDs}).
		ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQ, countArgs...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*VFilm{}, 0, nil
	}

	// list
	qb := squirrel.
		Select(vFilmRows).
		From(m.table + " AS f").
		Where(squirrel.Eq{"f.is_removed": consts.FilmIsNotRemoved}).
		Where(squirrel.Eq{"f.directory_id": dirIDs}).
		OrderBy(fmt.Sprintf("%s %s", col, order)).
		Limit(uint64(size)).
		Offset(uint64(offset))

	listQ, listArgs, err := qb.ToSql()
	if err != nil {
		return nil, 0, err
	}
	var rows []*VFilm
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, listQ, listArgs...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*VFilm{}, 0, nil
		}
		return nil, 0, err
	}
	return rows, total, nil
}

// 只允许这些字段排序
func normalizeFilmSort(in string) string {
	switch strings.ToLower(in) {
	case "updated_on":
		return "f.updated_on"
	case "created_on":
		return "f.created_on"
	case "birth_time", "releasing_date":
		return "f.birth_time"
	case "size":
		return "f.size"
	case "movie_name":
		return "f.movie_name"
	case "file_name":
		return "f.file_name"
	case "duration":
		return "f.duration"
	default:
		return "f.updated_on"
	}
}
