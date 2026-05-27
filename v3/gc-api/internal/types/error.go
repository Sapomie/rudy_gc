package types

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var ErrNotFound = sqlx.ErrNotFound

type BadRequestError struct {
	Message string
}

func (e *BadRequestError) Error() string {
	return e.Message
}
