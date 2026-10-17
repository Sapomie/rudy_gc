package dailybestremotex

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

type sourceConfigRow struct {
	JavAddress string `db:"jav_address"`
	CreatedOn  int64  `db:"created_on"`
	UpdatedOn  int64  `db:"updated_on"`
}

func (r *Repository) sourceConfigConn() (sqlx.SqlConn, error) {
	db, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	return sqlx.NewSqlConnFromDB(db), nil
}

func (r *Repository) FindSourceConfig(ctx context.Context) (dailybesttypes.SourceConfig, error) {
	conn, err := r.sourceConfigConn()
	if err != nil {
		return dailybesttypes.SourceConfig{}, err
	}
	var row sourceConfigRow
	err = conn.QueryRowCtx(ctx, &row, "SELECT jav_address, created_on, updated_on FROM d_dailybest_source_config WHERE id = ?", 1)
	if errors.Is(err, sqlx.ErrNotFound) {
		return dailybesttypes.SourceConfig{}, dailybesttypes.ErrNotFound
	}
	if err != nil {
		return dailybesttypes.SourceConfig{}, err
	}
	return dailybesttypes.SourceConfig{JavAddress: row.JavAddress, CreatedOn: row.CreatedOn, UpdatedOn: row.UpdatedOn}, nil
}

func (r *Repository) InsertSourceConfigIfMissing(ctx context.Context, address string) error {
	conn, err := r.sourceConfigConn()
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	_, err = conn.ExecCtx(ctx, `INSERT INTO d_dailybest_source_config (id, jav_address, created_on, updated_on)
 VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = id`, 1, address, now, now)
	return err
}

func (r *Repository) SaveSourceConfig(ctx context.Context, address string) (dailybesttypes.SourceConfig, error) {
	conn, err := r.sourceConfigConn()
	if err != nil {
		return dailybesttypes.SourceConfig{}, err
	}
	now := time.Now().Unix()
	_, err = conn.ExecCtx(ctx, `INSERT INTO d_dailybest_source_config (id, jav_address, created_on, updated_on)
 VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE jav_address = VALUES(jav_address), updated_on = VALUES(updated_on)`, 1, address, now, now)
	if err != nil {
		return dailybesttypes.SourceConfig{}, err
	}
	return r.FindSourceConfig(ctx)
}
