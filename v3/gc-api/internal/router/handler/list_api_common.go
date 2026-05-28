package handler

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func withFrontendPath(c *gin.Context, path string) func() {
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = path
	return func() {
		c.Request.URL.Path = originalPath
	}
}

func jsonBadRequest(c *gin.Context, prefix string, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": prefix + ": " + err.Error()})
}

func jsonServerError(c *gin.Context, prefix string, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": prefix + ": " + err.Error()})
}

func parseOptionalNonNegativeInt64(raw string) (int64, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}

	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false, err
	}
	if v < 0 {
		return 0, false, fmt.Errorf("必须是非负整数")
	}
	return v, true, nil
}

func parseOptionalNonNegativeFloat64(raw string) (float64, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}

	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false, err
	}
	if v < 0 {
		return 0, false, fmt.Errorf("必须是非负数")
	}
	return v, true, nil
}

func parseOptionalDateStart(raw string) (int64, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}

	t, err := time.ParseInLocation("2006-01-02", raw, time.Local)
	if err != nil {
		return 0, false, err
	}
	return t.Unix(), true, nil
}

func parseOptionalDateEnd(raw string) (int64, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false, nil
	}

	t, err := time.ParseInLocation("2006-01-02", raw, time.Local)
	if err != nil {
		return 0, false, err
	}
	return t.Add(24*time.Hour - time.Second).Unix(), true, nil
}

func parseOptionalDateUnixStart(raw string) (int64, bool, error) {
	return parseOptionalDateStart(raw)
}

func parseOptionalDateUnixEnd(raw string) (int64, bool, error) {
	return parseOptionalDateEnd(raw)
}

func gbToBytes(v float64) int64 {
	return int64(math.Round(v * 1024 * 1024 * 1024))
}

func minutesToSeconds(v int64) int64 {
	return v * 60
}
