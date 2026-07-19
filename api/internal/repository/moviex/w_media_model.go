package moviex

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
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

	MissingScNativeMediaRow struct {
		MovieJavId         string `db:"movie_jav_id"`
		MovieName          string `db:"movie_name"`
		ScTimes            int64  `db:"sc_times"`
		LastSelectedScTime int64  `db:"last_selected_sc_time"`
		ReleasingDate      int64  `db:"releasing_date"`
		MediaBirthTime     int64  `db:"media_birth_time"`
	}

	NativeMediaListRow struct {
		Id                 int64   `db:"id"`
		MovieJavId         string  `db:"movie_jav_id"`
		MovieName          string  `db:"movie_name"`
		FileName           string  `db:"file_name"`
		DirectoryId        int64   `db:"directory_id"`
		RootDir            string  `db:"root_dir"`
		FullDir            string  `db:"full_dir"`
		Alias              string  `db:"alias"`
		Size               int64   `db:"size"`
		Width              int64   `db:"width"`
		Height             int64   `db:"height"`
		BitRate            int64   `db:"bit_rate"`
		Duration           int64   `db:"duration"`
		FrameAverage       float64 `db:"frame_average"`
		HasSub             int64   `db:"has_sub"`
		SelfMake           int64   `db:"self_make"`
		HasMask            int64   `db:"has_mask"`
		NeedScanMeta       int64   `db:"need_scan_meta"`
		IsRemoved          int64   `db:"is_removed"`
		RemoveTime         int64   `db:"remove_time"`
		ScTimes            int64   `db:"sc_times"`
		ComeTimes          int64   `db:"come_times"`
		LastSelectedScTime int64   `db:"last_selected_sc_time"`
		BirthTime          int64   `db:"birth_time"`
		ReleasingDate      int64   `db:"releasing_date"`
		CreatedOn          int64   `db:"created_on"`
		UpdatedOn          int64   `db:"updated_on"`
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
		ListNativeMediaPage(ctx context.Context, offset, limit int64, orderBy string, filter types.MediaListFilter) ([]*NativeMediaListRow, error)
		ListNativeMediaPageWithTotal(ctx context.Context, offset, limit int64, orderBy string, filter types.MediaListFilter) ([]*NativeMediaListRow, int64, error)
		CountNativeMedia(ctx context.Context, filter types.MediaListFilter) (int64, error)
		FindOneByMovieJavIdSourceType(ctx context.Context, movieJavId string, sourceType int64) (*WMedia, error)
		FindOneByMovieNameSourceType(ctx context.Context, movieName string, sourceType int64) (*WMedia, error)
		ListByMovieNames(ctx context.Context, movieNames []string) ([]*WMedia, error)
		FindOneByFileNameSourceType(ctx context.Context, fileName string, sourceType int64) (*WMedia, error)
		ListScMissingNativeMedia(ctx context.Context) ([]*MissingScNativeMediaRow, error)
		CountByDirectoryID(ctx context.Context, directoryID int64) (int64, error)
		ListByDirectoryIDs(ctx context.Context, dirIDs []int64, page, size int, orderBy string) (all, paged []*WMedia, total int64, err error)
		ListDistinctBirthDays(ctx context.Context) ([]int64, error)
		ListNativeMediaAll(ctx context.Context) ([]*WMedia, error)
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

func (m *customWMediaModel) CountByDirectoryID(ctx context.Context, directoryID int64) (int64, error) {
	if directoryID <= 0 {
		return 0, nil
	}
	var count int64
	if err := m.QueryRowNoCacheCtx(ctx, &count, "SELECT COUNT(*) FROM "+m.table+" WHERE `directory_id` = ?", directoryID); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customWMediaModel) FindOneByMovieJavId(ctx context.Context, movieJavId string) (*WMedia, error) {
	return m.FindOneByMovieJavIdSourceType(ctx, movieJavId, consts.WMediaSourceNative)
}

func (m *customWMediaModel) FindOneByMovieName(ctx context.Context, movieName string) (*WMedia, error) {
	return m.FindOneByMovieNameSourceType(ctx, movieName, consts.WMediaSourceNative)
}

func (m *customWMediaModel) ListByMovieNames(ctx context.Context, movieNames []string) ([]*WMedia, error) {
	values := uniqueStringsForWMedia(movieNames)
	if len(values) == 0 {
		return []*WMedia{}, nil
	}
	query, args, err := squirrel.Select(wMediaRows).
		From(m.table).
		Where(squirrel.Eq{"movie_name": values, "source_type": consts.WMediaSourceNative}).
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

func uniqueStringsForWMedia(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
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
		"0 AS sc_times",
		"0 AS come_times",
		"0 AS last_selected_sc_time",
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
	return builder
}

func (m *customWMediaModel) ListNativeMediaPage(ctx context.Context, offset, limit int64, orderBy string, filter types.MediaListFilter) ([]*NativeMediaListRow, error) {
	orderParts := splitOrderClause(orderBy)
	if len(orderParts) == 0 {
		orderParts = []string{"wm.birth_time DESC", "wm.movie_name DESC"}
	}

	builder := applyNativeMediaListFilter(nativeMediaListBuilder(m.table, "wm"), filter)
	if nativeMediaNeedsStatScope(orderBy, filter) {
		rows, err := m.queryNativeMediaRows(ctx, builder)
		if err != nil {
			return nil, err
		}
		if err := m.attachNativeMediaStats(ctx, rows); err != nil {
			return nil, err
		}
		rows = filterNativeMediaStats(rows, filter)
		sortNativeMediaRows(rows, orderBy)
		return paginateNativeMediaRows(rows, offset, limit), nil
	}

	builder = builder.OrderBy(orderParts...)
	if limit > 0 {
		builder = builder.Limit(uint64(limit))
	}
	if offset > 0 {
		builder = builder.Offset(uint64(offset))
	}
	rows, err := m.queryNativeMediaRows(ctx, builder)
	if err != nil {
		return nil, err
	}
	if err := m.attachNativeMediaStats(ctx, rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *customWMediaModel) ListNativeMediaPageWithTotal(ctx context.Context, offset, limit int64, orderBy string, filter types.MediaListFilter) ([]*NativeMediaListRow, int64, error) {
	if nativeMediaNeedsStatScope(orderBy, filter) {
		rows, err := m.queryNativeMediaRows(ctx, applyNativeMediaListFilter(nativeMediaListBuilder(m.table, "wm"), filter))
		if err != nil {
			return nil, 0, err
		}
		if err := m.attachNativeMediaStats(ctx, rows); err != nil {
			return nil, 0, err
		}
		rows = filterNativeMediaStats(rows, filter)
		total := int64(len(rows))
		sortNativeMediaRows(rows, orderBy)
		return paginateNativeMediaRows(rows, offset, limit), total, nil
	}
	total, err := m.CountNativeMedia(ctx, filter)
	if err != nil || total == 0 {
		return []*NativeMediaListRow{}, total, err
	}
	rows, err := m.ListNativeMediaPage(ctx, offset, limit, orderBy, filter)
	return rows, total, err
}

func (m *customWMediaModel) CountNativeMedia(ctx context.Context, filter types.MediaListFilter) (int64, error) {
	if nativeMediaHasStatFilter(filter) {
		rows, err := m.queryNativeMediaRows(ctx, applyNativeMediaListFilter(nativeMediaListBuilder(m.table, "wm"), filter))
		if err != nil {
			return 0, err
		}
		if err := m.attachNativeMediaStats(ctx, rows); err != nil {
			return 0, err
		}
		return int64(len(filterNativeMediaStats(rows, filter))), nil
	}
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

func (m *customWMediaModel) ListScMissingNativeMedia(ctx context.Context) ([]*MissingScNativeMediaRow, error) {
	const query = `
SELECT
	movie_jav_id,
	movie_name,
	sc_times,
	last_selected_sc_time,
	releasing_date,
	media_birth_time
FROM g_sc_stat
WHERE media_birth_time = 0
ORDER BY sc_times DESC, last_selected_sc_time DESC, movie_name DESC, movie_jav_id DESC`

	var rows []*MissingScNativeMediaRow
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*MissingScNativeMediaRow{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func nativeMediaHasStatFilter(filter types.MediaListFilter) bool {
	return filter.HasScTimesMin || filter.HasScTimesMax || filter.HasLastSelectedScFrom || filter.HasLastSelectedScTo
}

func nativeMediaNeedsStatScope(orderBy string, filter types.MediaListFilter) bool {
	return nativeMediaHasStatFilter(filter) || strings.Contains(strings.ToLower(orderBy), "gss.")
}

func (m *customWMediaModel) queryNativeMediaRows(ctx context.Context, builder squirrel.SelectBuilder) ([]*NativeMediaListRow, error) {
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

func (m *customWMediaModel) attachNativeMediaStats(ctx context.Context, rows []*NativeMediaListRow) error {
	javIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		if row != nil {
			javIDs = append(javIDs, row.MovieJavId)
		}
	}
	javIDs = uniqueStringsForWMedia(javIDs)
	if len(javIDs) == 0 {
		return nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(javIDs)), ",")
	args := make([]any, len(javIDs))
	for index, javID := range javIDs {
		args[index] = javID
	}
	var statRows []*GScStat
	if err := m.QueryRowsNoCacheCtx(ctx, &statRows, "SELECT "+gScStatRows+" FROM `g_sc_stat` WHERE `movie_jav_id` IN ("+placeholders+")", args...); err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		return err
	}
	stats := make(map[string]*GScStat, len(statRows))
	for _, row := range statRows {
		if row != nil {
			stats[row.MovieJavId] = row
		}
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		if stat := stats[row.MovieJavId]; stat != nil {
			row.ScTimes = stat.ScTimes
			row.ComeTimes = stat.ComeTimes
			row.LastSelectedScTime = stat.LastSelectedScTime
		}
	}
	return nil
}

func filterNativeMediaStats(rows []*NativeMediaListRow, filter types.MediaListFilter) []*NativeMediaListRow {
	out := make([]*NativeMediaListRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		if filter.HasScTimesMin && row.ScTimes < filter.ScTimesMin {
			continue
		}
		if filter.HasScTimesMax && row.ScTimes > filter.ScTimesMax {
			continue
		}
		if filter.HasLastSelectedScFrom && row.LastSelectedScTime < filter.LastSelectedScFrom {
			continue
		}
		if filter.HasLastSelectedScTo && row.LastSelectedScTime > filter.LastSelectedScTo {
			continue
		}
		out = append(out, row)
	}
	return out
}

func paginateNativeMediaRows(rows []*NativeMediaListRow, offset, limit int64) []*NativeMediaListRow {
	if offset < 0 {
		offset = 0
	}
	if offset >= int64(len(rows)) {
		return []*NativeMediaListRow{}
	}
	end := int64(len(rows))
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return rows[int(offset):int(end)]
}

func sortNativeMediaRows(rows []*NativeMediaListRow, orderBy string) {
	parts := splitOrderClause(orderBy)
	if len(parts) == 0 {
		parts = []string{"wm.birth_time DESC", "wm.movie_name DESC"}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		left, right := rows[i], rows[j]
		for _, part := range parts {
			comparison := compareNativeMediaRows(left, right, part)
			if comparison == 0 {
				continue
			}
			if strings.HasSuffix(strings.ToUpper(strings.TrimSpace(part)), " ASC") {
				return comparison < 0
			}
			return comparison > 0
		}
		return false
	})
}

func compareNativeMediaRows(left, right *NativeMediaListRow, orderPart string) int {
	part := strings.ToLower(orderPart)
	switch {
	case strings.Contains(part, "sc_times"):
		return compareInt64(left.ScTimes, right.ScTimes)
	case strings.Contains(part, "last_selected_sc_time"):
		return compareInt64(left.LastSelectedScTime, right.LastSelectedScTime)
	case strings.Contains(part, "movie_name"):
		return strings.Compare(left.MovieName, right.MovieName)
	case strings.Contains(part, "frame_average"):
		return compareFloat64(left.FrameAverage, right.FrameAverage)
	case strings.Contains(part, "releasing_date"):
		return compareInt64(left.ReleasingDate, right.ReleasingDate)
	case strings.Contains(part, "birth_time"):
		return compareInt64(left.BirthTime, right.BirthTime)
	case strings.Contains(part, "bit_rate"):
		return compareInt64(left.BitRate, right.BitRate)
	case strings.Contains(part, "duration"):
		return compareInt64(left.Duration, right.Duration)
	case strings.Contains(part, "height"):
		return compareInt64(left.Height, right.Height)
	case strings.Contains(part, "self_make"):
		return compareInt64(left.SelfMake, right.SelfMake)
	case strings.Contains(part, "has_mask"):
		return compareInt64(left.HasMask, right.HasMask)
	case strings.Contains(part, "size"):
		return compareInt64(left.Size, right.Size)
	case strings.Contains(part, ".id"):
		return compareInt64(left.Id, right.Id)
	default:
		return 0
	}
}

func compareInt64(left, right int64) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func compareFloat64(left, right float64) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
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

	statOrder := strings.Contains(strings.ToLower(mapWMediaOrderBy(orderBy)), "gss.")
	builder := squirrel.
		Select("wm.*").
		From(m.table + " AS wm").
		Where(squirrel.Eq{"wm.source_type": consts.WMediaSourceNative}).
		Where(squirrel.Eq{"wm.is_removed": consts.FilmIsNotRemoved}).
		Where(squirrel.Eq{"wm.directory_id": dirIDs})
	if !statOrder {
		builder = builder.OrderBy(orderParts...)
	}
	qAll, argsAll, err := builder.ToSql()
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
	if statOrder {
		nativeRows := make([]*NativeMediaListRow, 0, len(rows))
		mediaByID := make(map[int64]*WMedia, len(rows))
		for _, row := range rows {
			if row == nil {
				continue
			}
			mediaByID[row.Id] = row
			nativeRows = append(nativeRows, nativeMediaListRowFromWMedia(row))
		}
		if err := m.attachNativeMediaStats(ctx, nativeRows); err != nil {
			return nil, nil, 0, err
		}
		sortNativeMediaRows(nativeRows, mapWMediaOrderBy(orderBy))
		rows = rows[:0]
		for _, nativeRow := range nativeRows {
			if nativeRow != nil && mediaByID[nativeRow.Id] != nil {
				rows = append(rows, mediaByID[nativeRow.Id])
			}
		}
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

func nativeMediaListRowFromWMedia(row *WMedia) *NativeMediaListRow {
	if row == nil {
		return nil
	}
	return &NativeMediaListRow{
		Id: row.Id, MovieJavId: row.MovieJavId, MovieName: row.MovieName, FileName: row.FileName,
		DirectoryId: row.DirectoryId, RootDir: row.RootDir, FullDir: row.FullDir, Alias: row.Alias,
		Size: row.Size, Width: row.Width, Height: row.Height, BitRate: row.BitRate, Duration: row.Duration,
		FrameAverage: row.FrameAverage, HasSub: row.HasSub, SelfMake: row.SelfMake, HasMask: row.HasMask,
		NeedScanMeta: row.NeedScanMeta, IsRemoved: row.IsRemoved, RemoveTime: row.RemoveTime,
		BirthTime: row.BirthTime, ReleasingDate: row.ReleasingDate, CreatedOn: row.CreatedOn, UpdatedOn: row.UpdatedOn,
	}
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

func (m *customWMediaModel) ListNativeMediaAll(ctx context.Context) ([]*WMedia, error) {
	query, args, err := squirrel.
		Select(wMediaRows).
		From(m.table + " AS wm").
		Where(squirrel.Eq{"wm.source_type": consts.WMediaSourceNative}).
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
		order = "wm.birth_time DESC,wm.movie_name DESC"
	case consts.OrderByMediaBirthTime:
		order = "wm.birth_time DESC,wm.movie_name DESC"
	case consts.OrderByScTimes:
		order = "COALESCE(gss.sc_times, 0) DESC,COALESCE(gss.last_selected_sc_time, 0) DESC,wm.movie_name DESC"
	case consts.OrderByComeTimes:
		order = "COALESCE(gss.come_times, 0) DESC,COALESCE(gss.last_selected_sc_time, 0) DESC,wm.movie_name DESC"
	case consts.OrderByLastSelectedScTime:
		order = "COALESCE(gss.last_selected_sc_time, 0) DESC,wm.movie_name DESC"
	case consts.OrderByReleasingDate:
		order = "wm.releasing_date DESC,wm.movie_name DESC"
	}
	return order
}
