package moviex

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const fc2ScStatRows = "`id`,`fc2_movie_id`,`fc2_key`,`movie_name`,`sc_event_times`,`sc_times`,`come_times`,`last_sc_time`,`post_time`,`media_birth_time`,`created_on`,`updated_on`"

const (
	cacheRudyGcFC2ScStatIDPrefix      = "cache:rudyGc:fc2ScStat:id:"
	cacheRudyGcFC2ScStatMovieIDPrefix = "cache:rudyGc:fc2ScStat:movieId:"
	cacheRudyGcFC2ScStatKeyPrefix     = "cache:rudyGc:fc2ScStat:key:"
)

type (
	fc2ScStatModel interface {
		Insert(ctx context.Context, data *FC2ScStat) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*FC2ScStat, error)
		FindOneByMovieID(ctx context.Context, movieID int64) (*FC2ScStat, error)
		FindOneByKey(ctx context.Context, key string) (*FC2ScStat, error)
		Update(ctx context.Context, data *FC2ScStat) error
		Delete(ctx context.Context, id int64) error
	}

	defaultFC2ScStatModel struct {
		sqlc.CachedConn
		table string
	}

	FC2ScStat struct {
		ID             int64  `db:"id"`
		FC2MovieID     int64  `db:"fc2_movie_id"`
		FC2Key         string `db:"fc2_key"`
		MovieName      string `db:"movie_name"`
		ScEventTimes   int64  `db:"sc_event_times"`
		ScTimes        int64  `db:"sc_times"`
		ComeTimes      int64  `db:"come_times"`
		LastScTime     int64  `db:"last_sc_time"`
		PostTime       int64  `db:"post_time"`
		MediaBirthTime int64  `db:"media_birth_time"`
		CreatedOn      int64  `db:"created_on"`
		UpdatedOn      int64  `db:"updated_on"`
	}
)

func newFC2ScStatModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) *defaultFC2ScStatModel {
	return &defaultFC2ScStatModel{CachedConn: sqlc.NewConn(conn, c, opts...), table: "`fc2_sc_stat`"}
}

func (m *defaultFC2ScStatModel) cacheKeys(data *FC2ScStat) []string {
	if data == nil {
		return nil
	}
	return []string{
		fmt.Sprintf("%s%d", cacheRudyGcFC2ScStatIDPrefix, data.ID),
		fmt.Sprintf("%s%d", cacheRudyGcFC2ScStatMovieIDPrefix, data.FC2MovieID),
		fmt.Sprintf("%s%s", cacheRudyGcFC2ScStatKeyPrefix, data.FC2Key),
	}
}

func (m *defaultFC2ScStatModel) Insert(ctx context.Context, data *FC2ScStat) (sql.Result, error) {
	return m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := "INSERT INTO " + m.table + " (`fc2_movie_id`,`fc2_key`,`movie_name`,`sc_event_times`,`sc_times`,`come_times`,`last_sc_time`,`post_time`,`media_birth_time`,`created_on`,`updated_on`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		return conn.ExecCtx(ctx, query, data.FC2MovieID, data.FC2Key, data.MovieName, data.ScEventTimes, data.ScTimes, data.ComeTimes, data.LastScTime, data.PostTime, data.MediaBirthTime, data.CreatedOn, data.UpdatedOn)
	}, m.cacheKeys(data)...)
}

func (m *defaultFC2ScStatModel) FindOne(ctx context.Context, id int64) (*FC2ScStat, error) {
	var row FC2ScStat
	cacheKey := fmt.Sprintf("%s%d", cacheRudyGcFC2ScStatIDPrefix, id)
	err := m.QueryRowCtx(ctx, &row, cacheKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		return conn.QueryRowCtx(ctx, v, "SELECT "+fc2ScStatRows+" FROM "+m.table+" WHERE `id` = ? LIMIT 1", id)
	})
	if err == nil {
		return &row, nil
	}
	if err == sqlc.ErrNotFound {
		return nil, ErrNotFound
	}
	return nil, err
}

func (m *defaultFC2ScStatModel) FindOneByMovieID(ctx context.Context, movieID int64) (*FC2ScStat, error) {
	var row FC2ScStat
	cacheKey := fmt.Sprintf("%s%d", cacheRudyGcFC2ScStatMovieIDPrefix, movieID)
	err := m.QueryRowIndexCtx(ctx, &row, cacheKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (any, error) {
		if err := conn.QueryRowCtx(ctx, &row, "SELECT "+fc2ScStatRows+" FROM "+m.table+" WHERE `fc2_movie_id` = ? LIMIT 1", movieID); err != nil {
			return nil, err
		}
		return row.ID, nil
	}, m.queryPrimary)
	if err == nil {
		return &row, nil
	}
	if err == sqlc.ErrNotFound {
		return nil, ErrNotFound
	}
	return nil, err
}

func (m *defaultFC2ScStatModel) FindOneByKey(ctx context.Context, key string) (*FC2ScStat, error) {
	var row FC2ScStat
	cacheKey := fmt.Sprintf("%s%s", cacheRudyGcFC2ScStatKeyPrefix, key)
	err := m.QueryRowIndexCtx(ctx, &row, cacheKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (any, error) {
		if err := conn.QueryRowCtx(ctx, &row, "SELECT "+fc2ScStatRows+" FROM "+m.table+" WHERE `fc2_key` = ? LIMIT 1", key); err != nil {
			return nil, err
		}
		return row.ID, nil
	}, m.queryPrimary)
	if err == nil {
		return &row, nil
	}
	if err == sqlc.ErrNotFound {
		return nil, ErrNotFound
	}
	return nil, err
}

func (m *defaultFC2ScStatModel) Update(ctx context.Context, data *FC2ScStat) error {
	old, err := m.FindOne(ctx, data.ID)
	if err != nil {
		return err
	}
	keys := append(m.cacheKeys(old), m.cacheKeys(data)...)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := "UPDATE " + m.table + " SET `fc2_movie_id`=?,`fc2_key`=?,`movie_name`=?,`sc_event_times`=?,`sc_times`=?,`come_times`=?,`last_sc_time`=?,`post_time`=?,`media_birth_time`=?,`created_on`=?,`updated_on`=? WHERE `id`=?"
		return conn.ExecCtx(ctx, query, data.FC2MovieID, data.FC2Key, data.MovieName, data.ScEventTimes, data.ScTimes, data.ComeTimes, data.LastScTime, data.PostTime, data.MediaBirthTime, data.CreatedOn, data.UpdatedOn, data.ID)
	}, keys...)
	return err
}

func (m *defaultFC2ScStatModel) Delete(ctx context.Context, id int64) error {
	row, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		return conn.ExecCtx(ctx, "DELETE FROM "+m.table+" WHERE `id` = ?", id)
	}, m.cacheKeys(row)...)
	return err
}

func (m *defaultFC2ScStatModel) formatPrimary(primary any) string {
	return fmt.Sprintf("%s%v", cacheRudyGcFC2ScStatIDPrefix, primary)
}

func (m *defaultFC2ScStatModel) queryPrimary(ctx context.Context, conn sqlx.SqlConn, v, primary any) error {
	return conn.QueryRowCtx(ctx, v, "SELECT "+fc2ScStatRows+" FROM "+m.table+" WHERE `id` = ? LIMIT 1", primary)
}
