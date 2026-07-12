package handler

import (
	"testing"
	"time"
)

func TestParseSHBucketDate(t *testing.T) {
	start, raw, err := parseSHBucketDate(" 2026-07-01 ", "起始时间", false)
	if err != nil {
		t.Fatal(err)
	}
	if raw != "2026-07-01" {
		t.Fatalf("raw = %q, want 2026-07-01", raw)
	}
	wantStart := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.Local).Unix()
	if start == nil || *start != wantStart {
		t.Fatalf("start = %v, want %d", start, wantStart)
	}

	end, _, err := parseSHBucketDate("2026-07-31", "结束时间", true)
	if err != nil {
		t.Fatal(err)
	}
	wantEnd := time.Date(2026, time.July, 31, 23, 59, 59, 0, time.Local).Unix()
	if end == nil || *end != wantEnd {
		t.Fatalf("end = %v, want %d", end, wantEnd)
	}
}

func TestParseSHBucketDateEmptyAndInvalid(t *testing.T) {
	value, raw, err := parseSHBucketDate("", "起始时间", false)
	if err != nil || value != nil || raw != "" {
		t.Fatalf("empty result = (%v, %q, %v)", value, raw, err)
	}
	if _, _, err := parseSHBucketDate("2026/07/01", "起始时间", false); err == nil {
		t.Fatal("invalid date should return error")
	}
}
