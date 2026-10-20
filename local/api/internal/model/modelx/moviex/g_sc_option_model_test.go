package moviex

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type gScOptionTestResult struct {
	id int64
}

func (r gScOptionTestResult) LastInsertId() (int64, error) { return r.id, nil }
func (gScOptionTestResult) RowsAffected() (int64, error)   { return 1, nil }

type gScOptionTestConn struct {
	sqlx.SqlConn
	queries []string
	args    [][]any
}

func (c *gScOptionTestConn) ExecCtx(_ context.Context, query string, args ...any) (sql.Result, error) {
	c.queries = append(c.queries, query)
	c.args = append(c.args, args)
	return gScOptionTestResult{id: 11}, nil
}

func (c *gScOptionTestConn) QueryRowCtx(_ context.Context, _ any, query string, args ...any) error {
	c.queries = append(c.queries, query)
	c.args = append(c.args, args)
	return nil
}

func (c *gScOptionTestConn) QueryRowsCtx(_ context.Context, _ any, query string, args ...any) error {
	c.queries = append(c.queries, query)
	c.args = append(c.args, args)
	return nil
}

func TestGScOptionModelsUseFixedTablesAndNameColumn(t *testing.T) {
	tests := []struct {
		name    string
		table   string
		factory func(sqlx.SqlConn) GScOptionModel
	}{
		{name: "kind", table: "`g_sc_option_kind`", factory: NewGScKindOptionModel},
		{name: "fg", table: "`g_sc_option_fg`", factory: NewGScFGOptionModel},
		{name: "vessel", table: "`g_sc_option_vessl`", factory: NewGScVesselOptionModel},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conn := &gScOptionTestConn{}
			model := test.factory(conn)
			ctx := context.Background()

			if _, err := model.FindOne(ctx, 7); err != nil {
				t.Fatalf("FindOne: %v", err)
			}
			if _, err := model.ListEnabled(ctx); err != nil {
				t.Fatalf("ListEnabled: %v", err)
			}
			if _, err := model.ListAll(ctx); err != nil {
				t.Fatalf("ListAll: %v", err)
			}

			row := &GScOption{
				Name:      "选项 A",
				SortOrder: 3,
				IsEnabled: 1,
				CreatedOn: 100,
				UpdatedOn: 101,
			}
			if err := model.Insert(ctx, row); err != nil {
				t.Fatalf("Insert: %v", err)
			}
			if row.ID != 11 {
				t.Fatalf("inserted ID = %d", row.ID)
			}
			row.Name = "选项 B"
			row.SortOrder = 8
			row.UpdatedOn = 123
			if err := model.Update(ctx, row); err != nil {
				t.Fatalf("Update: %v", err)
			}
			if err := model.Disable(ctx, row.ID, 124); err != nil {
				t.Fatalf("Disable: %v", err)
			}

			if len(conn.queries) != 6 || len(conn.args) != 6 {
				t.Fatalf("queries = %#v, args = %#v", conn.queries, conn.args)
			}
			for index, query := range conn.queries {
				if !strings.Contains(query, test.table) {
					t.Fatalf("query %d = %q, want table %s", index, query, test.table)
				}
				if strings.Contains(query, "option_type") || strings.Contains(query, "option_name") || strings.Contains(query, "option_code") {
					t.Fatalf("query %d contains legacy column: %q", index, query)
				}
			}
			if !strings.Contains(conn.queries[0], "SELECT `id`,`name`,`sort_order`,`is_enabled`,`sc_event_times`,`last_sc_event_time`,`created_on`,`updated_on`") || len(conn.args[0]) != 1 || conn.args[0][0] != int64(7) {
				t.Fatalf("find query = %q, args = %#v", conn.queries[0], conn.args[0])
			}
			if !strings.Contains(conn.queries[1], "WHERE `is_enabled` = 1") || strings.Contains(conn.queries[2], "WHERE `is_enabled` = 1") {
				t.Fatalf("list queries = %#v", conn.queries[1:3])
			}
			if !strings.Contains(conn.queries[3], "INSERT INTO") || len(conn.args[3]) != 5 || conn.args[3][0] != "选项 A" {
				t.Fatalf("insert query = %q, args = %#v", conn.queries[3], conn.args[3])
			}
			if !strings.Contains(conn.queries[4], "SET `name` = ?") || len(conn.args[4]) != 5 || conn.args[4][0] != "选项 B" {
				t.Fatalf("update query = %q, args = %#v", conn.queries[4], conn.args[4])
			}
			if !strings.Contains(conn.queries[5], "SET `is_enabled` = 0") || len(conn.args[5]) != 2 || conn.args[5][1] != int64(11) {
				t.Fatalf("disable query = %q, args = %#v", conn.queries[5], conn.args[5])
			}
		})
	}
}
