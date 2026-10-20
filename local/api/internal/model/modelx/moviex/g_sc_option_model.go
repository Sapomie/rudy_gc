package moviex

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	GScOptionModel interface {
		FindOne(ctx context.Context, id int64) (*GScOption, error)
		ListEnabled(ctx context.Context) ([]*GScOption, error)
		ListAll(ctx context.Context) ([]*GScOption, error)
		Insert(ctx context.Context, data *GScOption) error
		Update(ctx context.Context, data *GScOption) error
		Disable(ctx context.Context, id, updatedOn int64) error
	}

	defaultGScOptionModel struct {
		conn  sqlx.SqlConn
		table string
	}

	GScOption struct {
		ID              int64  `db:"id" json:"id"`
		Name            string `db:"name" json:"name"`
		SortOrder       int64  `db:"sort_order" json:"sortOrder"`
		IsEnabled       int64  `db:"is_enabled" json:"isEnabled"`
		ScEventTimes    int64  `db:"sc_event_times" json:"scEventTimes"`
		LastScEventTime int64  `db:"last_sc_event_time" json:"lastScEventTime"`
		CreatedOn       int64  `db:"created_on" json:"createdOn"`
		UpdatedOn       int64  `db:"updated_on" json:"updatedOn"`
	}
)

const (
	gScKindOptionTable   = "`g_sc_option_kind`"
	gScFGOptionTable     = "`g_sc_option_fg`"
	gScVesselOptionTable = "`g_sc_option_vessl`"
	gScOptionRows        = "`id`,`name`,`sort_order`,`is_enabled`,`sc_event_times`,`last_sc_event_time`,`created_on`,`updated_on`"
)

func NewGScKindOptionModel(conn sqlx.SqlConn) GScOptionModel {
	return newGScOptionModel(conn, gScKindOptionTable)
}

func NewGScFGOptionModel(conn sqlx.SqlConn) GScOptionModel {
	return newGScOptionModel(conn, gScFGOptionTable)
}

func NewGScVesselOptionModel(conn sqlx.SqlConn) GScOptionModel {
	return newGScOptionModel(conn, gScVesselOptionTable)
}

func newGScOptionModel(conn sqlx.SqlConn, table string) GScOptionModel {
	return &defaultGScOptionModel{conn: conn, table: table}
}

func (m *defaultGScOptionModel) FindOne(ctx context.Context, id int64) (*GScOption, error) {
	if id <= 0 {
		return nil, ErrNotFound
	}
	var row GScOption
	if err := m.conn.QueryRowCtx(ctx, &row, "SELECT "+gScOptionRows+" FROM "+m.table+" WHERE `id` = ? LIMIT 1", id); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultGScOptionModel) ListEnabled(ctx context.Context) ([]*GScOption, error) {
	return m.list(ctx, true)
}

func (m *defaultGScOptionModel) ListAll(ctx context.Context) ([]*GScOption, error) {
	return m.list(ctx, false)
}

func (m *defaultGScOptionModel) list(ctx context.Context, enabledOnly bool) ([]*GScOption, error) {
	var rows []*GScOption
	query := "SELECT " + gScOptionRows + " FROM " + m.table
	if enabledOnly {
		query += " WHERE `is_enabled` = 1"
	}
	query += " ORDER BY `sort_order` ASC, `name` ASC, `id` ASC"
	if err := m.conn.QueryRowsCtx(ctx, &rows, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*GScOption{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultGScOptionModel) Insert(ctx context.Context, data *GScOption) error {
	if data == nil {
		return errors.New("SC 选项不能为空")
	}
	result, err := m.conn.ExecCtx(ctx, "INSERT INTO "+m.table+" (`name`,`sort_order`,`is_enabled`,`created_on`,`updated_on`) VALUES (?, ?, ?, ?, ?)",
		data.Name, data.SortOrder, data.IsEnabled, data.CreatedOn, data.UpdatedOn)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	data.ID = id
	return nil
}

func (m *defaultGScOptionModel) Update(ctx context.Context, data *GScOption) error {
	if data == nil || data.ID <= 0 {
		return ErrNotFound
	}
	_, err := m.conn.ExecCtx(ctx, "UPDATE "+m.table+" SET `name` = ?, `sort_order` = ?, `is_enabled` = ?, `updated_on` = ? WHERE `id` = ?",
		data.Name, data.SortOrder, data.IsEnabled, data.UpdatedOn, data.ID)
	return err
}

func (m *defaultGScOptionModel) Disable(ctx context.Context, id, updatedOn int64) error {
	if id <= 0 {
		return ErrNotFound
	}
	_, err := m.conn.ExecCtx(ctx, "UPDATE "+m.table+" SET `is_enabled` = 0, `updated_on` = ? WHERE `id` = ?", updatedOn, id)
	return err
}
