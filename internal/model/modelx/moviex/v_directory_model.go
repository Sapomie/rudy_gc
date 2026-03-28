// internal/model/modelx/moviex/v_directory_model_custom.go
package moviex

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ VDirectoryModel = (*customVDirectoryModel)(nil)

type (
	// VDirectoryModel: 在 goctl 生成的 vDirectoryModel 基础上扩展
	VDirectoryModel interface {
		vDirectoryModel
		// 已有你贴的：
		FindOneByName(ctx context.Context, name string) (*VDirectory, error)

		ListByParent(ctx context.Context, parentID int64, page, size int) (rows []*VDirectory, total int64, err error)
		ListSiblings(ctx context.Context, parentID int64, limit int) ([]*VDirectory, error)
		ListSubtreeIDsByPath(ctx context.Context, basePath string) ([]int64, error)
	}

	customVDirectoryModel struct {
		*defaultVDirectoryModel
	}

	// 目录 + 聚合结果
	VDirWithAgg struct {
		Id        int64  `db:"id"`
		ParentId  int64  `db:"parent_id"`
		Name      string `db:"name"`
		Depth     int64  `db:"depth"`
		Path      string `db:"path"`
		UpdatedOn int64  `db:"updated_on"`

		FilmCount     sql.NullInt64 `db:"film_count"`
		TotalSize     sql.NullInt64 `db:"total_size"`
		LastFilmBirth sql.NullInt64 `db:"last_film_birth"`
		LastUpdatedOn sql.NullInt64 `db:"last_updated_on"`
	}
)

// NewVDirectoryModel returns a model for the database table.
func NewVDirectoryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) VDirectoryModel {
	return &customVDirectoryModel{
		defaultVDirectoryModel: newVDirectoryModel(conn, c, opts...),
	}
}

// ---------- 列表（基础） ----------
func (m *customVDirectoryModel) ListByParent(ctx context.Context, parentID int64, page, size int) ([]*VDirectory, int64, error) {
	offset := (page - 1) * size

	// total
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
		return []*VDirectory{}, 0, nil
	}

	// list
	q, args, err := squirrel.
		Select(vDirectoryRows).
		From(m.table + " AS d").
		Where(squirrel.Eq{"d.parent_id": parentID}).
		Limit(uint64(size)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, 0, err
	}
	var rows []*VDirectory
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, q, args...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// ---------- 同级 ----------
func (m *customVDirectoryModel) ListSiblings(ctx context.Context, parentID int64, limit int) ([]*VDirectory, error) {
	if limit <= 0 || limit > 500 {
		limit = 300
	}
	q, args, err := squirrel.
		Select(vDirectoryRows).
		From(m.table + " AS d").
		Where(squirrel.Eq{"d.parent_id": parentID}).
		OrderBy("d.name ASC").
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*VDirectory
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, q, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

// ---------- 子树 ID 列表 ----------
func (m *customVDirectoryModel) ListSubtreeIDsByPath(ctx context.Context, basePath string) ([]int64, error) {
	// path = basePath OR path LIKE basePath + '/%'
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
	for _, r := range rows {
		out = append(out, r.Id)
	}
	return out, nil
}
