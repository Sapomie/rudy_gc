package dailybest

import "errors"

var (
	ErrBlankPage = errors.New("dailybest blank page")
	ErrNotFound  = errors.New("dailybest data not found")
)
