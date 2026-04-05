package moviex

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WFolderModel = (*customWFolderModel)(nil)

type (
	// WFolderModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWFolderModel.
	WFolderModel interface {
		wFolderModel
		ListByParent(ctx context.Context, parentID int64, page, size int) (rows []*WFolder, total int64, err error)
		ListAll(ctx context.Context) ([]*WFolder, error)
		ListSubtreeIDsByPath(ctx context.Context, basePath string) ([]int64, error)
	}

	customWFolderModel struct {
		*defaultWFolderModel
	}
)

// NewWFolderModel returns a model for the database table.
func NewWFolderModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WFolderModel {
	return &customWFolderModel{
		defaultWFolderModel: newWFolderModel(conn, c, opts...),
	}
}

func (m *customWFolderModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customWFolderModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customWFolderModel) ListByParent(ctx context.Context, parentID int64, page, size int) ([]*WFolder, int64, error) {
	offset := (page - 1) * size

	countQ, countArgs, err := squirrel.
		Select("COUNT(*)").
		From(m.table + " AS d").
		Where(squirrel.Eq{"d.parent_id": parentID}).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQ, countArgs...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*WFolder{}, 0, nil
	}

	q, args, err := squirrel.
		Select(wFolderRows).
		From(m.table + " AS d").
		Where(squirrel.Eq{"d.parent_id": parentID}).
		OrderBy("d.id ASC").
		Limit(uint64(size)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	var rows []*WFolder
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, q, args...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (m *customWFolderModel) ListAll(ctx context.Context) ([]*WFolder, error) {
	q, args, err := squirrel.
		Select(wFolderRows).
		From(m.table+" AS d").
		OrderBy("CHAR_LENGTH(d.path) ASC", "d.id ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*WFolder
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, q, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *customWFolderModel) ListSubtreeIDsByPath(ctx context.Context, basePath string) ([]int64, error) {
	q, args, err := squirrel.
		Select("d.id").
		From(m.table + " AS d").
		Where(squirrel.Or{
			squirrel.Eq{"d.path": basePath},
			squirrel.Like{"d.path": basePath + "/%"},
		}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []struct {
		Id int64 `db:"id"`
	}
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, q, args...); err != nil {
		return nil, err
	}

	out := make([]int64, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Id)
	}
	return out, nil
}
