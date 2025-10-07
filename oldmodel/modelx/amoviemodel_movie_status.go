package modelx

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
)

func (m *defaultAMovieModel) FindMovieHasNoCover(ctx context.Context) (*AMovie, int64, error) {
	return m.FindMovieWithCondition(ctx, "`has_downloaded_cover` = ?", MovieHasNoLocalCover)
}

func (m *defaultAMovieModel) FindMovieHasNoChinese(ctx context.Context) (*AMovie, int64, error) {
	return m.FindMovieWithCondition(ctx, "`has_chinese` = ?", MovieHasNoChinese)
}

// FindMovieWithCondition 查找满足特定条件的电影并返回电影和总数
func (m *defaultAMovieModel) FindMovieWithCondition(ctx context.Context, condition string, value interface{}) (*AMovie, int64, error) {
	// 使用 squirrel 构建 SQL 查询
	db := squirrel.Select("*").
		From(m.tableName()).
		Where(condition, value)

	// Count query
	var total int64
	countQuery, countArgs, err := db.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("building count SQL query failed: %w", err)
	}
	finalCountQuery := "SELECT COUNT(*) FROM (" + countQuery + ") AS count_query"
	if err := m.QueryRowNoCacheCtx(ctx, &total, finalCountQuery, countArgs...); err != nil {
		return nil, 0, err
	}

	// Result query
	var movie AMovie
	resultQuery, resultArgs, err := db.Limit(1).ToSql()
	if err != nil {
		return nil, 0, err
	}
	if err := m.QueryRowNoCacheCtx(ctx, &movie, resultQuery, resultArgs...); err != nil {
		return nil, 0, err
	}

	return &movie, total, nil
}
