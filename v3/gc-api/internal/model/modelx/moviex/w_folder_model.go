package moviex

import (
	"context"
	"rudy_gc/internal/consts"

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
		FindOneByName(ctx context.Context, name string) (*WFolder, error)
		FindOneByNameSourceType(ctx context.Context, name string, sourceType int64) (*WFolder, error)
		FindOneByParentIdName(ctx context.Context, parentID int64, name string) (*WFolder, error)
		FindOneByParentIdNameSourceType(ctx context.Context, parentID int64, name string, sourceType int64) (*WFolder, error)
		FindOneByPath(ctx context.Context, path string) (*WFolder, error)
		FindOneByPathSourceType(ctx context.Context, path string, sourceType int64) (*WFolder, error)
		ListByParent(ctx context.Context, parentID int64, page, size int) (rows []*WFolder, total int64, err error)
		ListByParentSourceType(ctx context.Context, parentID int64, sourceType int64, page, size int) (rows []*WFolder, total int64, err error)
		ListAll(ctx context.Context) ([]*WFolder, error)
		ListAllBySourceType(ctx context.Context, sourceType int64) ([]*WFolder, error)
		ListSubtreeIDsByPath(ctx context.Context, basePath string) ([]int64, error)
		ListSubtreeIDsByPathSourceType(ctx context.Context, basePath string, sourceType int64) ([]int64, error)
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

func (m *customWFolderModel) FindOneByName(ctx context.Context, name string) (*WFolder, error) {
	return m.FindOneByNameSourceType(ctx, name, consts.WFolderSourceNative)
}

func (m *customWFolderModel) FindOneByNameSourceType(ctx context.Context, name string, sourceType int64) (*WFolder, error) {
	q, args, err := squirrel.
		Select(wFolderRows).
		From(m.table+" AS d").
		Where(squirrel.Eq{"d.name": name, "d.source_type": sourceType}).
		OrderBy("d.depth ASC", "d.id ASC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}

	var row WFolder
	if err := m.QueryRowNoCacheCtx(ctx, &row, q, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *customWFolderModel) FindOneByParentIdName(ctx context.Context, parentID int64, name string) (*WFolder, error) {
	return m.FindOneByParentIdNameSourceType(ctx, parentID, name, consts.WFolderSourceNative)
}

func (m *customWFolderModel) FindOneByPath(ctx context.Context, path string) (*WFolder, error) {
	return m.FindOneByPathSourceType(ctx, path, consts.WFolderSourceNative)
}

func (m *customWFolderModel) ListByParent(ctx context.Context, parentID int64, page, size int) ([]*WFolder, int64, error) {
	return m.ListByParentSourceType(ctx, parentID, consts.WFolderSourceNative, page, size)
}

func (m *customWFolderModel) ListByParentSourceType(ctx context.Context, parentID int64, sourceType int64, page, size int) ([]*WFolder, int64, error) {
	offset := (page - 1) * size

	countQ, countArgs, err := squirrel.
		Select("COUNT(*)").
		From(m.table + " AS d").
		Where(squirrel.Eq{"d.parent_id": parentID, "d.source_type": sourceType}).
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
		Where(squirrel.Eq{"d.parent_id": parentID, "d.source_type": sourceType}).
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
	return m.ListAllBySourceType(ctx, consts.WFolderSourceNative)
}

func (m *customWFolderModel) ListAllBySourceType(ctx context.Context, sourceType int64) ([]*WFolder, error) {
	q, args, err := squirrel.
		Select(wFolderRows).
		From(m.table+" AS d").
		Where(squirrel.Eq{"d.source_type": sourceType}).
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
	return m.ListSubtreeIDsByPathSourceType(ctx, basePath, consts.WFolderSourceNative)
}

func (m *customWFolderModel) ListSubtreeIDsByPathSourceType(ctx context.Context, basePath string, sourceType int64) ([]int64, error) {
	q, args, err := squirrel.
		Select("d.id").
		From(m.table + " AS d").
		Where(squirrel.Eq{"d.source_type": sourceType}).
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
