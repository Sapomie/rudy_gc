package moviex

import (
	"context"
	"errors"
	"fmt"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WMediaModel = (*customWMediaModel)(nil)

type (
	WMediaBirthBucketCalc struct {
		MediaCount      int64 `db:"media_count"`
		RemovedCount    int64 `db:"removed_count"`
		SizeBytes       int64 `db:"size_bytes"`
		HasSubCount     int64 `db:"has_sub_count"`
		LatestBirthTime int64 `db:"latest_birth_time"`
	}

	WMediaBirthTopCalc struct {
		AggKey     string `db:"agg_key"`
		AggId      int64  `db:"agg_id"`
		AggName    string `db:"agg_name"`
		MediaCount int64  `db:"media_count"`
		SizeBytes  int64  `db:"size_bytes"`
	}

	LegacyFilm struct {
		Id            int64   `db:"id"`
		MovieJavId    string  `db:"movie_jav_id"`
		MovieName     string  `db:"movie_name"`
		FileName      string  `db:"file_name"`
		DirectoryId   int64   `db:"directory_id"`
		RootDir       string  `db:"root_dir"`
		FullDir       string  `db:"full_dir"`
		Dir1Id        int64   `db:"dir1_id"`
		Dir2Id        int64   `db:"dir2_id"`
		Dir3Id        int64   `db:"dir3_id"`
		Dir4Id        int64   `db:"dir4_id"`
		Alias         string  `db:"alias"`
		Size          int64   `db:"size"`
		Width         int64   `db:"width"`
		Height        int64   `db:"height"`
		BitRate       int64   `db:"bit_rate"`
		Duration      int64   `db:"duration"`
		FrameAverage  float64 `db:"frame_average"`
		HasSub        int64   `db:"has_sub"`
		SelfMake      int64   `db:"self_make"`
		HasMask       int64   `db:"has_mask"`
		NeedScanMeta  int64   `db:"need_scan_meta"`
		IsRemoved     int64   `db:"is_removed"`
		RemoveTime    int64   `db:"remove_time"`
		ScTimes       int64   `db:"sc_times"`
		ComeTimes     int64   `db:"come_times"`
		LastScTime    int64   `db:"last_sc_time"`
		BirthTime     int64   `db:"birth_time"`
		ReleasingDate int64   `db:"releasing_date"`
		CreatedOn     int64   `db:"created_on"`
		UpdatedOn     int64   `db:"updated_on"`
	}

	NativeMediaListRow struct {
		Id            int64   `db:"id"`
		MovieJavId    string  `db:"movie_jav_id"`
		MovieName     string  `db:"movie_name"`
		FileName      string  `db:"file_name"`
		DirectoryId   int64   `db:"directory_id"`
		RootDir       string  `db:"root_dir"`
		FullDir       string  `db:"full_dir"`
		Alias         string  `db:"alias"`
		Size          int64   `db:"size"`
		Width         int64   `db:"width"`
		Height        int64   `db:"height"`
		BitRate       int64   `db:"bit_rate"`
		Duration      int64   `db:"duration"`
		FrameAverage  float64 `db:"frame_average"`
		HasSub        int64   `db:"has_sub"`
		SelfMake      int64   `db:"self_make"`
		HasMask       int64   `db:"has_mask"`
		NeedScanMeta  int64   `db:"need_scan_meta"`
		IsRemoved     int64   `db:"is_removed"`
		RemoveTime    int64   `db:"remove_time"`
		ScTimes       int64   `db:"sc_times"`
		ComeTimes     int64   `db:"come_times"`
		LastScTime    int64   `db:"last_sc_time"`
		BirthTime     int64   `db:"birth_time"`
		ReleasingDate int64   `db:"releasing_date"`
		CreatedOn     int64   `db:"created_on"`
		UpdatedOn     int64   `db:"updated_on"`
	}

	// WMediaModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWMediaModel.
	WMediaModel interface {
		wMediaModel
		CalcBirthBucket(ctx context.Context, start, end int64) (*WMediaBirthBucketCalc, error)
		CalcTopCastsByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error)
		CalcTopDirectorsByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error)
		CalcTopLabelsByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error)
		CalcTopPrefixesByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error)
		ListByMovieJavIds(ctx context.Context, movieJavIds []string) ([]*WMedia, error)
		FindOneLegacyFilmByID(ctx context.Context, id int64) (*LegacyFilm, error)
		FindOneLegacyFilmByMovieJavId(ctx context.Context, movieJavId string) (*LegacyFilm, error)
		FindOneLegacyFilmByMovieName(ctx context.Context, movieName string) (*LegacyFilm, error)
		FindAllLegacyFilms(ctx context.Context, removedStatus int64) ([]*LegacyFilm, error)
		ListLegacyFilmsByMovieJavIds(ctx context.Context, movieJavIds []string) ([]*LegacyFilm, error)
		ListLegacyFilmsPage(ctx context.Context, offset, limit int64, orderBy string, filter types.FilmListFilter) ([]*LegacyFilm, error)
		CountLegacyFilms(ctx context.Context, filter types.FilmListFilter) (int64, error)
		ListLegacyFilmsByDirectoryIDs(ctx context.Context, dirIDs []int64, page, size int, orderBy string) (all, paged []*LegacyFilm, total int64, err error)
		ListNativeMediaPage(ctx context.Context, offset, limit int64, orderBy string, filter types.MediaListFilter) ([]*NativeMediaListRow, error)
		CountNativeMedia(ctx context.Context, filter types.MediaListFilter) (int64, error)
		FindOneByMovieJavIdSourceType(ctx context.Context, movieJavId string, sourceType int64) (*WMedia, error)
		FindOneByMovieNameSourceType(ctx context.Context, movieName string, sourceType int64) (*WMedia, error)
		FindOneByFileNameSourceType(ctx context.Context, fileName string, sourceType int64) (*WMedia, error)
		ListByDirectoryIDs(ctx context.Context, dirIDs []int64, page, size int, orderBy string) (all, paged []*WMedia, total int64, err error)
		ListDistinctBirthDays(ctx context.Context) ([]int64, error)
		ListByFullDirPrefixes(ctx context.Context, prefixes []string) ([]*WMedia, error)
		TableName() string
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
	}

	customWMediaModel struct {
		*defaultWMediaModel
	}
)

// NewWMediaModel returns a model for the database table.
func NewWMediaModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WMediaModel {
	return &customWMediaModel{
		defaultWMediaModel: newWMediaModel(conn, c, opts...),
	}
}

func (m *customWMediaModel) TableName() string { return m.table }

func (m *customWMediaModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customWMediaModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customWMediaModel) FindOneByMovieJavId(ctx context.Context, movieJavId string) (*WMedia, error) {
	return m.FindOneByMovieJavIdSourceType(ctx, movieJavId, consts.WMediaSourceNative)
}

func (m *customWMediaModel) FindOneByMovieName(ctx context.Context, movieName string) (*WMedia, error) {
	return m.FindOneByMovieNameSourceType(ctx, movieName, consts.WMediaSourceNative)
}

func (m *customWMediaModel) FindOneByFileName(ctx context.Context, fileName string) (*WMedia, error) {
	return m.FindOneByFileNameSourceType(ctx, fileName, consts.WMediaSourceNative)
}

func (m *customWMediaModel) FindOneByMovieJavIdSourceType(ctx context.Context, movieJavId string, sourceType int64) (*WMedia, error) {
	return m.findOneByFieldAndSourceType(ctx, "movie_jav_id", movieJavId, sourceType)
}

func (m *customWMediaModel) FindOneByMovieNameSourceType(ctx context.Context, movieName string, sourceType int64) (*WMedia, error) {
	return m.findOneByFieldAndSourceType(ctx, "movie_name", movieName, sourceType)
}

func (m *customWMediaModel) FindOneByFileNameSourceType(ctx context.Context, fileName string, sourceType int64) (*WMedia, error) {
	return m.findOneByFieldAndSourceType(ctx, "file_name", fileName, sourceType)
}

func (m *customWMediaModel) ListByMovieJavIds(ctx context.Context, movieJavIds []string) ([]*WMedia, error) {
	if len(movieJavIds) == 0 {
		return []*WMedia{}, nil
	}

	query, args, err := squirrel.
		Select(wMediaRows).
		From(m.table).
		Where(squirrel.Eq{"movie_jav_id": movieJavIds}).
		Where(squirrel.Eq{"source_type": consts.WMediaSourceNative}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*WMedia
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*WMedia{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func legacyFilmColumns(alias string) []string {
	return []string{
		alias + ".id AS id",
		alias + ".movie_jav_id AS movie_jav_id",
		alias + ".movie_name AS movie_name",
		alias + ".file_name AS file_name",
		alias + ".directory_id AS directory_id",
		alias + ".root_dir AS root_dir",
		alias + ".full_dir AS full_dir",
		"COALESCE(vd1.id, 0) AS dir1_id",
		"COALESCE(vd2.id, 0) AS dir2_id",
		"COALESCE(vd3.id, 0) AS dir3_id",
		"COALESCE(vd4.id, 0) AS dir4_id",
		alias + ".alias AS alias",
		alias + ".size AS size",
		alias + ".width AS width",
		alias + ".height AS height",
		alias + ".bit_rate AS bit_rate",
		alias + ".duration AS duration",
		alias + ".frame_average AS frame_average",
		alias + ".has_sub AS has_sub",
		alias + ".self_make AS self_make",
		alias + ".has_mask AS has_mask",
		alias + ".need_scan_meta AS need_scan_meta",
		alias + ".is_removed AS is_removed",
		alias + ".remove_time AS remove_time",
		"COALESCE(gss.sc_times, 0) AS sc_times",
		"COALESCE(gss.come_times, 0) AS come_times",
		"COALESCE(gss.last_sc_time, 0) AS last_sc_time",
		alias + ".birth_time AS birth_time",
		alias + ".releasing_date AS releasing_date",
		alias + ".created_on AS created_on",
		alias + ".updated_on AS updated_on",
	}
}

func legacyFilmBuilder(tableName string, alias string, columns ...string) squirrel.SelectBuilder {
	if len(columns) == 0 {
		columns = legacyFilmColumns(alias)
	}
	return squirrel.
		Select(columns...).
		From(tableName + " " + alias).
		LeftJoin("`g_sc_stat` gss ON gss.movie_jav_id = " + alias + ".movie_jav_id").
		LeftJoin("`w_folder` vd1 ON vd1.id = " + alias + ".directory_id AND vd1.source_type = " + fmt.Sprintf("%d", consts.WFolderSourceLegacyVFilm)).
		LeftJoin("`w_folder` vd2 ON vd2.id = vd1.parent_id AND vd2.source_type = " + fmt.Sprintf("%d", consts.WFolderSourceLegacyVFilm)).
		LeftJoin("`w_folder` vd3 ON vd3.id = vd2.parent_id AND vd3.source_type = " + fmt.Sprintf("%d", consts.WFolderSourceLegacyVFilm)).
		LeftJoin("`w_folder` vd4 ON vd4.id = vd3.parent_id AND vd4.source_type = " + fmt.Sprintf("%d", consts.WFolderSourceLegacyVFilm)).
		Where(squirrel.Eq{alias + ".source_type": consts.WMediaSourceLegacyVFilm})
}

func (m *customWMediaModel) findOneLegacyFilmByField(ctx context.Context, fieldName string, fieldValue string) (*LegacyFilm, error) {
	fieldValue = strings.TrimSpace(fieldValue)
	if fieldValue == "" {
		return nil, ErrNotFound
	}

	query, args, err := legacyFilmBuilder(m.table, "f").
		Where(squirrel.Eq{"f." + fieldName: fieldValue}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}

	var row LegacyFilm
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *customWMediaModel) FindOneLegacyFilmByID(ctx context.Context, id int64) (*LegacyFilm, error) {
	query, args, err := legacyFilmBuilder(m.table, "f").
		Where(squirrel.Eq{"f.id": id}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}

	var row LegacyFilm
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *customWMediaModel) FindOneLegacyFilmByMovieJavId(ctx context.Context, movieJavId string) (*LegacyFilm, error) {
	return m.findOneLegacyFilmByField(ctx, "movie_jav_id", movieJavId)
}

func (m *customWMediaModel) FindOneLegacyFilmByMovieName(ctx context.Context, movieName string) (*LegacyFilm, error) {
	return m.findOneLegacyFilmByField(ctx, "movie_name", movieName)
}

func (m *customWMediaModel) FindAllLegacyFilms(ctx context.Context, removedStatus int64) ([]*LegacyFilm, error) {
	builder := legacyFilmBuilder(m.table, "f")
	if removedStatus > 0 {
		builder = builder.Where(squirrel.Eq{"f.is_removed": removedStatus})
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*LegacyFilm
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*LegacyFilm{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customWMediaModel) ListLegacyFilmsByMovieJavIds(ctx context.Context, movieJavIds []string) ([]*LegacyFilm, error) {
	if len(movieJavIds) == 0 {
		return []*LegacyFilm{}, nil
	}

	query, args, err := legacyFilmBuilder(m.table, "f").
		Where(squirrel.Eq{"f.movie_jav_id": movieJavIds}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*LegacyFilm
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*LegacyFilm{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func applyLegacyFilmListFilter(builder squirrel.SelectBuilder, filter types.FilmListFilter) squirrel.SelectBuilder {
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
	if filter.HasScTimesMin {
		builder = builder.Where(squirrel.Expr("COALESCE(gss.sc_times, 0) >= ?", filter.ScTimesMin))
	}
	if filter.HasScTimesMax {
		builder = builder.Where(squirrel.Expr("COALESCE(gss.sc_times, 0) <= ?", filter.ScTimesMax))
	}
	if filter.HasLastScFrom {
		builder = builder.Where(squirrel.Expr("COALESCE(gss.last_sc_time, 0) >= ?", filter.LastScFrom))
	}
	if filter.HasLastScTo {
		builder = builder.Where(squirrel.Expr("COALESCE(gss.last_sc_time, 0) <= ?", filter.LastScTo))
	}
	return builder
}

func (m *customWMediaModel) ListLegacyFilmsPage(ctx context.Context, offset, limit int64, orderBy string, filter types.FilmListFilter) ([]*LegacyFilm, error) {
	orderParts := splitOrderClause(orderBy)
	if len(orderParts) == 0 {
		orderParts = []string{"f.birth_time DESC", "f.movie_name DESC"}
	}

	builder := applyLegacyFilmListFilter(legacyFilmBuilder(m.table, "f"), filter).OrderBy(orderParts...)
	if limit > 0 {
		builder = builder.Limit(uint64(limit))
	}
	if offset > 0 {
		builder = builder.Offset(uint64(offset))
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*LegacyFilm
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*LegacyFilm{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customWMediaModel) CountLegacyFilms(ctx context.Context, filter types.FilmListFilter) (int64, error) {
	query, args, err := applyLegacyFilmListFilter(legacyFilmBuilder(m.table, "f", "COUNT(1)"), filter).ToSql()
	if err != nil {
		return 0, err
	}

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return total, nil
}

func (m *customWMediaModel) ListLegacyFilmsByDirectoryIDs(ctx context.Context, dirIDs []int64, page, size int, orderBy string) (all, paged []*LegacyFilm, total int64, err error) {
	if len(dirIDs) == 0 {
		return []*LegacyFilm{}, []*LegacyFilm{}, 0, nil
	}

	orderParts := splitOrderClause(orderBy)
	if len(orderParts) == 0 {
		orderParts = []string{"f.birth_time DESC"}
	}

	query, args, err := legacyFilmBuilder(m.table, "f").
		Where(squirrel.Eq{"f.is_removed": consts.FilmIsNotRemoved}).
		Where(squirrel.Eq{"f.directory_id": dirIDs}).
		OrderBy(orderParts...).
		ToSql()
	if err != nil {
		return nil, nil, 0, err
	}

	var rows []*LegacyFilm
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*LegacyFilm{}, []*LegacyFilm{}, 0, nil
		}
		return nil, nil, 0, err
	}

	total = int64(len(rows))
	if total == 0 {
		return []*LegacyFilm{}, []*LegacyFilm{}, 0, nil
	}
	start := (page - 1) * size
	if start >= len(rows) {
		return rows, []*LegacyFilm{}, total, nil
	}
	end := start + size
	if end > len(rows) {
		end = len(rows)
	}
	paged = rows[start:end]
	return rows, paged, total, nil
}

func nativeMediaListColumns(alias string) []string {
	return []string{
		alias + ".id AS id",
		alias + ".movie_jav_id AS movie_jav_id",
		alias + ".movie_name AS movie_name",
		alias + ".file_name AS file_name",
		alias + ".directory_id AS directory_id",
		alias + ".root_dir AS root_dir",
		alias + ".full_dir AS full_dir",
		alias + ".alias AS alias",
		alias + ".size AS size",
		alias + ".width AS width",
		alias + ".height AS height",
		alias + ".bit_rate AS bit_rate",
		alias + ".duration AS duration",
		alias + ".frame_average AS frame_average",
		alias + ".has_sub AS has_sub",
		alias + ".self_make AS self_make",
		alias + ".has_mask AS has_mask",
		alias + ".need_scan_meta AS need_scan_meta",
		alias + ".is_removed AS is_removed",
		alias + ".remove_time AS remove_time",
		"COALESCE(gss.sc_times, 0) AS sc_times",
		"COALESCE(gss.come_times, 0) AS come_times",
		"COALESCE(gss.last_sc_time, 0) AS last_sc_time",
		alias + ".birth_time AS birth_time",
		alias + ".releasing_date AS releasing_date",
		alias + ".created_on AS created_on",
		alias + ".updated_on AS updated_on",
	}
}

func nativeMediaListBuilder(tableName string, alias string, columns ...string) squirrel.SelectBuilder {
	if len(columns) == 0 {
		columns = nativeMediaListColumns(alias)
	}
	return squirrel.
		Select(columns...).
		From(tableName + " " + alias).
		LeftJoin("`g_sc_stat` gss ON gss.movie_jav_id = " + alias + ".movie_jav_id").
		Where(squirrel.Eq{alias + ".source_type": consts.WMediaSourceNative})
}

func applyNativeMediaListFilter(builder squirrel.SelectBuilder, filter types.MediaListFilter) squirrel.SelectBuilder {
	builder = builder.Where(squirrel.Eq{"wm.is_removed": consts.FilmIsNotRemoved})

	if kw := strings.TrimSpace(filter.MovieNameKeyword); kw != "" {
		builder = builder.Where(squirrel.Like{"wm.movie_name": "%" + kw + "%"})
	}
	if filter.HasSizeMin {
		builder = builder.Where(squirrel.GtOrEq{"wm.size": filter.SizeMin})
	}
	if filter.HasSizeMax {
		builder = builder.Where(squirrel.LtOrEq{"wm.size": filter.SizeMax})
	}
	if filter.HasHeightMin {
		builder = builder.Where(squirrel.GtOrEq{"wm.height": filter.HeightMin})
	}
	if filter.HasHeightMax {
		builder = builder.Where(squirrel.LtOrEq{"wm.height": filter.HeightMax})
	}
	if filter.HasDurationMin {
		builder = builder.Where(squirrel.GtOrEq{"wm.duration": filter.DurationMin})
	}
	if filter.HasDurationMax {
		builder = builder.Where(squirrel.LtOrEq{"wm.duration": filter.DurationMax})
	}
	if filter.HasBitRateMin {
		builder = builder.Where(squirrel.GtOrEq{"wm.bit_rate": filter.BitRateMin})
	}
	if filter.HasBitRateMax {
		builder = builder.Where(squirrel.LtOrEq{"wm.bit_rate": filter.BitRateMax})
	}
	if filter.HasFrameAverageMin {
		builder = builder.Where(squirrel.GtOrEq{"wm.frame_average": filter.FrameAverageMin})
	}
	if filter.HasFrameAverageMax {
		builder = builder.Where(squirrel.LtOrEq{"wm.frame_average": filter.FrameAverageMax})
	}
	if filter.HasSelfMake {
		builder = builder.Where(squirrel.Eq{"wm.self_make": filter.SelfMake})
	}
	if filter.HasHasMask {
		builder = builder.Where(squirrel.Eq{"wm.has_mask": filter.HasMask})
	}
	if filter.HasBirthTimeFrom {
		builder = builder.Where(squirrel.GtOrEq{"wm.birth_time": filter.BirthTimeFrom})
	}
	if filter.HasBirthTimeTo {
		builder = builder.Where(squirrel.LtOrEq{"wm.birth_time": filter.BirthTimeTo})
	}
	if filter.HasReleasingDateFrom {
		builder = builder.Where(squirrel.GtOrEq{"wm.releasing_date": filter.ReleasingDateFrom})
	}
	if filter.HasReleasingDateTo {
		builder = builder.Where(squirrel.LtOrEq{"wm.releasing_date": filter.ReleasingDateTo})
	}
	if filter.HasScTimesMin {
		builder = builder.Where(squirrel.Expr("COALESCE(gss.sc_times, 0) >= ?", filter.ScTimesMin))
	}
	if filter.HasScTimesMax {
		builder = builder.Where(squirrel.Expr("COALESCE(gss.sc_times, 0) <= ?", filter.ScTimesMax))
	}
	if filter.HasLastScFrom {
		builder = builder.Where(squirrel.Expr("COALESCE(gss.last_sc_time, 0) >= ?", filter.LastScFrom))
	}
	if filter.HasLastScTo {
		builder = builder.Where(squirrel.Expr("COALESCE(gss.last_sc_time, 0) <= ?", filter.LastScTo))
	}
	return builder
}

func (m *customWMediaModel) ListNativeMediaPage(ctx context.Context, offset, limit int64, orderBy string, filter types.MediaListFilter) ([]*NativeMediaListRow, error) {
	orderParts := splitOrderClause(orderBy)
	if len(orderParts) == 0 {
		orderParts = []string{"wm.birth_time DESC", "wm.movie_name DESC"}
	}

	builder := applyNativeMediaListFilter(nativeMediaListBuilder(m.table, "wm"), filter).OrderBy(orderParts...)
	if limit > 0 {
		builder = builder.Limit(uint64(limit))
	}
	if offset > 0 {
		builder = builder.Offset(uint64(offset))
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*NativeMediaListRow
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*NativeMediaListRow{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customWMediaModel) CountNativeMedia(ctx context.Context, filter types.MediaListFilter) (int64, error) {
	query, args, err := applyNativeMediaListFilter(nativeMediaListBuilder(m.table, "wm", "COUNT(1)"), filter).ToSql()
	if err != nil {
		return 0, err
	}

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return total, nil
}

func (m *customWMediaModel) findOneByFieldAndSourceType(ctx context.Context, fieldName string, fieldValue string, sourceType int64) (*WMedia, error) {
	fieldValue = strings.TrimSpace(fieldValue)
	if fieldValue == "" || sourceType <= 0 {
		return nil, ErrNotFound
	}

	query := `
SELECT ` + wMediaRows + `
FROM ` + m.table + `
WHERE ` + "`" + fieldName + "` = ?" + ` AND source_type = ?
LIMIT 1`

	var row WMedia
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, fieldValue, sourceType); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *customWMediaModel) ListDistinctBirthDays(ctx context.Context) ([]int64, error) {
	const query = `
SELECT birth_time
FROM w_media
WHERE source_type = ?
  AND birth_time > 0
ORDER BY birth_time ASC`

	var birthTimes []int64
	if err := m.QueryRowsNoCacheCtx(ctx, &birthTimes, query, consts.WMediaSourceNative); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []int64{}, nil
		}
		return nil, err
	}

	seen := make(map[int64]struct{}, len(birthTimes))
	days := make([]int64, 0, len(birthTimes))
	for _, birthTime := range birthTimes {
		bucketDay := normalizeLocalBirthBucketDay(birthTime)
		if bucketDay <= 0 {
			continue
		}
		if _, ok := seen[bucketDay]; ok {
			continue
		}
		seen[bucketDay] = struct{}{}
		days = append(days, bucketDay)
	}
	return days, nil
}

func (m *customWMediaModel) CalcBirthBucket(ctx context.Context, start, end int64) (*WMediaBirthBucketCalc, error) {
	const query = `
SELECT
	COALESCE(SUM(CASE WHEN is_removed = ? THEN 1 ELSE 0 END), 0) AS media_count,
	COALESCE(SUM(CASE WHEN is_removed = ? THEN 1 ELSE 0 END), 0) AS removed_count,
	COALESCE(SUM(CASE WHEN is_removed = ? THEN size ELSE 0 END), 0) AS size_bytes,
	COALESCE(SUM(CASE WHEN is_removed = ? AND has_sub = ? THEN 1 ELSE 0 END), 0) AS has_sub_count,
	COALESCE(MAX(CASE WHEN is_removed = ? THEN birth_time ELSE 0 END), 0) AS latest_birth_time
FROM w_media
WHERE source_type = ?
  AND birth_time >= ?
  AND birth_time <= ?`

	var row WMediaBirthBucketCalc
	if err := m.QueryRowNoCacheCtx(
		ctx,
		&row,
		query,
		consts.FilmIsNotRemoved,
		consts.FilmIsRemoved,
		consts.FilmIsNotRemoved,
		consts.FilmIsNotRemoved,
		consts.FilmHasSub,
		consts.FilmIsNotRemoved,
		consts.WMediaSourceNative,
		start,
		end,
	); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return &WMediaBirthBucketCalc{}, nil
		}
		return nil, err
	}
	return &row, nil
}

func (m *customWMediaModel) CalcTopCastsByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error) {
	if limit <= 0 {
		limit = 30
	}
	const query = `
SELECT
	src.agg_key AS agg_key,
	src.agg_id AS agg_id,
	src.agg_name AS agg_name,
	COUNT(*) AS media_count,
	COALESCE(SUM(src.size), 0) AS size_bytes
FROM (
	SELECT DISTINCT
		wm.movie_jav_id,
		wm.size,
		COALESCE(ac.person_id, 0) AS agg_id,
		CASE
			WHEN TRIM(COALESCE(cp.chinese, '')) <> '' THEN TRIM(cp.chinese)
			WHEN TRIM(COALESCE(cp.name, '')) <> '' THEN TRIM(cp.name)
			ELSE TRIM(ac.name)
		END AS agg_name,
		CONCAT(COALESCE(ac.person_id, 0), ':', TRIM(COALESCE(ac.name, ''))) AS agg_key
	FROM w_media wm
	INNER JOIN amr_movie_cast amr ON amr.movie_jav_id = wm.movie_jav_id
	INNER JOIN am_cast ac ON ac.id = amr.cast_id
	LEFT JOIN c_person cp ON cp.id = ac.person_id
	WHERE wm.source_type = ?
	  AND wm.is_removed = ?
	  AND wm.birth_time >= ?
	  AND wm.birth_time <= ?
) src
WHERE TRIM(COALESCE(src.agg_name, '')) <> ''
GROUP BY src.agg_key, src.agg_id, src.agg_name
ORDER BY media_count DESC, size_bytes DESC, agg_name ASC
LIMIT ?`

	var rows []*WMediaBirthTopCalc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, consts.WMediaSourceNative, consts.FilmIsNotRemoved, start, end, limit); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*WMediaBirthTopCalc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customWMediaModel) CalcTopDirectorsByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error) {
	return m.calcTopByMovieField(ctx, start, end, limit, "director")
}

func (m *customWMediaModel) CalcTopLabelsByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error) {
	return m.calcTopByMovieField(ctx, start, end, limit, "label")
}

func (m *customWMediaModel) CalcTopPrefixesByBirthRange(ctx context.Context, start, end int64, limit int) ([]*WMediaBirthTopCalc, error) {
	return m.calcTopByMovieField(ctx, start, end, limit, "prefix")
}

func (m *customWMediaModel) calcTopByMovieField(ctx context.Context, start, end int64, limit int, aggType string) ([]*WMediaBirthTopCalc, error) {
	if limit <= 0 {
		limit = 30
	}

	type joinInfo struct {
		joinTable string
		joinAlias string
		nameExpr  string
		idExpr    string
		joinOn    string
	}

	infoMap := map[string]joinInfo{
		"director": {
			joinTable: "am_director",
			joinAlias: "ad",
			nameExpr:  "TRIM(COALESCE(ad.name, ''))",
			idExpr:    "COALESCE(ad.id, 0)",
			joinOn:    "ad.id = am.director_id",
		},
		"label": {
			joinTable: "am_label",
			joinAlias: "al",
			nameExpr:  "TRIM(COALESCE(al.name, ''))",
			idExpr:    "COALESCE(al.id, 0)",
			joinOn:    "al.id = am.label_id",
		},
		"prefix": {
			joinTable: "am_prefix",
			joinAlias: "ap",
			nameExpr:  "TRIM(COALESCE(ap.name, ''))",
			idExpr:    "COALESCE(ap.id, 0)",
			joinOn:    "ap.id = am.prefix_id",
		},
	}

	info, ok := infoMap[aggType]
	if !ok {
		return []*WMediaBirthTopCalc{}, nil
	}

	query := `
SELECT
	CONCAT(` + info.idExpr + `, ':', ` + info.nameExpr + `) AS agg_key,
	` + info.idExpr + ` AS agg_id,
	` + info.nameExpr + ` AS agg_name,
	COUNT(*) AS media_count,
	COALESCE(SUM(wm.size), 0) AS size_bytes
FROM w_media wm
INNER JOIN a_movie am ON am.jav_id = wm.movie_jav_id
LEFT JOIN ` + info.joinTable + ` ` + info.joinAlias + ` ON ` + info.joinOn + `
WHERE wm.source_type = ?
  AND wm.is_removed = ?
  AND wm.birth_time >= ?
  AND wm.birth_time <= ?
  AND ` + info.nameExpr + ` <> ''
GROUP BY agg_key, agg_id, agg_name
ORDER BY media_count DESC, size_bytes DESC, agg_name ASC
LIMIT ?`

	var rows []*WMediaBirthTopCalc
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, consts.WMediaSourceNative, consts.FilmIsNotRemoved, start, end, limit); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*WMediaBirthTopCalc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func normalizeLocalBirthBucketDay(birthTime int64) int64 {
	if birthTime <= 0 {
		return 0
	}
	t := time.Unix(birthTime, 0).In(time.Local)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Unix()
}

func (m *customWMediaModel) ListByDirectoryIDs(ctx context.Context, dirIDs []int64, page, size int, orderBy string) (all, paged []*WMedia, total int64, err error) {
	if len(dirIDs) == 0 {
		return []*WMedia{}, []*WMedia{}, 0, nil
	}

	orderParts := splitOrderClause(mapWMediaOrderBy(orderBy))
	if len(orderParts) == 0 {
		orderParts = []string{"wm.birth_time DESC", "wm.movie_name DESC"}
	}

	qAll, argsAll, err := squirrel.
		Select("wm.*").
		From(m.table + " AS wm").
		LeftJoin(buildLegacyWMediaJoin("`w_media`", "vf", "wm.movie_jav_id")).
		LeftJoin("`g_sc_stat` gss ON gss.movie_jav_id = wm.movie_jav_id").
		Where(squirrel.Eq{"wm.source_type": consts.WMediaSourceNative}).
		Where(squirrel.Eq{"wm.is_removed": consts.FilmIsNotRemoved}).
		Where(squirrel.Eq{"wm.directory_id": dirIDs}).
		OrderBy(orderParts...).
		ToSql()
	if err != nil {
		return nil, nil, 0, err
	}

	var rows []*WMedia
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, qAll, argsAll...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*WMedia{}, []*WMedia{}, 0, nil
		}
		return nil, nil, 0, err
	}

	total = int64(len(rows))
	if total == 0 {
		return []*WMedia{}, []*WMedia{}, 0, nil
	}

	start := (page - 1) * size
	if start >= len(rows) {
		return rows, []*WMedia{}, total, nil
	}

	end := start + size
	if end > len(rows) {
		end = len(rows)
	}
	paged = rows[start:end]
	return rows, paged, total, nil
}

func (m *customWMediaModel) ListByFullDirPrefixes(ctx context.Context, prefixes []string) ([]*WMedia, error) {
	if len(prefixes) == 0 {
		return []*WMedia{}, nil
	}

	conditions := make(squirrel.Or, 0, len(prefixes)*2)
	for _, prefix := range prefixes {
		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			continue
		}
		conditions = append(conditions,
			squirrel.Eq{"wm.full_dir": prefix},
			squirrel.Like{"wm.full_dir": prefix + "/%"},
		)
	}
	if len(conditions) == 0 {
		return []*WMedia{}, nil
	}

	query, args, err := squirrel.
		Select(wMediaRows).
		From(m.table + " AS wm").
		Where(squirrel.Eq{"wm.source_type": consts.WMediaSourceNative}).
		Where(conditions).
		OrderBy("wm.id ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*WMedia
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*WMedia{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func mapWMediaOrderBy(orderBy string) string {
	order := "wm.birth_time DESC,wm.movie_name DESC"
	switch orderBy {
	case consts.OrderByBirthTime:
		order = "CASE WHEN vf.birth_time IS NULL THEN wm.birth_time ELSE vf.birth_time END DESC,wm.movie_name DESC"
	case consts.OrderByMediaBirthTime:
		order = "wm.birth_time DESC,wm.movie_name DESC"
	case consts.OrderByScTimes:
		order = "COALESCE(gss.sc_times, 0) DESC,COALESCE(gss.last_sc_time, 0) DESC,wm.movie_name DESC"
	case consts.OrderByComeTimes:
		order = "COALESCE(gss.come_times, 0) DESC,COALESCE(gss.last_sc_time, 0) DESC,wm.movie_name DESC"
	case consts.OrderByLastScTime:
		order = "COALESCE(gss.last_sc_time, 0) DESC,wm.movie_name DESC"
	case consts.OrderByReleasingDate:
		order = "wm.releasing_date DESC,wm.movie_name DESC"
	}
	return order
}
