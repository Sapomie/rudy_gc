package handler

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const bytesPerGB = 1024 * 1024 * 1024

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

func gbToBytes(v float64) int64 {
	if v <= 0 {
		return 0
	}
	return int64(math.Round(v * bytesPerGB))
}

func minutesToSeconds(v int64) int64 {
	if v <= 0 {
		return 0
	}
	return v * 60
}

func parseOptionalDateUnixStart(raw string) (int64, bool, error) {
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

func parseOptionalDateUnixEnd(raw string) (int64, bool, error) {
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
