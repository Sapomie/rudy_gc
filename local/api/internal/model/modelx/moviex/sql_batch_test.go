package moviex

import "testing"

func TestSQLValueBatchesCoversValuesBeyondPlaceholderLimit(t *testing.T) {
	values := make([]int64, mysqlStatementPlaceholderLimit+137)
	batches := sqlValueBatches(values, 2)
	if len(batches) != 2 {
		t.Fatalf("batch count = %d", len(batches))
	}
	if len(batches[0])+2 > mysqlStatementPlaceholderLimit {
		t.Fatalf("first statement placeholders = %d", len(batches[0])+2)
	}
	if got := len(batches[0]) + len(batches[1]); got != len(values) {
		t.Fatalf("merged values = %d, want %d", got, len(values))
	}
}
