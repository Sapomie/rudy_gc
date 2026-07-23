package dailybestremotex

import (
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestIsActiveFetchRunDuplicate(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "active run constraint",
			err: &mysql.MySQLError{
				Number:  1062,
				Message: "Duplicate entry '2026-07-22:1' for key 'd_dailybest_fetch_run.uidx_dailybest_active_run'",
			},
			want: true,
		},
		{
			name: "other unique constraint",
			err: &mysql.MySQLError{
				Number:  1062,
				Message: "Duplicate entry 'rank-key' for key 'c_rank.rank_key'",
			},
		},
		{name: "other error", err: errors.New("database unavailable")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isActiveFetchRunDuplicate(test.err); got != test.want {
				t.Fatalf("isActiveFetchRunDuplicate()=%v want=%v", got, test.want)
			}
		})
	}
}
