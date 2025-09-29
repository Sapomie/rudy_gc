package spiderx

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ DDetailModel = (*customDDetailModel)(nil)

type (
	// DDetailModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDDetailModel.
	DDetailModel interface {
		dDetailModel
		withSession(session sqlx.Session) DDetailModel
	}

	customDDetailModel struct {
		*defaultDDetailModel
	}
)

// NewDDetailModel returns a model for the database table.
func NewDDetailModel(conn sqlx.SqlConn) DDetailModel {
	return &customDDetailModel{
		defaultDDetailModel: newDDetailModel(conn),
	}
}

func (m *customDDetailModel) withSession(session sqlx.Session) DDetailModel {
	return NewDDetailModel(sqlx.NewSqlConnFromSession(session))
}
