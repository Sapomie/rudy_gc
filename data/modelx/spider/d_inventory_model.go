package spider

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ DInventoryModel = (*customDInventoryModel)(nil)

type (
	// DInventoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDInventoryModel.
	DInventoryModel interface {
		dInventoryModel
		withSession(session sqlx.Session) DInventoryModel
	}

	customDInventoryModel struct {
		*defaultDInventoryModel
	}
)

// NewDInventoryModel returns a model for the database table.
func NewDInventoryModel(conn sqlx.SqlConn) DInventoryModel {
	return &customDInventoryModel{
		defaultDInventoryModel: newDInventoryModel(conn),
	}
}

func (m *customDInventoryModel) withSession(session sqlx.Session) DInventoryModel {
	return NewDInventoryModel(sqlx.NewSqlConnFromSession(session))
}
