package moviex

import (
	"context"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/local/api/internal/consts"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

type MovieCardBatchCastRow struct {
	MovieJavID             string `db:"movie_jav_id"`
	CastID                 int64  `db:"cast_id"`
	CastName               string `db:"cast_name"`
	PersonID               int64  `db:"person_id"`
	LastSelectedScTime     int64  `db:"last_selected_sc_time"`
	LastParticipatedScTime int64  `db:"last_participated_sc_time"`
	PersonExists           int64  `db:"person_exists"`
	PersonName             string `db:"person_name"`
	PersonChinese          string `db:"person_chinese"`
	PersonBirthDay         int64  `db:"person_birth_day"`
}

type MovieCardBatchGenreRow struct {
	MovieJavID string `db:"movie_jav_id"`
	GenreID    int64  `db:"genre_id"`
	GenreName  string `db:"genre_name"`
}

type MovieCardBatchData struct {
	Movies       map[string]*AMovie
	Murls        map[string]*BmMurl
	Minfos       map[string]*BmMinfo
	Items        map[string]*EItem
	ScStats      map[string]*GScStat
	Media        map[string]*WMedia
	Directors    map[int64]*AmDirector
	Prefixes     map[int64]*AmPrefix
	Labels       map[int64]*AmLabel
	Makers       map[int64]*AmMaker
	Casts        map[string][]*MovieCardBatchCastRow
	Genres       map[string][]*MovieCardBatchGenreRow
	NeedDownload map[string]struct{}
}

type MovieCardBatchRepoSqlx struct {
	conn sqlx.SqlConn
}

func NewMovieCardBatchRepoSqlx(conn sqlx.SqlConn) *MovieCardBatchRepoSqlx {
	return &MovieCardBatchRepoSqlx{conn: conn}
}

func (r *MovieCardBatchRepoSqlx) Load(ctx context.Context, javIDs []string) (*MovieCardBatchData, error) {
	ids := uniqueMovieCardBatchStrings(javIDs)
	data := &MovieCardBatchData{
		Movies:       make(map[string]*AMovie, len(ids)),
		Murls:        make(map[string]*BmMurl, len(ids)),
		Minfos:       make(map[string]*BmMinfo, len(ids)),
		Items:        make(map[string]*EItem, len(ids)),
		ScStats:      make(map[string]*GScStat, len(ids)),
		Media:        make(map[string]*WMedia, len(ids)),
		Directors:    make(map[int64]*AmDirector),
		Prefixes:     make(map[int64]*AmPrefix),
		Labels:       make(map[int64]*AmLabel),
		Makers:       make(map[int64]*AmMaker),
		Casts:        make(map[string][]*MovieCardBatchCastRow, len(ids)),
		Genres:       make(map[string][]*MovieCardBatchGenreRow, len(ids)),
		NeedDownload: make(map[string]struct{}, len(ids)),
	}
	if len(ids) == 0 {
		return data, nil
	}

	movies, err := queryMovieCardBatchRows[AMovie](ctx, r.conn, "a_movie", "jav_id", ids, nil)
	if err != nil {
		return nil, err
	}
	for _, row := range movies {
		if row != nil {
			data.Movies[row.JavId] = row
		}
	}

	murls, err := queryMovieCardBatchRows[BmMurl](ctx, r.conn, "bm_murl", "jav_id", ids, nil)
	if err != nil {
		return nil, err
	}
	for _, row := range murls {
		if row != nil {
			data.Murls[row.JavId] = row
		}
	}

	minfos, err := queryMovieCardBatchRows[BmMinfo](ctx, r.conn, "bm_minfo", "jav_id", ids, nil)
	if err != nil {
		return nil, err
	}
	for _, row := range minfos {
		if row != nil {
			data.Minfos[row.JavId] = row
		}
	}

	items, err := queryMovieCardBatchRows[EItem](ctx, r.conn, "e_item", "jav_id", ids, nil)
	if err != nil {
		return nil, err
	}
	for _, row := range items {
		if row != nil {
			data.Items[row.JavId] = row
		}
	}

	scStats, err := queryMovieCardBatchRows[GScStat](ctx, r.conn, "g_sc_stat", "movie_jav_id", ids, nil)
	if err != nil {
		return nil, err
	}
	for _, row := range scStats {
		if row != nil {
			data.ScStats[row.MovieJavId] = row
		}
	}

	media, err := queryMovieCardBatchRows[WMedia](
		ctx,
		r.conn,
		"w_media",
		"movie_jav_id",
		ids,
		squirrel.Eq{"source_type": consts.WMediaSourceNative},
	)
	if err != nil {
		return nil, err
	}
	for _, row := range media {
		if row != nil {
			data.Media[row.MovieJavId] = row
		}
	}

	if err := r.loadMovieCardBatchLookups(ctx, data); err != nil {
		return nil, err
	}
	if err := r.loadMovieCardBatchCasts(ctx, ids, data); err != nil {
		return nil, err
	}
	if err := r.loadMovieCardBatchGenres(ctx, ids, data); err != nil {
		return nil, err
	}
	if err := r.loadMovieCardBatchNeedDownload(ctx, ids, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (r *MovieCardBatchRepoSqlx) loadMovieCardBatchLookups(ctx context.Context, data *MovieCardBatchData) error {
	directorIDs := make([]int64, 0, len(data.Movies))
	prefixIDs := make([]int64, 0, len(data.Movies))
	labelIDs := make([]int64, 0, len(data.Movies))
	makerIDs := make([]int64, 0, len(data.Movies))
	for _, row := range data.Movies {
		if row == nil {
			continue
		}
		directorIDs = append(directorIDs, row.DirectorId)
		prefixIDs = append(prefixIDs, row.PrefixId)
		labelIDs = append(labelIDs, row.LabelId)
		makerIDs = append(makerIDs, row.MakerId)
	}

	directors, err := queryMovieCardBatchRows[AmDirector](ctx, r.conn, "am_director", "id", uniqueMovieCardBatchInt64s(directorIDs), nil)
	if err != nil {
		return err
	}
	for _, row := range directors {
		if row != nil {
			data.Directors[row.Id] = row
		}
	}
	prefixes, err := queryMovieCardBatchRows[AmPrefix](ctx, r.conn, "am_prefix", "id", uniqueMovieCardBatchInt64s(prefixIDs), nil)
	if err != nil {
		return err
	}
	for _, row := range prefixes {
		if row != nil {
			data.Prefixes[row.Id] = row
		}
	}
	labels, err := queryMovieCardBatchRows[AmLabel](ctx, r.conn, "am_label", "id", uniqueMovieCardBatchInt64s(labelIDs), nil)
	if err != nil {
		return err
	}
	for _, row := range labels {
		if row != nil {
			data.Labels[row.Id] = row
		}
	}
	makers, err := queryMovieCardBatchRows[AmMaker](ctx, r.conn, "am_maker", "id", uniqueMovieCardBatchInt64s(makerIDs), nil)
	if err != nil {
		return err
	}
	for _, row := range makers {
		if row != nil {
			data.Makers[row.Id] = row
		}
	}
	return nil
}

func (r *MovieCardBatchRepoSqlx) loadMovieCardBatchCasts(ctx context.Context, javIDs []string, data *MovieCardBatchData) error {
	for _, batch := range sqlValueBatches(javIDs, 0) {
		query, args, err := squirrel.
			Select(
				"amr.movie_jav_id",
				"ac.id AS cast_id",
				"ac.name AS cast_name",
				"ac.person_id",
				"ac.last_selected_sc_time",
				"ac.last_participated_sc_time",
				"CASE WHEN cp.id IS NULL THEN 0 ELSE 1 END AS person_exists",
				"COALESCE(cp.name, '') AS person_name",
				"COALESCE(cp.chinese, '') AS person_chinese",
				"COALESCE(cp.birth_day, 0) AS person_birth_day",
			).
			From("amr_movie_cast amr").
			Join("am_cast ac ON ac.id = amr.cast_id").
			LeftJoin("c_person cp ON cp.id = ac.person_id").
			Where(squirrel.Eq{"amr.movie_jav_id": batch}).
			OrderBy("amr.movie_jav_id ASC", "amr.cast_id ASC").
			ToSql()
		if err != nil {
			return err
		}
		var rows []*MovieCardBatchCastRow
		if err := queryMovieCardBatch(ctx, r.conn, &rows, query, args...); err != nil {
			return err
		}
		for _, row := range rows {
			if row != nil {
				data.Casts[row.MovieJavID] = append(data.Casts[row.MovieJavID], row)
			}
		}
	}
	return nil
}

func (r *MovieCardBatchRepoSqlx) loadMovieCardBatchGenres(ctx context.Context, javIDs []string, data *MovieCardBatchData) error {
	for _, batch := range sqlValueBatches(javIDs, 0) {
		query, args, err := squirrel.
			Select("amr.movie_jav_id", "ag.id AS genre_id", "ag.name AS genre_name").
			From("amr_movie_genre amr").
			Join("am_genre ag ON ag.id = amr.genre_id").
			Where(squirrel.Eq{"amr.movie_jav_id": batch}).
			OrderBy("amr.movie_jav_id ASC", "amr.genre_id ASC").
			ToSql()
		if err != nil {
			return err
		}
		var rows []*MovieCardBatchGenreRow
		if err := queryMovieCardBatch(ctx, r.conn, &rows, query, args...); err != nil {
			return err
		}
		for _, row := range rows {
			if row != nil {
				data.Genres[row.MovieJavID] = append(data.Genres[row.MovieJavID], row)
			}
		}
	}
	return nil
}

func (r *MovieCardBatchRepoSqlx) loadMovieCardBatchNeedDownload(ctx context.Context, javIDs []string, data *MovieCardBatchData) error {
	for _, batch := range sqlValueBatches(javIDs, 3) {
		query, args, err := squirrel.
			Select("cai.movie_jav_id").
			From("c_movie_album_item cai").
			Join("c_movie_album ca ON ca.id = cai.album_id").
			Where(squirrel.Eq{
				"ca.name":          consts.MovieNeedDownloadAlbumName,
				"cai.movie_type":   moviedomain.SehuatangThreadMovieTypeJAV,
				"cai.fc2_key":      "",
				"cai.movie_jav_id": batch,
			}).
			ToSql()
		if err != nil {
			return err
		}
		var ids []string
		if err := queryMovieCardBatch(ctx, r.conn, &ids, query, args...); err != nil {
			return err
		}
		for _, javID := range ids {
			data.NeedDownload[javID] = struct{}{}
		}
	}
	return nil
}

func queryMovieCardBatchRows[T any, V comparable](
	ctx context.Context,
	conn sqlx.SqlConn,
	tableName string,
	keyColumn string,
	values []V,
	extra squirrel.Sqlizer,
) ([]*T, error) {
	if len(values) == 0 {
		return []*T{}, nil
	}
	rows := make([]*T, 0, len(values))
	reserved := 0
	if extra != nil {
		reserved = 1
	}
	for _, batch := range sqlValueBatches(values, reserved) {
		builder := squirrel.Select("*").From(tableName).Where(squirrel.Eq{keyColumn: batch})
		if extra != nil {
			builder = builder.Where(extra)
		}
		query, args, err := builder.ToSql()
		if err != nil {
			return nil, err
		}
		var batchRows []*T
		if err := queryMovieCardBatch(ctx, conn, &batchRows, query, args...); err != nil {
			return nil, err
		}
		rows = append(rows, batchRows...)
	}
	return rows, nil
}

func queryMovieCardBatch(ctx context.Context, conn sqlx.SqlConn, dest any, query string, args ...any) error {
	if err := conn.QueryRowsCtx(ctx, dest, query, args...); err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		return err
	}
	return nil
}

func uniqueMovieCardBatchStrings(values []string) []string {
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

func uniqueMovieCardBatchInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	out := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
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
