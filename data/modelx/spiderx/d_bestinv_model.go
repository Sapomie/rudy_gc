package spiderx

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ DBestinvModel = (*customDBestinvModel)(nil)

type (
	// DBestinvModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDBestinvModel.
	DBestinvModel interface {
		dBestinvModel
		withSession(session sqlx.Session) DBestinvModel
	}

	customDBestinvModel struct {
		*defaultDBestinvModel
	}
)

// NewDBestinvModel returns a model for the database table.
func NewDBestinvModel(conn sqlx.SqlConn) DBestinvModel {
	return &customDBestinvModel{
		defaultDBestinvModel: newDBestinvModel(conn),
	}
}

func (m *customDBestinvModel) withSession(session sqlx.Session) DBestinvModel {
	return NewDBestinvModel(sqlx.NewSqlConnFromSession(session))
}
