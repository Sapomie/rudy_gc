package modelx

import (
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strconv"
)

var ErrNotFound = sqlx.ErrNotFound

var ErrInvalidData = errors.New("invalid data")

// makeFakeNameMovieColumn 通过格式化字符串生成一个伪名字
func makeFakeNameMovieColumn(movieId, columnId int64) string {
	return fmt.Sprintf("%05d%06d", columnId, movieId)
}

// decodeFakeNameMovieColumn 从伪名字中解码出 movieId 和 columnId
func decodeFakeNameMovieColumn(name string) (int64, int64, error) {
	const (
		columnLength = 5
	)

	if len(name) != 11 {
		return 0, 0, errors.New("illegal length of fake name")
	}

	cId, err := strconv.ParseInt(name[:columnLength], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse columnId: %w", err)
	}

	mId, err := strconv.ParseInt(name[columnLength:], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse movieId: %w", err)
	}

	return mId, cId, nil
}
