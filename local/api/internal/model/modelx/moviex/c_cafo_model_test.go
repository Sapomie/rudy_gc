package moviex

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type cCafoTestResult struct{}

func (cCafoTestResult) LastInsertId() (int64, error) { return 0, nil }
func (cCafoTestResult) RowsAffected() (int64, error) { return 1, nil }

type cCafoTestConn struct {
	sqlx.SqlConn
	queries []string
	args    [][]any
}

type cCafoNoopCache struct{}

func (cCafoNoopCache) Del(...string) error                            { return nil }
func (cCafoNoopCache) DelCtx(context.Context, ...string) error        { return nil }
func (cCafoNoopCache) Get(string, any) error                          { return nil }
func (cCafoNoopCache) GetCtx(context.Context, string, any) error      { return nil }
func (cCafoNoopCache) IsNotFound(error) bool                          { return false }
func (cCafoNoopCache) Set(string, any) error                          { return nil }
func (cCafoNoopCache) SetCtx(context.Context, string, any) error      { return nil }
func (cCafoNoopCache) SetWithExpire(string, any, time.Duration) error { return nil }
func (cCafoNoopCache) SetWithExpireCtx(context.Context, string, any, time.Duration) error {
	return nil
}
func (cCafoNoopCache) Take(val any, _ string, query func(any) error) error {
	return query(val)
}
func (cCafoNoopCache) TakeCtx(_ context.Context, val any, _ string, query func(any) error) error {
	return query(val)
}
func (cCafoNoopCache) TakeWithExpire(val any, _ string, query func(any, time.Duration) error) error {
	return query(val, 0)
}
func (cCafoNoopCache) TakeWithExpireCtx(_ context.Context, val any, _ string, query func(any, time.Duration) error) error {
	return query(val, 0)
}

func (c *cCafoTestConn) ExecCtx(_ context.Context, query string, args ...any) (sql.Result, error) {
	c.queries = append(c.queries, query)
	c.args = append(c.args, args)
	return cCafoTestResult{}, nil
}

func TestCCafoPersonLinkWrites(t *testing.T) {
	conn := &cCafoTestConn{}
	model := &customCCafoModel{defaultCCafoModel: &defaultCCafoModel{
		CachedConn: sqlc.NewConnWithCache(conn, cCafoNoopCache{}),
		table:      "`c_cafo`",
	}}
	ctx := context.Background()

	if err := model.SetPersonID(ctx, 7, 11); err != nil {
		t.Fatalf("SetPersonID: %v", err)
	}
	if err := model.SetPersonID(ctx, 8, 0); err != nil {
		t.Fatalf("clear SetPersonID: %v", err)
	}
	if err := model.ReassignPersonIDs(ctx, 11, []int64{22, 33, 22, 11, 0}); err != nil {
		t.Fatalf("ReassignPersonIDs: %v", err)
	}

	if len(conn.queries) != 3 {
		t.Fatalf("queries = %#v", conn.queries)
	}
	if !strings.Contains(conn.queries[0], "set `person_id` = ? where `id` = ?") {
		t.Fatalf("set query = %q", conn.queries[0])
	}
	if got := conn.args[0]; len(got) != 2 || got[0] != int64(11) || got[1] != int64(7) {
		t.Fatalf("set args = %#v", got)
	}
	if got := conn.args[1]; len(got) != 2 || got[0] != nil || got[1] != int64(8) {
		t.Fatalf("clear args = %#v", got)
	}
	if !strings.Contains(conn.queries[2], "where `person_id` in (?,?)") {
		t.Fatalf("reassign query = %q", conn.queries[2])
	}
	if got := conn.args[2]; len(got) != 3 || got[0] != int64(11) || got[1] != int64(22) || got[2] != int64(33) {
		t.Fatalf("reassign args = %#v", got)
	}
}
