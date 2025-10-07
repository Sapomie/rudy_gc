package modelx

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
)

const (
	MovieHasNoChinese = 1 + iota
	MovieHasChinese
	TranslationError
	TranslationSensitiveWord
)

const (
	MovieHasNoLocalCover = 1 + iota
	MovieHasLocalCover
	MovieCoverWrongUrl
)

const (
	MovieNoNeedDownload = 1 + iota
	MovieNeedDownload
)
const (
	MovieNotOwned = 1 + iota
	MovieOwned
	MovieOwnedAndSub
	MovieIsRemoved
)

func (m *defaultAMovieModel) FindMoviesNeedRefresh(ctx context.Context) ([]*AMovie, error) {
	const (
		RefreshInterval   = 40 * 24 * 60 * 60 // 40 days in seconds
		MinAgeRequirement = 10 * 24 * 60 * 60 // 10 days in seconds
	)

	query, args, err := squirrel.Select("*").
		From(m.tableName()).
		Where(
			"(last_query_jav_time - releasing_date < ?) AND (UNIX_TIMESTAMP(NOW()) - releasing_date > ?)",
			RefreshInterval, MinAgeRequirement,
		).ToSql()
	if err != nil {
		return nil, fmt.Errorf("constructing SQL query failed: %w", err)
	}

	var movies []*AMovie
	if err := m.QueryRowsNoCacheCtx(ctx, &movies, query, args...); err != nil {
		return nil, err
	}

	return movies, nil
}

func (m *defaultAMovieModel) FindMoviesHasNoChinese(ctx context.Context) ([]*AMovie, error) {
	// 使用 squirrel 构建 SQL 查询
	query, args, err := squirrel.Select("*").
		From(m.tableName()).
		Where("`has_chinese` = ?", MovieHasNoChinese).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var movies []*AMovie
	// 执行查询
	if err := m.QueryRowsNoCacheCtx(ctx, &movies, query, args...); err != nil {
		return nil, err
	}

	return movies, nil
}

func (m *defaultAMovieModel) FindMoviesByName(ctx context.Context, name string) ([]*AMovie, error) {
	// 使用 squirrel 构建 SQL 查询
	query, args, err := squirrel.Select("*").
		From(m.tableName()).
		OrderBy("viewers_number_watched desc").
		Where("`name` = ?", name).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var movies []*AMovie
	// 执行查询
	if err := m.QueryRowsNoCacheCtx(ctx, &movies, query, args...); err != nil {
		return nil, err
	}

	return movies, nil
}

func (m *defaultAMovieModel) FindMoviesByEncodeName(ctx context.Context, encodeName string) ([]*AMovie, error) {
	// 使用 squirrel 构建 SQL 查询
	query, args, err := squirrel.Select("*").
		From(m.tableName()).
		OrderBy("viewers_number_watched desc").
		Where("`encode_name` = ?", encodeName).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var movies []*AMovie
	// 执行查询
	if err := m.QueryRowsNoCacheCtx(ctx, &movies, query, args...); err != nil {
		return nil, err
	}

	return movies, nil
}

func (m *defaultAMovieModel) FindMoviesHasNoCover(ctx context.Context) ([]*AMovie, error) {
	// 使用 squirrel 构建 SQL 查询
	query, args, err := squirrel.Select("*").
		From(m.tableName()).
		Where("`has_downloaded_cover` = ?", MovieHasNoLocalCover).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var movies []*AMovie
	// 执行查询
	if err := m.QueryRowsNoCacheCtx(ctx, &movies, query, args...); err != nil {
		return nil, err
	}

	return movies, nil
}

func (m *defaultAMovieModel) FindMoviesAll(ctx context.Context, limit uint64) ([]*AMovie, error) {
	// 使用 squirrel 构建 SQL 查询
	query, args, err := squirrel.Select("*").
		From(m.tableName()).Limit(limit).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var movies []*AMovie
	// 执行查询
	if err := m.QueryRowsNoCacheCtx(ctx, &movies, query, args...); err != nil {
		return nil, err
	}

	return movies, nil
}
