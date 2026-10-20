package common

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	DefaultAggPageSize = 100
)

func AtoiDefault(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func ClampPageSize(size int) int {
	if size <= 0 {
		return DefaultAggPageSize
	}
	if size > MaxPageSize {
		return MaxPageSize
	}
	return size
}

func BucketListOptionalInt64(c *gin.Context, key string) (*int64, string) {
	raw, ok := c.GetQuery(key)
	if !ok {
		return nil, ""
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ""
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, raw
	}
	return &value, raw
}
