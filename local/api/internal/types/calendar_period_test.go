package types

import (
	"testing"
	"time"
)

func TestCalendarPeriodBoundaries(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	for _, tc := range []struct{ kind, key, start, end string }{
		{"year", "year:2026", "2026-01-01", "2026-12-31"},
		{"quarter", "quarter:2026-Q4", "2026-10-01", "2026-12-31"},
		{"month", "month:2024-02", "2024-02-01", "2024-02-29"},
		{"week", "week:2020-W53", "2020-12-28", "2021-01-03"},
		{"week", "week:2021-W01", "2021-01-04", "2021-01-10"},
		{"week", "week:2026-W01", "2025-12-29", "2026-01-04"},
		{"day", "day:2026-10-11", "2026-10-11", "2026-10-11"},
	} {
		t.Run(tc.key, func(t *testing.T) {
			p, err := ParseCalendarPeriod(tc.kind, tc.key, location)
			if err != nil {
				t.Fatal(err)
			}
			if p.Start.Format(time.DateOnly) != tc.start || p.End.Format(time.DateOnly) != tc.end || p.End.Format(time.TimeOnly) != "23:59:59" {
				t.Fatalf("周期范围错误: %+v", p)
			}
			for _, date := range []time.Time{p.Start, p.End} {
				got, err := CalendarPeriodAt(tc.kind, date)
				if err != nil || got.Key != tc.key {
					t.Fatalf("边界归属错误: %+v, %v", got, err)
				}
			}
		})
	}
}

func TestCalendarPeriodRejectsInvalidKeys(t *testing.T) {
	for _, tc := range []struct{ kind, key string }{
		{"week", "week:2021-W53"}, {"week", "week:2020-W00"}, {"week", "week:2020-W54"},
		{"week", "week:2026-W1"}, {"month", "month:2026-13"}, {"quarter", "quarter:2026-Q5"},
		{"day", "day:2026-02-29"}, {"day", "day:2024-02-29-extra"}, {"year", "day:2026-01-01"},
		{"quarter", "quarter:2026-Q1-extra"}, {"invalid", "invalid:2026"},
	} {
		if _, err := ParseCalendarPeriod(tc.kind, tc.key, time.UTC); err == nil {
			t.Errorf("接受了无效周期: %+v", tc)
		}
	}
}
