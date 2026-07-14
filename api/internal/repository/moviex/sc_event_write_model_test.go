package moviex

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type scEventWriteTestResult int64

func (r scEventWriteTestResult) LastInsertId() (int64, error) { return int64(r), nil }
func (r scEventWriteTestResult) RowsAffected() (int64, error) { return 1, nil }

type scEventWriteTestSession struct {
	sqlx.Session
	execCalls int
	failAt    int
	queries   []string
}

func (s *scEventWriteTestSession) ExecCtx(_ context.Context, query string, _ ...any) (sql.Result, error) {
	s.execCalls++
	s.queries = append(s.queries, query)
	if s.failAt > 0 && s.execCalls == s.failAt {
		return nil, errors.New("forced transaction failure")
	}
	return scEventWriteTestResult(100 + s.execCalls), nil
}

type scEventWriteTestConn struct {
	sqlx.SqlConn
	session    *scEventWriteTestSession
	committed  bool
	rolledBack bool
}

func (c *scEventWriteTestConn) TransactCtx(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	err := fn(ctx, c.session)
	if err != nil {
		c.rolledBack = true
		return err
	}
	c.committed = true
	return nil
}

type scEventWriteTestGListModel struct {
	GListModel
	invalidated int
}

func (m *scEventWriteTestGListModel) Invalidate(_ context.Context, rows ...*GList) error {
	m.invalidated += len(rows)
	return nil
}

type scEventWriteTestGScModel struct {
	GScModel
	invalidated int
}

func (m *scEventWriteTestGScModel) Invalidate(_ context.Context, _ ...*GSc) error {
	m.invalidated++
	return nil
}

func TestScEventWriteModelCreateEventCommitsMixedMovies(t *testing.T) {
	session := &scEventWriteTestSession{}
	conn := &scEventWriteTestConn{session: session}
	gLists := &scEventWriteTestGListModel{}
	events := &scEventWriteTestGScModel{}
	model := NewScEventWriteModel(conn, gLists, events)
	event := &GSc{Name: "2026-07-14-06-00", CreatedOn: 1000, UpdatedOn: 1000}
	movies := []*GList{
		{Name: event.Name + "__JAV", ScName: event.Name, MovieType: "jav", MovieJavId: "jav-id", IsSc: 2, IsCome: 1},
		{Name: event.Name + "__FC2", ScName: event.Name, MovieType: "fc2ppv", FC2MovieId: 42, IsSc: 2, IsCome: 2},
	}

	if err := model.CreateEvent(context.Background(), event, movies); err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}
	if !conn.committed || conn.rolledBack {
		t.Fatalf("transaction flags: committed=%v rolledBack=%v", conn.committed, conn.rolledBack)
	}
	if event.MovieNumber != 2 || event.Id == 0 || movies[0].Id == 0 || movies[1].Id == 0 {
		t.Fatalf("persisted ids/count: event=%#v movies=%#v", event, movies)
	}
	if gLists.invalidated != 2 || events.invalidated != 1 {
		t.Fatalf("cache invalidation: g_list=%d g_sc=%d", gLists.invalidated, events.invalidated)
	}
	if session.execCalls != 3 {
		t.Fatalf("exec calls = %d", session.execCalls)
	}
	for _, query := range session.queries {
		upper := strings.ToUpper(query)
		if strings.Contains(upper, "JOIN") || strings.Contains(upper, "EXISTS (") {
			t.Fatalf("transaction query contains related query: %s", query)
		}
	}
}

func TestScEventWriteModelCreateEventRollsBackOnMovieInsertFailure(t *testing.T) {
	session := &scEventWriteTestSession{failAt: 3}
	conn := &scEventWriteTestConn{session: session}
	gLists := &scEventWriteTestGListModel{}
	events := &scEventWriteTestGScModel{}
	model := NewScEventWriteModel(conn, gLists, events)
	event := &GSc{Name: "2026-07-14-06-15", CreatedOn: 1000, UpdatedOn: 1000}
	movies := []*GList{
		{Name: event.Name + "__JAV", ScName: event.Name, MovieType: "jav", MovieJavId: "jav-id", IsSc: 2, IsCome: 1},
		{Name: event.Name + "__FC2", ScName: event.Name, MovieType: "fc2ppv", FC2MovieId: 42, IsSc: 2, IsCome: 1},
	}

	err := model.CreateEvent(context.Background(), event, movies)
	if err == nil || !strings.Contains(err.Error(), "forced transaction failure") {
		t.Fatalf("CreateEvent error = %v", err)
	}
	if conn.committed || !conn.rolledBack {
		t.Fatalf("transaction flags: committed=%v rolledBack=%v", conn.committed, conn.rolledBack)
	}
	if gLists.invalidated != 0 || events.invalidated != 0 {
		t.Fatalf("cache invalidated after rollback: g_list=%d g_sc=%d", gLists.invalidated, events.invalidated)
	}
}
