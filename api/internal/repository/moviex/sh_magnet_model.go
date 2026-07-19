package moviex

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/internal/consts"
	moviedomain "rudy_gc/internal/domain/movie"
)

type (
	ShMagnetModel interface {
		Upsert(ctx context.Context, data *ShMagnet) error
		FindOne(ctx context.Context, id int64) (*ShMagnet, error)
		ListByIDs(ctx context.Context, ids []int64) ([]*ShMagnet, error)
		ListByMovieKey(ctx context.Context, movieJavID string, movieName string) ([]*ShMagnet, error)
		ListByFC2Key(ctx context.Context, fc2Key string) ([]*ShMagnet, error)
		ListMissingMovieJavIDByMovieName(ctx context.Context, movieName string) ([]*ShMagnet, error)
		UpdateMovieFields(ctx context.Context, id int64, movieJavID string, movieName string, updatedOn int64) error
		RefreshDownloadStatesByInfoHashes(ctx context.Context, infoHashes ...string) error
		ListPage(ctx context.Context, offset, limit int64, sortField string, sortOrder string, filter ShMagnetListFilter) ([]*ShMagnet, error)
		CountPage(ctx context.Context, filter ShMagnetListFilter) (int64, error)
		CountAll(ctx context.Context) (int64, error)
		TableName() string
	}

	defaultShMagnetModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ShMagnet struct {
		Id            int64  `db:"id"`
		ThreadTID     string `db:"thread_tid"`
		MovieJavId    string `db:"movie_jav_id"`
		MovieName     string `db:"movie_name"`
		MovieType     string `db:"movie_type"`
		FC2Key        string `db:"fc2_key"`
		SourceForumID string `db:"source_forum_id"`
		PostTime      int64  `db:"post_time"`
		MagnetURL     string `db:"magnet_url"`
		InfoHash      string `db:"info_hash"`
		CreatedOn     int64  `db:"created_on"`
		UpdatedOn     int64  `db:"updated_on"`
		ThreadTitle   string `db:"thread_title"`
		ThreadURL     string `db:"thread_url"`
		DownloadState int64  `db:"download_state"`
	}

	ShMagnetListFilter struct {
		Keyword         string
		MovieJavID      string
		InfoHash        string
		MovieType       string
		DownloadState   string
		SourceForumID   string
		PostTimeFrom    int64
		HasPostTimeFrom bool
		PostTimeTo      int64
		HasPostTimeTo   bool
	}
)

const shMagnetSelectRows = `
	m.id, m.thread_tid, m.movie_jav_id, m.movie_name, m.movie_type,
	m.source_forum_id, m.post_time, m.magnet_url, m.info_hash,
	m.download_state, m.created_on, m.updated_on,
	coalesce(t.thread_title, '') as thread_title,
	coalesce(t.thread_url, '') as thread_url,
	coalesce(fm.fc2_key, '') as fc2_key`

func NewShMagnetModel(conn sqlx.SqlConn) ShMagnetModel {
	return &defaultShMagnetModel{conn: conn, table: "`sh_magnet`"}
}

func (m *defaultShMagnetModel) TableName() string { return m.table }

func (m *defaultShMagnetModel) Upsert(ctx context.Context, data *ShMagnet) error {
	if data == nil || strings.TrimSpace(data.InfoHash) == "" {
		return nil
	}
	query := "insert into " + m.table + ` (
		thread_tid, movie_jav_id, movie_name, movie_type, source_forum_id,
		post_time, magnet_url, info_hash, created_on, updated_on
	) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	on duplicate key update
		thread_tid = case when trim(values(thread_tid)) <> '' then values(thread_tid) else thread_tid end,
		movie_jav_id = case when trim(values(movie_jav_id)) <> '' then values(movie_jav_id) else movie_jav_id end,
		movie_name = case when trim(values(movie_name)) <> '' then values(movie_name) else movie_name end,
		movie_type = case when trim(values(movie_type)) <> '' then values(movie_type) else movie_type end,
		source_forum_id = case when trim(values(source_forum_id)) <> '' then values(source_forum_id) else source_forum_id end,
		post_time = case when values(post_time) > 0 then values(post_time) else post_time end,
		magnet_url = case when trim(values(magnet_url)) <> '' then values(magnet_url) else magnet_url end,
		updated_on = greatest(updated_on, values(updated_on))`
	_, err := m.conn.ExecCtx(ctx, query,
		data.ThreadTID, data.MovieJavId, data.MovieName, data.MovieType, data.SourceForumID,
		data.PostTime, data.MagnetURL, strings.ToUpper(strings.TrimSpace(data.InfoHash)), data.CreatedOn, data.UpdatedOn,
	)
	if err != nil {
		return err
	}
	return m.RefreshDownloadStatesByInfoHashes(ctx, data.InfoHash)
}

func (m *defaultShMagnetModel) FindOne(ctx context.Context, id int64) (*ShMagnet, error) {
	var row ShMagnet
	query := "select " + shMagnetSelectRows + " from " + m.table + " m left join `sh_thread` t on t.thread_tid = m.thread_tid left join `fc2_movie` fm on fm.id = t.fc2_movie_id where m.id = ? limit 1"
	if err := m.conn.QueryRowCtx(ctx, &row, query, id); err != nil {
		return nil, err
	}
	return &row, nil
}

func (m *defaultShMagnetModel) ListByIDs(ctx context.Context, ids []int64) ([]*ShMagnet, error) {
	uniqueIDs := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	if len(uniqueIDs) == 0 {
		return []*ShMagnet{}, nil
	}

	builder := squirrel.
		Select(shMagnetSelectRows).
		From(m.table + " m").
		LeftJoin("`sh_thread` t ON t.thread_tid = m.thread_tid").
		LeftJoin("`fc2_movie` fm ON fm.id = t.fc2_movie_id").
		Where(squirrel.Eq{"m.id": uniqueIDs}).
		OrderBy("m.id ASC")
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	rows := make([]*ShMagnet, 0, len(uniqueIDs))
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*ShMagnet{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultShMagnetModel) ListByMovieKey(ctx context.Context, movieJavID string, movieName string) ([]*ShMagnet, error) {
	movieJavID = strings.TrimSpace(movieJavID)
	movieName = strings.TrimSpace(movieName)
	if movieJavID == "" && movieName == "" {
		return []*ShMagnet{}, nil
	}

	query := "select " + shMagnetSelectRows + " from " + m.table + " m left join `sh_thread` t on t.thread_tid = m.thread_tid left join `fc2_movie` fm on fm.id = t.fc2_movie_id"
	args := make([]any, 0, 2)
	if movieJavID != "" && movieName != "" {
		query += " where m.movie_jav_id = ? or m.movie_name = ?"
		args = append(args, movieJavID, movieName)
	} else if movieJavID != "" {
		query += " where m.movie_jav_id = ?"
		args = append(args, movieJavID)
	} else {
		query += " where m.movie_name = ?"
		args = append(args, movieName)
	}
	query += " order by m.post_time desc, m.updated_on desc, m.id desc"

	rows := make([]*ShMagnet, 0)
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*ShMagnet{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultShMagnetModel) ListByFC2Key(ctx context.Context, fc2Key string) ([]*ShMagnet, error) {
	fc2Key = strings.TrimSpace(fc2Key)
	if fc2Key == "" {
		return []*ShMagnet{}, nil
	}

	query := "select " + shMagnetSelectRows + " from " + m.table + " m left join `sh_thread` t on t.thread_tid = m.thread_tid left join `fc2_movie` fm on fm.id = t.fc2_movie_id where fm.fc2_key = ? order by m.post_time desc, m.updated_on desc, m.id desc"
	rows := make([]*ShMagnet, 0)
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, fc2Key); err != nil {
		if err == sqlx.ErrNotFound {
			return []*ShMagnet{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultShMagnetModel) ListMissingMovieJavIDByMovieName(ctx context.Context, movieName string) ([]*ShMagnet, error) {
	movieName = strings.TrimSpace(movieName)
	if movieName == "" {
		return []*ShMagnet{}, nil
	}

	query := "select " + shMagnetSelectRows + " from " + m.table + " m left join `sh_thread` t on t.thread_tid = m.thread_tid left join `fc2_movie` fm on fm.id = t.fc2_movie_id where m.movie_name = ? and m.movie_jav_id = '' order by m.id asc"
	rows := make([]*ShMagnet, 0)
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, movieName); err != nil {
		if err == sqlx.ErrNotFound {
			return []*ShMagnet{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultShMagnetModel) UpdateMovieFields(ctx context.Context, id int64, movieJavID string, movieName string, updatedOn int64) error {
	_, err := m.conn.ExecCtx(ctx, "update "+m.table+" set movie_jav_id = ?, movie_name = ?, updated_on = ? where id = ?", strings.TrimSpace(movieJavID), strings.TrimSpace(movieName), updatedOn, id)
	return err
}

func (m *defaultShMagnetModel) RefreshDownloadStatesByInfoHashes(ctx context.Context, infoHashes ...string) error {
	hashes := make([]string, 0, len(infoHashes))
	seen := make(map[string]struct{}, len(infoHashes))
	for _, value := range infoHashes {
		hash := strings.ToUpper(strings.TrimSpace(value))
		if hash == "" {
			continue
		}
		if _, exists := seen[hash]; exists {
			continue
		}
		seen[hash] = struct{}{}
		hashes = append(hashes, hash)
	}
	if len(hashes) == 0 {
		return nil
	}

	downloadState := squirrel.Expr(`case
		when lower(trim(m.movie_type)) = ? then case
			when exists (
				select 1 from fc2_media fm_active
				where fm_active.source_torrent_hash = m.info_hash and fm_active.is_removed = ?
			) then ?
			when exists (
				select 1 from fc2_media fm_any
				where fm_any.source_torrent_hash = m.info_hash
			) then ?
			else ?
		end
		else case
			when exists (
				select 1 from w_media wm_active
				where wm_active.source_type = ? and wm_active.source_torrent_hash = m.info_hash and wm_active.is_removed = ?
			) then ?
			when exists (
				select 1 from w_media wm_any
				where wm_any.source_type = ? and wm_any.source_torrent_hash = m.info_hash
			) then ?
			else ?
		end
	end`,
		moviedomain.SehuatangThreadMovieTypeFC2PPV,
		consts.FilmIsNotRemoved, consts.SHMagnetDownloadStateDownloaded,
		consts.SHMagnetDownloadStateRemoved, consts.SHMagnetDownloadStateMissing,
		consts.WMediaSourceNative, consts.FilmIsNotRemoved, consts.SHMagnetDownloadStateDownloaded,
		consts.WMediaSourceNative, consts.SHMagnetDownloadStateRemoved, consts.SHMagnetDownloadStateMissing,
	)
	query, args, err := squirrel.Update(m.table+" m").
		Set("download_state", downloadState).
		Where(squirrel.Eq{"m.info_hash": hashes}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = m.conn.ExecCtx(ctx, query, args...)
	return err
}

func (m *defaultShMagnetModel) ListPage(ctx context.Context, offset, limit int64, sortField string, sortOrder string, filter ShMagnetListFilter) ([]*ShMagnet, error) {
	builder := applyShMagnetListFilter(squirrel.
		Select(shMagnetSelectRows).
		From(m.table+" m").
		LeftJoin("`sh_thread` t ON t.thread_tid = m.thread_tid").
		LeftJoin("`fc2_movie` fm ON fm.id = t.fc2_movie_id"), filter).
		OrderBy(shMagnetOrderBy(sortField, sortOrder), "m.id DESC")
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
	rows := make([]*ShMagnet, 0)
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*ShMagnet{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultShMagnetModel) CountPage(ctx context.Context, filter ShMagnetListFilter) (int64, error) {
	builder := applyShMagnetListFilter(squirrel.
		Select("COUNT(*)").
		From(m.table+" m").
		LeftJoin("`sh_thread` t ON t.thread_tid = m.thread_tid"), filter)
	query, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}
	var count int64
	if err := m.conn.QueryRowCtx(ctx, &count, query, args...); err != nil {
		return 0, err
	}
	return count, nil
}

func applyShMagnetListFilter(builder squirrel.SelectBuilder, filter ShMagnetListFilter) squirrel.SelectBuilder {
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		builder = builder.Where(squirrel.Or{
			squirrel.Like{"m.movie_name": like},
			squirrel.Like{"m.movie_jav_id": like},
			squirrel.Like{"m.info_hash": like},
			squirrel.Like{"m.thread_tid": like},
			squirrel.Like{"t.thread_title": like},
		})
	}
	if movieTypes := moviedomain.NormalizeSehuatangThreadMovieTypes(filter.MovieType); len(movieTypes) == 1 {
		builder = builder.Where(squirrel.Eq{"m.movie_type": movieTypes[0]})
	}
	if downloadState := strings.TrimSpace(filter.DownloadState); downloadState == "0" || downloadState == "1" || downloadState == "2" {
		builder = builder.Where(squirrel.Eq{"m.download_state": downloadState})
	}
	if movieJavID := strings.TrimSpace(filter.MovieJavID); movieJavID != "" {
		builder = builder.Where(squirrel.Like{"m.movie_jav_id": "%" + movieJavID + "%"})
	}
	if infoHash := strings.TrimSpace(filter.InfoHash); infoHash != "" {
		builder = builder.Where(squirrel.Like{"m.info_hash": "%" + strings.ToUpper(infoHash) + "%"})
	}
	if sourceForumID := strings.TrimSpace(filter.SourceForumID); sourceForumID != "" {
		builder = builder.Where(squirrel.Eq{"m.source_forum_id": sourceForumID})
	}
	if filter.HasPostTimeFrom {
		builder = builder.Where(squirrel.GtOrEq{"m.post_time": filter.PostTimeFrom})
	}
	if filter.HasPostTimeTo {
		builder = builder.Where(squirrel.LtOrEq{"m.post_time": filter.PostTimeTo})
	}
	return builder
}

func shMagnetOrderBy(sortField string, sortOrder string) string {
	column := "m.post_time"
	switch strings.TrimSpace(sortField) {
	case "movie_name":
		column = "m.movie_name"
	case "updated_on":
		column = "m.updated_on"
	}
	direction := "DESC"
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		direction = "ASC"
	}
	return column + " " + direction
}

func (m *defaultShMagnetModel) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, "select count(*) from "+m.table)
	return count, err
}
