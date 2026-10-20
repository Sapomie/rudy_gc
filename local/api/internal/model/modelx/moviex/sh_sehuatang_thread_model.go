package moviex

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	moviedomain "rudy_gc/local/api/internal/types/movie"
)

type (
	ShSehuatangThreadModel interface {
		Insert(ctx context.Context, data *ShSehuatangThread) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*ShSehuatangThread, error)
		ListByIDs(ctx context.Context, ids []int64) ([]*ShSehuatangThread, error)
		FindOneByThreadTID(ctx context.Context, threadTID string) (*ShSehuatangThread, error)
		FindExistingThreadTIDs(ctx context.Context, threadTIDs []string) (map[string]struct{}, error)
		FindThreadAggTargetsByTIDs(ctx context.Context, threadTIDs []string) (map[string]ShThreadAggTarget, error)
		DeleteByThreadTIDs(ctx context.Context, threadTIDs []string) (int64, error)
		UpdateFC2MetaJSONByID(ctx context.Context, id int64, metaJSON string) error
		UpdateFC2MovieIDByID(ctx context.Context, id, fc2MovieID int64) error
		UpdateJAVMovieFieldsByMovieName(ctx context.Context, movieName, movieJavID string, updatedOn int64) (int64, error)
		ListFetchedByFC2MovieID(ctx context.Context, fc2MovieID int64) ([]*ShSehuatangThread, error)
		ListDistinctPostDayForumTargets(ctx context.Context) ([]ShThreadAggTarget, error)
		CalcBucket(ctx context.Context, startUnix, endUnix int64, sourceForumID string) (*ShThreadBucketCalc, error)
		ListPage(ctx context.Context, offset, limit int64, orderBy string, filter ShSehuatangThreadListFilter) ([]*ShSehuatangThread, error)
		CountAll(ctx context.Context, filter ShSehuatangThreadListFilter) (int64, error)
		Aggregate(ctx context.Context, filter ShSehuatangThreadListFilter) (*ShSehuatangThreadAggregate, error)
		UpdateLastSeenByThreadTID(ctx context.Context, threadTID string, now int64) (*ShSehuatangThread, error)
		UpdateDetailFailureByThreadTID(ctx context.Context, threadTID string, errText string, now int64) (*ShSehuatangThread, error)
		UpsertListBasic(ctx context.Context, data *ShSehuatangThread) (*ShSehuatangThread, error)
		Upsert(ctx context.Context, data *ShSehuatangThread) (*ShSehuatangThread, error)
		TableName() string
	}

	defaultShSehuatangThreadModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ShSehuatangThread struct {
		Id              int64  `db:"id"`
		ThreadTID       string `db:"thread_tid"`
		ThreadURL       string `db:"thread_url"`
		ThreadTitle     string `db:"thread_title"`
		MovieJavId      string `db:"movie_jav_id"`
		MovieName       string `db:"movie_name"`
		MovieType       string `db:"movie_type"`
		FC2MovieID      int64  `db:"fc2_movie_id"`
		FC2Key          string `db:"fc2_key"`
		AuthorName      string `db:"author_name"`
		PostTime        int64  `db:"post_time"`
		SourceForumID   string `db:"source_forum_id"`
		SourcePage      int64  `db:"source_page"`
		SourcePageURL   string `db:"source_page_url"`
		DetailStatus    string `db:"detail_status"`
		DetailFetchedOn int64  `db:"detail_fetched_on"`
		DetailFailedOn  int64  `db:"detail_failed_on"`
		DetailError     string `db:"detail_error"`
		RawContentText  string `db:"raw_content_text"`
		MagnetText      string `db:"magnet_text"`
		InfoHashText    string `db:"info_hash_text"`
		FC2MetaJSON     string `db:"fc2_meta_json"`
		ContentHash     string `db:"content_hash"`
		FirstSeenTime   int64  `db:"first_seen_time"`
		LastSeenTime    int64  `db:"last_seen_time"`
		CreatedOn       int64  `db:"created_on"`
		UpdatedOn       int64  `db:"updated_on"`
	}

	ShSehuatangThreadListFilter struct {
		Keyword             string
		ThreadTID           string
		AuthorName          string
		InfoHash            string
		MovieType           string
		MovieTypes          []string
		SourceForumID       string
		SourceForumIDs      []string
		DetailStatus        string
		DetailStatuses      []string
		PostTimeFrom        int64
		HasPostTimeFrom     bool
		PostTimeTo          int64
		HasPostTimeTo       bool
		LastSeenTimeFrom    int64
		HasLastSeenTimeFrom bool
		LastSeenTimeTo      int64
		HasLastSeenTimeTo   bool
	}

	ShSehuatangThreadAggregate struct {
		Total              int64 `db:"total"`
		PendingDetailCount int64 `db:"pending_detail_count"`
		FetchedDetailCount int64 `db:"fetched_detail_count"`
		FailedDetailCount  int64 `db:"failed_detail_count"`
		Forum36Count       int64 `db:"forum_36_count"`
		Forum103Count      int64 `db:"forum_103_count"`
	}

	ShThreadAggTarget struct {
		ThreadTID     string `db:"thread_tid"`
		PostTime      int64  `db:"post_time"`
		SourceForumID string `db:"source_forum_id"`
		DetailStatus  string `db:"detail_status"`
	}

	ShThreadBucketCalc struct {
		CountAll          int64 `db:"count_all"`
		CountPending      int64 `db:"count_pending"`
		CountFetched      int64 `db:"count_fetched"`
		CountFailed       int64 `db:"count_failed"`
		CountWithMagnet   int64 `db:"count_with_magnet"`
		CountWithInfoHash int64 `db:"count_with_info_hash"`
		LatestPostTime    int64 `db:"latest_post_time"`
	}
)

var shSehuatangThreadRows = "`id`,`thread_tid`,`thread_url`,`thread_title`,`movie_jav_id`,`movie_name`,`movie_type`,`fc2_movie_id`,'' AS `fc2_key`,`author_name`,`post_time`,`source_forum_id`,`source_page`,`source_page_url`,`detail_status`,`detail_fetched_on`,`detail_failed_on`,COALESCE(`detail_error`, '') AS `detail_error`,`raw_content_text`,`magnet_text`,`info_hash_text`,COALESCE(CAST(`fc2_meta_json` AS CHAR), '') AS `fc2_meta_json`,`content_hash`,`first_seen_time`,`last_seen_time`,`created_on`,`updated_on`"

func NewShSehuatangThreadModel(conn sqlx.SqlConn) ShSehuatangThreadModel {
	return &defaultShSehuatangThreadModel{
		conn:  conn,
		table: "`sh_thread`",
	}
}

func (m *defaultShSehuatangThreadModel) TableName() string {
	return m.table
}

func (m *defaultShSehuatangThreadModel) Insert(ctx context.Context, data *ShSehuatangThread) (sql.Result, error) {
	data.MovieType = moviedomain.ResolveSehuatangThreadMovieType(data.MovieName, data.MovieJavId)
	query := "insert into " + m.table + " (`thread_tid`,`thread_url`,`thread_title`,`movie_jav_id`,`movie_name`,`movie_type`,`author_name`,`post_time`,`source_forum_id`,`source_page`,`source_page_url`,`detail_status`,`detail_fetched_on`,`detail_failed_on`,`detail_error`,`raw_content_text`,`magnet_text`,`info_hash_text`,`content_hash`,`first_seen_time`,`last_seen_time`,`created_on`,`updated_on`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	return m.conn.ExecCtx(ctx, query, data.ThreadTID, data.ThreadURL, data.ThreadTitle, data.MovieJavId, data.MovieName, data.MovieType, data.AuthorName, data.PostTime, data.SourceForumID, data.SourcePage, data.SourcePageURL, data.DetailStatus, data.DetailFetchedOn, data.DetailFailedOn, data.DetailError, data.RawContentText, data.MagnetText, data.InfoHashText, data.ContentHash, data.FirstSeenTime, data.LastSeenTime, data.CreatedOn, data.UpdatedOn)
}

func (m *defaultShSehuatangThreadModel) FindOne(ctx context.Context, id int64) (*ShSehuatangThread, error) {
	var row ShSehuatangThread
	query := "select " + shSehuatangThreadRows + " from " + m.table + " where `id` = ? limit 1"
	if err := m.conn.QueryRowCtx(ctx, &row, query, id); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultShSehuatangThreadModel) ListByIDs(ctx context.Context, ids []int64) ([]*ShSehuatangThread, error) {
	values := uniquePositiveInt64s(ids)
	if len(values) == 0 {
		return []*ShSehuatangThread{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(values)), ",")
	args := make([]any, len(values))
	for index, id := range values {
		args[index] = id
	}
	var rows []*ShSehuatangThread
	if err := m.conn.QueryRowsCtx(ctx, &rows, "SELECT "+shSehuatangThreadRows+" FROM "+m.table+" WHERE `id` IN ("+placeholders+")", args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*ShSehuatangThread{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultShSehuatangThreadModel) FindOneByThreadTID(ctx context.Context, threadTID string) (*ShSehuatangThread, error) {
	threadTID = strings.TrimSpace(threadTID)
	if threadTID == "" {
		return nil, ErrNotFound
	}
	var row ShSehuatangThread
	query := "select " + shSehuatangThreadRows + " from " + m.table + " where `thread_tid` = ? limit 1"
	if err := m.conn.QueryRowCtx(ctx, &row, query, threadTID); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultShSehuatangThreadModel) FindExistingThreadTIDs(ctx context.Context, threadTIDs []string) (map[string]struct{}, error) {
	values := make([]string, 0, len(threadTIDs))
	seen := make(map[string]struct{}, len(threadTIDs))
	for _, threadTID := range threadTIDs {
		threadTID = strings.TrimSpace(threadTID)
		if threadTID == "" {
			continue
		}
		if _, ok := seen[threadTID]; ok {
			continue
		}
		seen[threadTID] = struct{}{}
		values = append(values, threadTID)
	}
	if len(values) == 0 {
		return map[string]struct{}{}, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(values)), ",")
	args := make([]any, len(values))
	for index, value := range values {
		args[index] = value
	}
	var rows []struct {
		ThreadTID string `db:"thread_tid"`
	}
	query := "select `thread_tid` from " + m.table + " where `thread_tid` in (" + placeholders + ")"
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	out := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		out[row.ThreadTID] = struct{}{}
	}
	return out, nil
}

func (m *defaultShSehuatangThreadModel) FindThreadAggTargetsByTIDs(ctx context.Context, threadTIDs []string) (map[string]ShThreadAggTarget, error) {
	values := normalizeShThreadTIDs(threadTIDs)
	if len(values) == 0 {
		return map[string]ShThreadAggTarget{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(values)), ",")
	args := make([]any, len(values))
	for index, value := range values {
		args[index] = value
	}
	var rows []ShThreadAggTarget
	query := "select `thread_tid`,`post_time`,`source_forum_id`,`detail_status` from " + m.table + " where `thread_tid` in (" + placeholders + ")"
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[string]ShThreadAggTarget{}, nil
		}
		return nil, err
	}
	out := make(map[string]ShThreadAggTarget, len(rows))
	for _, row := range rows {
		out[row.ThreadTID] = row
	}
	return out, nil
}

func (m *defaultShSehuatangThreadModel) DeleteByThreadTIDs(ctx context.Context, threadTIDs []string) (int64, error) {
	values := normalizeShThreadTIDs(threadTIDs)
	if len(values) == 0 {
		return 0, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(values)), ",")
	args := make([]any, len(values))
	for index, value := range values {
		args[index] = value
	}
	result, err := m.conn.ExecCtx(ctx, "delete from "+m.table+" where `thread_tid` in ("+placeholders+")", args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (m *defaultShSehuatangThreadModel) UpdateFC2MetaJSONByID(ctx context.Context, id int64, metaJSON string) error {
	_, err := m.conn.ExecCtx(ctx, "UPDATE `sh_thread` SET `fc2_meta_json` = ? WHERE `id` = ?", metaJSON, id)
	return err
}

func (m *defaultShSehuatangThreadModel) UpdateFC2MovieIDByID(ctx context.Context, id, fc2MovieID int64) error {
	_, err := m.conn.ExecCtx(ctx, "UPDATE `sh_thread` SET `fc2_movie_id` = ? WHERE `id` = ?", fc2MovieID, id)
	return err
}

func (m *defaultShSehuatangThreadModel) UpdateJAVMovieFieldsByMovieName(ctx context.Context, movieName, movieJavID string, updatedOn int64) (int64, error) {
	movieName = strings.TrimSpace(movieName)
	movieJavID = strings.TrimSpace(movieJavID)
	if movieName == "" || movieJavID == "" {
		return 0, nil
	}
	result, err := m.conn.ExecCtx(
		ctx,
		`UPDATE sh_thread
SET movie_jav_id = ?, movie_type = ?, updated_on = GREATEST(updated_on, ?)
WHERE movie_name = ?
  AND LOWER(TRIM(COALESCE(movie_type, ''))) <> ?`,
		movieJavID,
		moviedomain.SehuatangThreadMovieTypeJAV,
		updatedOn,
		movieName,
		moviedomain.SehuatangThreadMovieTypeFC2PPV,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (m *defaultShSehuatangThreadModel) ListFetchedByFC2MovieID(ctx context.Context, fc2MovieID int64) ([]*ShSehuatangThread, error) {
	var rows []*ShSehuatangThread
	query := "SELECT " + shSehuatangThreadRows + " FROM `sh_thread` WHERE `movie_type` = ? AND `fc2_movie_id` = ? AND `detail_status` = 'fetched' ORDER BY `post_time` DESC, `id` DESC"
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, moviedomain.SehuatangThreadMovieTypeFC2PPV, fc2MovieID); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*ShSehuatangThread{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultShSehuatangThreadModel) ListDistinctPostDayForumTargets(ctx context.Context) ([]ShThreadAggTarget, error) {
	query := `
select
	cast(unix_timestamp(date(from_unixtime(post_time))) as signed) as post_time,
	source_forum_id
from sh_thread
where post_time > 0
group by cast(unix_timestamp(date(from_unixtime(post_time))) as signed), source_forum_id
order by post_time asc, source_forum_id asc`
	type row struct {
		PostTime      int64  `db:"post_time"`
		SourceForumID string `db:"source_forum_id"`
	}
	var rows []row
	if err := m.conn.QueryRowsCtx(ctx, &rows, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []ShThreadAggTarget{}, nil
		}
		return nil, err
	}
	out := make([]ShThreadAggTarget, 0, len(rows))
	for _, item := range rows {
		out = append(out, ShThreadAggTarget{PostTime: item.PostTime, SourceForumID: item.SourceForumID})
	}
	return out, nil
}

func (m *defaultShSehuatangThreadModel) CalcBucket(ctx context.Context, startUnix, endUnix int64, sourceForumID string) (*ShThreadBucketCalc, error) {
	query := `
select
	count(*) as count_all,
	coalesce(sum(case when detail_status = 'pending' then 1 else 0 end), 0) as count_pending,
	coalesce(sum(case when detail_status = 'fetched' then 1 else 0 end), 0) as count_fetched,
	coalesce(sum(case when detail_status = 'failed' then 1 else 0 end), 0) as count_failed,
	coalesce(sum(case when trim(coalesce(magnet_text, '')) <> '' then 1 else 0 end), 0) as count_with_magnet,
	coalesce(sum(case when trim(coalesce(info_hash_text, '')) <> '' then 1 else 0 end), 0) as count_with_info_hash,
	coalesce(max(post_time), 0) as latest_post_time
from sh_thread
where post_time >= ? and post_time <= ?`
	args := []any{startUnix, endUnix}
	if sourceForumID = strings.TrimSpace(sourceForumID); sourceForumID != "" {
		query += " and source_forum_id = ?"
		args = append(args, sourceForumID)
	}
	var out ShThreadBucketCalc
	if err := m.conn.QueryRowCtx(ctx, &out, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return &ShThreadBucketCalc{}, nil
		}
		return nil, err
	}
	return &out, nil
}

func normalizeShThreadTIDs(threadTIDs []string) []string {
	values := make([]string, 0, len(threadTIDs))
	seen := make(map[string]struct{}, len(threadTIDs))
	for _, threadTID := range threadTIDs {
		threadTID = strings.TrimSpace(threadTID)
		if threadTID == "" {
			continue
		}
		if _, ok := seen[threadTID]; ok {
			continue
		}
		seen[threadTID] = struct{}{}
		values = append(values, threadTID)
	}
	return values
}

func (m *defaultShSehuatangThreadModel) ListPage(ctx context.Context, offset, limit int64, orderBy string, filter ShSehuatangThreadListFilter) ([]*ShSehuatangThread, error) {
	if offset < 0 {
		offset = 0
	}
	if strings.TrimSpace(orderBy) == "" {
		orderBy = "`post_time` DESC, `id` DESC"
	}

	builder := applyShSehuatangThreadListFilter(squirrel.
		Select(shSehuatangThreadRows).
		From(m.table), filter)
	builder = builder.OrderBy(orderBy)
	if offset > 0 {
		builder = builder.Offset(uint64(offset))
	}
	if limit > 0 {
		builder = builder.Limit(uint64(limit))
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*ShSehuatangThread
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*ShSehuatangThread{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultShSehuatangThreadModel) CountAll(ctx context.Context, filter ShSehuatangThreadListFilter) (int64, error) {
	builder := applyShSehuatangThreadListFilter(squirrel.
		Select("COUNT(*)").
		From(m.table), filter)
	query, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func (m *defaultShSehuatangThreadModel) Aggregate(ctx context.Context, filter ShSehuatangThreadListFilter) (*ShSehuatangThreadAggregate, error) {
	builder := applyShSehuatangThreadListFilter(squirrel.
		Select(
			"COUNT(*) AS total",
			"COALESCE(SUM(CASE WHEN `detail_status` = 'pending' THEN 1 ELSE 0 END), 0) AS pending_detail_count",
			"COALESCE(SUM(CASE WHEN `detail_status` = 'fetched' THEN 1 ELSE 0 END), 0) AS fetched_detail_count",
			"COALESCE(SUM(CASE WHEN `detail_status` = 'failed' THEN 1 ELSE 0 END), 0) AS failed_detail_count",
			"COALESCE(SUM(CASE WHEN `source_forum_id` = '36' THEN 1 ELSE 0 END), 0) AS forum_36_count",
			"COALESCE(SUM(CASE WHEN `source_forum_id` = '103' THEN 1 ELSE 0 END), 0) AS forum_103_count",
		).
		From(m.table), filter)
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var aggregate ShSehuatangThreadAggregate
	if err := m.conn.QueryRowCtx(ctx, &aggregate, query, args...); err != nil {
		return nil, err
	}
	return &aggregate, nil
}

func (m *defaultShSehuatangThreadModel) UpdateLastSeenByThreadTID(ctx context.Context, threadTID string, now int64) (*ShSehuatangThread, error) {
	threadTID = strings.TrimSpace(threadTID)
	if threadTID == "" {
		return nil, ErrNotFound
	}
	query := "update " + m.table + " set `last_seen_time` = ?, `updated_on` = ? where `thread_tid` = ?"
	if _, err := m.conn.ExecCtx(ctx, query, now, now, threadTID); err != nil {
		return nil, err
	}
	return m.FindOneByThreadTID(ctx, threadTID)
}

func (m *defaultShSehuatangThreadModel) UpdateDetailFailureByThreadTID(ctx context.Context, threadTID string, errText string, now int64) (*ShSehuatangThread, error) {
	threadTID = strings.TrimSpace(threadTID)
	if threadTID == "" {
		return nil, ErrNotFound
	}
	query := "update " + m.table + " set `detail_status` = ?, `detail_failed_on` = ?, `detail_error` = ?, `updated_on` = ? where `thread_tid` = ?"
	if _, err := m.conn.ExecCtx(ctx, query, "failed", now, errText, now, threadTID); err != nil {
		return nil, err
	}
	return m.FindOneByThreadTID(ctx, threadTID)
}

func (m *defaultShSehuatangThreadModel) UpsertListBasic(ctx context.Context, data *ShSehuatangThread) (*ShSehuatangThread, error) {
	data.MovieType = moviedomain.ResolveSehuatangThreadMovieType(data.MovieName, data.MovieJavId)
	query := "insert into " + m.table + " (`thread_tid`,`thread_url`,`thread_title`,`movie_jav_id`,`movie_name`,`movie_type`,`author_name`,`post_time`,`source_forum_id`,`source_page`,`source_page_url`,`detail_status`,`detail_fetched_on`,`detail_failed_on`,`detail_error`,`raw_content_text`,`magnet_text`,`info_hash_text`,`content_hash`,`first_seen_time`,`last_seen_time`,`created_on`,`updated_on`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) on duplicate key update `thread_url` = values(`thread_url`), `thread_title` = values(`thread_title`), `movie_jav_id` = case when trim(coalesce(values(`movie_name`), '')) <> '' and values(`movie_name`) <> `movie_name` then values(`movie_jav_id`) when trim(coalesce(values(`movie_jav_id`), '')) <> '' then values(`movie_jav_id`) else `movie_jav_id` end, `movie_name` = case when trim(coalesce(values(`movie_name`), '')) <> '' then values(`movie_name`) else `movie_name` end, `movie_type` = case when trim(coalesce(values(`movie_type`), '')) <> '' then values(`movie_type`) else `movie_type` end, `author_name` = values(`author_name`), `post_time` = values(`post_time`), `source_forum_id` = values(`source_forum_id`), `source_page` = values(`source_page`), `source_page_url` = values(`source_page_url`), `detail_status` = case when trim(coalesce(`detail_status`, '')) = '' then values(`detail_status`) else `detail_status` end, `first_seen_time` = case when `first_seen_time` = 0 then values(`first_seen_time`) else `first_seen_time` end, `last_seen_time` = values(`last_seen_time`), `updated_on` = values(`updated_on`)"
	if _, err := m.conn.ExecCtx(ctx, query, data.ThreadTID, data.ThreadURL, data.ThreadTitle, data.MovieJavId, data.MovieName, data.MovieType, data.AuthorName, data.PostTime, data.SourceForumID, data.SourcePage, data.SourcePageURL, data.DetailStatus, data.DetailFetchedOn, data.DetailFailedOn, data.DetailError, data.RawContentText, data.MagnetText, data.InfoHashText, data.ContentHash, data.FirstSeenTime, data.LastSeenTime, data.CreatedOn, data.UpdatedOn); err != nil {
		return nil, err
	}
	return m.FindOneByThreadTID(ctx, data.ThreadTID)
}

func (m *defaultShSehuatangThreadModel) Upsert(ctx context.Context, data *ShSehuatangThread) (*ShSehuatangThread, error) {
	data.MovieType = moviedomain.ResolveSehuatangThreadMovieType(data.MovieName, data.MovieJavId)
	query := "insert into " + m.table + " (`thread_tid`,`thread_url`,`thread_title`,`movie_jav_id`,`movie_name`,`movie_type`,`author_name`,`post_time`,`source_forum_id`,`source_page`,`source_page_url`,`detail_status`,`detail_fetched_on`,`detail_failed_on`,`detail_error`,`raw_content_text`,`magnet_text`,`info_hash_text`,`content_hash`,`first_seen_time`,`last_seen_time`,`created_on`,`updated_on`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) on duplicate key update `thread_url` = values(`thread_url`), `thread_title` = values(`thread_title`), `movie_jav_id` = case when trim(coalesce(values(`movie_name`), '')) <> '' and values(`movie_name`) <> `movie_name` then values(`movie_jav_id`) when trim(coalesce(values(`movie_jav_id`), '')) <> '' then values(`movie_jav_id`) else `movie_jav_id` end, `movie_name` = case when trim(coalesce(values(`movie_name`), '')) <> '' then values(`movie_name`) else `movie_name` end, `movie_type` = values(`movie_type`), `author_name` = values(`author_name`), `post_time` = values(`post_time`), `source_forum_id` = values(`source_forum_id`), `source_page` = values(`source_page`), `source_page_url` = values(`source_page_url`), `detail_status` = values(`detail_status`), `detail_fetched_on` = values(`detail_fetched_on`), `detail_failed_on` = values(`detail_failed_on`), `detail_error` = values(`detail_error`), `raw_content_text` = values(`raw_content_text`), `magnet_text` = values(`magnet_text`), `info_hash_text` = values(`info_hash_text`), `content_hash` = values(`content_hash`), `first_seen_time` = case when `first_seen_time` = 0 then values(`first_seen_time`) else `first_seen_time` end, `last_seen_time` = values(`last_seen_time`), `updated_on` = values(`updated_on`)"
	if _, err := m.conn.ExecCtx(ctx, query, data.ThreadTID, data.ThreadURL, data.ThreadTitle, data.MovieJavId, data.MovieName, data.MovieType, data.AuthorName, data.PostTime, data.SourceForumID, data.SourcePage, data.SourcePageURL, data.DetailStatus, data.DetailFetchedOn, data.DetailFailedOn, data.DetailError, data.RawContentText, data.MagnetText, data.InfoHashText, data.ContentHash, data.FirstSeenTime, data.LastSeenTime, data.CreatedOn, data.UpdatedOn); err != nil {
		return nil, err
	}
	return m.FindOneByThreadTID(ctx, data.ThreadTID)
}

func applyShSehuatangThreadListFilter(builder squirrel.SelectBuilder, filter ShSehuatangThreadListFilter) squirrel.SelectBuilder {
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		builder = builder.Where(squirrel.Or{
			squirrel.Like{"thread_title": like},
			squirrel.Like{"movie_name": like},
			squirrel.Like{"movie_jav_id": like},
			squirrel.Like{"raw_content_text": like},
			squirrel.Like{"thread_tid": like},
		})
	}
	if threadTID := strings.TrimSpace(filter.ThreadTID); threadTID != "" {
		builder = builder.Where(squirrel.Like{"thread_tid": "%" + threadTID + "%"})
	}
	if authorName := strings.TrimSpace(filter.AuthorName); authorName != "" {
		builder = builder.Where(squirrel.Like{"author_name": "%" + authorName + "%"})
	}
	if infoHash := strings.TrimSpace(filter.InfoHash); infoHash != "" {
		builder = builder.Where(squirrel.Like{"info_hash_text": "%" + infoHash + "%"})
	}
	movieTypes := moviedomain.NormalizeSehuatangThreadMovieTypes(filter.MovieTypes...)
	if len(movieTypes) > 0 {
		builder = builder.Where(squirrel.Eq{"movie_type": movieTypes})
	} else if movieType := strings.TrimSpace(filter.MovieType); movieType != "" {
		if normalized := moviedomain.NormalizeSehuatangThreadMovieTypes(movieType); len(normalized) > 0 {
			builder = builder.Where(squirrel.Eq{"movie_type": normalized[0]})
		}
	}
	if sourceForumID := strings.TrimSpace(filter.SourceForumID); sourceForumID != "" {
		builder = builder.Where(squirrel.Eq{"source_forum_id": sourceForumID})
	}
	if len(filter.SourceForumIDs) > 0 {
		ids := make([]string, 0, len(filter.SourceForumIDs))
		for _, item := range filter.SourceForumIDs {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			ids = append(ids, item)
		}
		if len(ids) > 0 {
			builder = builder.Where(squirrel.Eq{"source_forum_id": ids})
		}
	}
	detailStatuses := make([]string, 0, len(filter.DetailStatuses))
	seenDetailStatuses := make(map[string]struct{}, len(filter.DetailStatuses))
	for _, detailStatus := range filter.DetailStatuses {
		detailStatus = strings.TrimSpace(detailStatus)
		if detailStatus == "" {
			continue
		}
		if _, ok := seenDetailStatuses[detailStatus]; ok {
			continue
		}
		seenDetailStatuses[detailStatus] = struct{}{}
		detailStatuses = append(detailStatuses, detailStatus)
	}
	if len(detailStatuses) > 0 {
		builder = builder.Where(squirrel.Eq{"detail_status": detailStatuses})
	} else if detailStatus := strings.TrimSpace(filter.DetailStatus); detailStatus != "" {
		builder = builder.Where(squirrel.Eq{"detail_status": detailStatus})
	}
	if filter.HasPostTimeFrom {
		builder = builder.Where(squirrel.GtOrEq{"post_time": filter.PostTimeFrom})
	}
	if filter.HasPostTimeTo {
		builder = builder.Where(squirrel.LtOrEq{"post_time": filter.PostTimeTo})
	}
	if filter.HasLastSeenTimeFrom {
		builder = builder.Where(squirrel.GtOrEq{"last_seen_time": filter.LastSeenTimeFrom})
	}
	if filter.HasLastSeenTimeTo {
		builder = builder.Where(squirrel.LtOrEq{"last_seen_time": filter.LastSeenTimeTo})
	}
	return builder
}
