package moviex

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	ShSehuatangThreadModel interface {
		Insert(ctx context.Context, data *ShSehuatangThread) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*ShSehuatangThread, error)
		FindOneByThreadTID(ctx context.Context, threadTID string) (*ShSehuatangThread, error)
		ListPage(ctx context.Context, offset, limit int64, orderBy string, filter ShSehuatangThreadListFilter) ([]*ShSehuatangThread, error)
		CountAll(ctx context.Context, filter ShSehuatangThreadListFilter) (int64, error)
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
		ContentHash     string `db:"content_hash"`
		LastSeenTime    int64  `db:"last_seen_time"`
		CreatedOn       int64  `db:"created_on"`
		UpdatedOn       int64  `db:"updated_on"`
	}

	ShSehuatangThreadListFilter struct {
		Keyword             string
		ThreadTID           string
		AuthorName          string
		InfoHash            string
		SourceForumID       string
		SourceForumIDs      []string
		DetailStatus        string
		DetailStatuses      []string
		HasMagnet           bool
		HasMagnetSet        bool
		PostTimeFrom        int64
		HasPostTimeFrom     bool
		PostTimeTo          int64
		HasPostTimeTo       bool
		LastSeenTimeFrom    int64
		HasLastSeenTimeFrom bool
		LastSeenTimeTo      int64
		HasLastSeenTimeTo   bool
	}
)

var shSehuatangThreadRows = "`id`,`thread_tid`,`thread_url`,`thread_title`,`author_name`,`post_time`,`source_forum_id`,`source_page`,`source_page_url`,`detail_status`,`detail_fetched_on`,`detail_failed_on`,COALESCE(`detail_error`, '') AS `detail_error`,`raw_content_text`,`magnet_text`,`info_hash_text`,`content_hash`,`last_seen_time`,`created_on`,`updated_on`"

func NewShSehuatangThreadModel(conn sqlx.SqlConn) ShSehuatangThreadModel {
	return &defaultShSehuatangThreadModel{
		conn:  conn,
		table: "`sh_sehuatang_thread`",
	}
}

func (m *defaultShSehuatangThreadModel) TableName() string {
	return m.table
}

func (m *defaultShSehuatangThreadModel) Insert(ctx context.Context, data *ShSehuatangThread) (sql.Result, error) {
	query := "insert into " + m.table + " (`thread_tid`,`thread_url`,`thread_title`,`author_name`,`post_time`,`source_forum_id`,`source_page`,`source_page_url`,`detail_status`,`detail_fetched_on`,`detail_failed_on`,`detail_error`,`raw_content_text`,`magnet_text`,`info_hash_text`,`content_hash`,`last_seen_time`,`created_on`,`updated_on`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	return m.conn.ExecCtx(ctx, query, data.ThreadTID, data.ThreadURL, data.ThreadTitle, data.AuthorName, data.PostTime, data.SourceForumID, data.SourcePage, data.SourcePageURL, data.DetailStatus, data.DetailFetchedOn, data.DetailFailedOn, data.DetailError, data.RawContentText, data.MagnetText, data.InfoHashText, data.ContentHash, data.LastSeenTime, data.CreatedOn, data.UpdatedOn)
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
	query := "insert into " + m.table + " (`thread_tid`,`thread_url`,`thread_title`,`author_name`,`post_time`,`source_forum_id`,`source_page`,`source_page_url`,`detail_status`,`detail_fetched_on`,`detail_failed_on`,`detail_error`,`raw_content_text`,`magnet_text`,`info_hash_text`,`content_hash`,`last_seen_time`,`created_on`,`updated_on`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) on duplicate key update `thread_url` = values(`thread_url`), `thread_title` = values(`thread_title`), `author_name` = values(`author_name`), `post_time` = values(`post_time`), `source_forum_id` = values(`source_forum_id`), `source_page` = values(`source_page`), `source_page_url` = values(`source_page_url`), `detail_status` = case when trim(coalesce(`detail_status`, '')) = '' then values(`detail_status`) else `detail_status` end, `last_seen_time` = values(`last_seen_time`), `updated_on` = values(`updated_on`)"
	if _, err := m.conn.ExecCtx(ctx, query, data.ThreadTID, data.ThreadURL, data.ThreadTitle, data.AuthorName, data.PostTime, data.SourceForumID, data.SourcePage, data.SourcePageURL, data.DetailStatus, data.DetailFetchedOn, data.DetailFailedOn, data.DetailError, data.RawContentText, data.MagnetText, data.InfoHashText, data.ContentHash, data.LastSeenTime, data.CreatedOn, data.UpdatedOn); err != nil {
		return nil, err
	}
	return m.FindOneByThreadTID(ctx, data.ThreadTID)
}

func (m *defaultShSehuatangThreadModel) Upsert(ctx context.Context, data *ShSehuatangThread) (*ShSehuatangThread, error) {
	query := "insert into " + m.table + " (`thread_tid`,`thread_url`,`thread_title`,`author_name`,`post_time`,`source_forum_id`,`source_page`,`source_page_url`,`detail_status`,`detail_fetched_on`,`detail_failed_on`,`detail_error`,`raw_content_text`,`magnet_text`,`info_hash_text`,`content_hash`,`last_seen_time`,`created_on`,`updated_on`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) on duplicate key update `thread_url` = values(`thread_url`), `thread_title` = values(`thread_title`), `author_name` = values(`author_name`), `post_time` = values(`post_time`), `source_forum_id` = values(`source_forum_id`), `source_page` = values(`source_page`), `source_page_url` = values(`source_page_url`), `detail_status` = values(`detail_status`), `detail_fetched_on` = values(`detail_fetched_on`), `detail_failed_on` = values(`detail_failed_on`), `detail_error` = values(`detail_error`), `raw_content_text` = values(`raw_content_text`), `magnet_text` = values(`magnet_text`), `info_hash_text` = values(`info_hash_text`), `content_hash` = values(`content_hash`), `last_seen_time` = values(`last_seen_time`), `updated_on` = values(`updated_on`)"
	if _, err := m.conn.ExecCtx(ctx, query, data.ThreadTID, data.ThreadURL, data.ThreadTitle, data.AuthorName, data.PostTime, data.SourceForumID, data.SourcePage, data.SourcePageURL, data.DetailStatus, data.DetailFetchedOn, data.DetailFailedOn, data.DetailError, data.RawContentText, data.MagnetText, data.InfoHashText, data.ContentHash, data.LastSeenTime, data.CreatedOn, data.UpdatedOn); err != nil {
		return nil, err
	}
	return m.FindOneByThreadTID(ctx, data.ThreadTID)
}

func applyShSehuatangThreadListFilter(builder squirrel.SelectBuilder, filter ShSehuatangThreadListFilter) squirrel.SelectBuilder {
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		builder = builder.Where(squirrel.Or{
			squirrel.Like{"thread_title": like},
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
	if filter.HasMagnetSet {
		if filter.HasMagnet {
			builder = builder.Where(squirrel.Expr("TRIM(`magnet_text`) <> ''"))
		} else {
			builder = builder.Where(squirrel.Expr("TRIM(`magnet_text`) = ''"))
		}
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
