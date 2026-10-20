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
	args      [][]any
	event     *GSc
	previous  *GSc
	next      *GSc
	movies    []*GList
	personIDs []int64
}

func (s *scEventWriteTestSession) ExecCtx(_ context.Context, query string, args ...any) (sql.Result, error) {
	s.execCalls++
	s.queries = append(s.queries, query)
	s.args = append(s.args, args)
	if s.failAt > 0 && s.execCalls == s.failAt {
		return nil, errors.New("forced transaction failure")
	}
	return scEventWriteTestResult(100 + s.execCalls), nil
}

func (s *scEventWriteTestSession) QueryRowCtx(_ context.Context, value any, query string, _ ...any) error {
	upper := strings.ToUpper(query)
	var source *GSc
	switch {
	case strings.Contains(upper, "WHERE `NAME` = ?"):
		source = s.event
	case strings.Contains(upper, "ORDER BY `SC_TIME` DESC"):
		source = s.previous
	case strings.Contains(upper, "ORDER BY `SC_TIME` ASC"):
		source = s.next
	}
	if source == nil {
		return sqlx.ErrNotFound
	}
	target, ok := value.(*GSc)
	if !ok {
		return errors.New("unexpected query row target")
	}
	*target = *source
	return nil
}

func (s *scEventWriteTestSession) QueryRowsCtx(_ context.Context, value any, query string, _ ...any) error {
	upper := strings.ToUpper(query)
	switch target := value.(type) {
	case *[]*GList:
		if !strings.Contains(upper, "FROM `G_LIST`") {
			return errors.New("unexpected g_list query")
		}
		*target = append([]*GList(nil), s.movies...)
	case *[]int64:
		if !strings.Contains(upper, "FROM `C_PERSON_SC`") {
			return errors.New("unexpected c_person_sc query")
		}
		*target = append([]int64(nil), s.personIDs...)
	default:
		return errors.New("unexpected query rows target")
	}
	return nil
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

func (m *scEventWriteTestGScModel) Invalidate(_ context.Context, rows ...*GSc) error {
	m.invalidated += len(rows)
	return nil
}

func TestScEventWriteModelCreateEventCommitsMixedMovies(t *testing.T) {
	session := &scEventWriteTestSession{}
	conn := &scEventWriteTestConn{session: session}
	gLists := &scEventWriteTestGListModel{}
	events := &scEventWriteTestGScModel{}
	model := NewScEventWriteModel(conn, gLists, events)
	event := &GSc{
		Name: "2026-07-14-06-00", CreatedOn: 1000, UpdatedOn: 1000,
		KindOptionId:      sql.NullInt64{Int64: 1, Valid: true},
		FgOptionId:        sql.NullInt64{Int64: 2, Valid: true},
		VesselOptionId:    sql.NullInt64{Int64: 3, Valid: true},
		MovieCastPersonId: sql.NullInt64{Int64: 91, Valid: true},
	}
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
	if session.execCalls != 6 {
		t.Fatalf("exec calls = %d", session.execCalls)
	}
	if strings.Contains(session.queries[0], "`kind`,") || strings.Contains(session.queries[0], "`fg`,") || strings.Contains(session.queries[0], "`vessel`,") {
		t.Fatalf("event insert still writes legacy option columns: %s", session.queries[0])
	}
	if len(session.args[0]) != 14 {
		t.Fatalf("event insert args = %d, want 14", len(session.args[0]))
	}
	if !strings.Contains(session.queries[0], "`movie_cast_person_id`") || strings.Contains(session.queries[0], "`movie_cast`") {
		t.Fatalf("event insert actor columns = %s", session.queries[0])
	}
	if got, ok := session.args[0][13].(sql.NullInt64); !ok || !got.Valid || got.Int64 != 91 {
		t.Fatalf("event insert actor person_id = %#v", session.args[0][13])
	}
	for index, table := range []string{"`g_sc_option_kind`", "`g_sc_option_fg`", "`g_sc_option_vessl`"} {
		queryIndex := index + 3
		if !strings.Contains(session.queries[queryIndex], "UPDATE "+table) || !strings.Contains(session.queries[queryIndex], "`sc_event_times`") || !strings.Contains(session.queries[queryIndex], "`last_sc_event_time`") {
			t.Fatalf("aggregate query %d = %q", queryIndex, session.queries[queryIndex])
		}
	}
	for _, query := range session.queries {
		upper := strings.ToUpper(query)
		if strings.Contains(upper, "JOIN") || strings.Contains(upper, "EXISTS (") {
			t.Fatalf("transaction query contains related query: %s", query)
		}
	}
}

func TestScEventWriteModelUpdateEventMetaOnlyWritesOptionIDs(t *testing.T) {
	session := &scEventWriteTestSession{
		event: &GSc{
			Id: 20, Name: "2026-07-28-12-00",
			ImagePath:    "current-image.png",
			KindOptionId: sql.NullInt64{Int64: 9, Valid: true},
			FgOptionId:   sql.NullInt64{Int64: 2, Valid: true},
		},
		movies: []*GList{},
	}
	conn := &scEventWriteTestConn{session: session}
	model := NewScEventWriteModel(conn, &scEventWriteTestGListModel{}, &scEventWriteTestGScModel{})
	event := &GSc{
		Name:              session.event.Name,
		ImagePath:         "stale-image.png",
		KindOptionId:      sql.NullInt64{Int64: 1, Valid: true},
		FgOptionId:        sql.NullInt64{Int64: 2, Valid: true},
		VesselOptionId:    sql.NullInt64{Int64: 3, Valid: true},
		Duration:          12,
		MovieCastPersonId: sql.NullInt64{Int64: 92, Valid: true},
	}

	_, updated, err := model.UpdateEventMeta(context.Background(), event, 0, 999, nil)
	if err != nil {
		t.Fatalf("UpdateEventMeta: %v", err)
	}
	if updated == nil || updated.KindOptionId.Int64 != 1 || updated.FgOptionId.Int64 != 2 || updated.VesselOptionId.Int64 != 3 {
		t.Fatalf("updated event = %#v", updated)
	}
	if updated.ImagePath != "current-image.png" {
		t.Fatalf("metadata update replaced the current image: %q", updated.ImagePath)
	}
	if len(session.queries) != 5 || len(session.args) != 5 || len(session.args[0]) != 9 {
		t.Fatalf("queries = %#v, args = %#v", session.queries, session.args)
	}
	query := session.queries[0]
	if !strings.Contains(query, "`kind_option_id`") || !strings.Contains(query, "`fg_option_id`") || !strings.Contains(query, "`vessel_option_id`") {
		t.Fatalf("option id columns missing from update: %s", query)
	}
	if strings.Contains(query, "`kind` =") || strings.Contains(query, "`fg` =") || strings.Contains(query, "`vessel` =") {
		t.Fatalf("event update still writes legacy option columns: %s", query)
	}
	if !strings.Contains(query, "`movie_cast_person_id` =") || strings.Contains(query, "`movie_cast` =") {
		t.Fatalf("event update actor columns = %s", query)
	}
	if got, ok := session.args[0][4].(sql.NullInt64); !ok || !got.Valid || got.Int64 != 92 {
		t.Fatalf("event update actor person_id = %#v", session.args[0][4])
	}
	wantAggregateTargets := []struct {
		table string
		id    int64
	}{
		{table: "`g_sc_option_kind`", id: 9},
		{table: "`g_sc_option_fg`", id: 2},
		{table: "`g_sc_option_kind`", id: 1},
		{table: "`g_sc_option_vessl`", id: 3},
	}
	for index, target := range wantAggregateTargets {
		queryIndex := index + 1
		if !strings.Contains(session.queries[queryIndex], "UPDATE "+target.table) {
			t.Fatalf("aggregate query %d = %q", queryIndex, session.queries[queryIndex])
		}
		if len(session.args[queryIndex]) != 3 || session.args[queryIndex][2] != target.id {
			t.Fatalf("aggregate args %d = %#v", queryIndex, session.args[queryIndex])
		}
	}
}

func TestScEventWriteModelUpdateEventMetaImage(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "rollback"}[fail], func(t *testing.T) {
			session := &scEventWriteTestSession{event: &GSc{Id: 42, Name: "event-a", ImagePath: "original.png"}}
			if fail {
				session.failAt = 1
			}
			conn := &scEventWriteTestConn{session: session}
			model := NewScEventWriteModel(conn, &scEventWriteTestGListModel{}, &scEventWriteTestGScModel{})
			path := "new.png"
			_, updated, err := model.UpdateEventMeta(context.Background(), &GSc{Name: "event-a", Remarks: "new remarks"}, 0, 999, &path)
			if fail {
				if err == nil || updated != nil || !conn.rolledBack || conn.committed {
					t.Fatalf("failure result = %#v, %v", updated, err)
				}
				return
			}
			if err != nil || !conn.committed || updated.ImagePath != path || updated.Remarks != "new remarks" {
				t.Fatalf("saved event = %#v, err = %v", updated, err)
			}
			if !strings.Contains(session.queries[0], "`image_path` = ?") || len(session.args[0]) != 10 || session.args[0][8] != path {
				t.Fatalf("image update SQL = %s, args = %#v", session.queries[0], session.args[0])
			}
		})
	}
}

type scEventImageCacheFailure struct{ GScModel }

func (*scEventImageCacheFailure) Invalidate(context.Context, ...*GSc) error {
	return errors.New("cache failed")
}

func TestScEventWriteModelImageCacheFailureReturnsCommittedEvent(t *testing.T) {
	session := &scEventWriteTestSession{event: &GSc{Id: 42, Name: "event-a"}}
	conn := &scEventWriteTestConn{session: session}
	model := NewScEventWriteModel(conn, &scEventWriteTestGListModel{}, &scEventImageCacheFailure{})
	path := "new.png"
	_, updated, err := model.UpdateEventMeta(context.Background(), &GSc{Name: "event-a"}, 0, 999, &path)
	if err == nil || !conn.committed || updated == nil || updated.ImagePath != path {
		t.Fatalf("committed image result = %#v, err = %v", updated, err)
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

func TestScEventWriteModelDeleteEventCommitsAndRebuildsNextCooldown(t *testing.T) {
	session := &scEventWriteTestSession{
		event: &GSc{
			Id: 20, Name: "2026-07-28-12-00", ScTime: 200, Cooldown: 100,
			KindOptionId:   sql.NullInt64{Int64: 1, Valid: true},
			FgOptionId:     sql.NullInt64{Int64: 2, Valid: true},
			VesselOptionId: sql.NullInt64{Int64: 3, Valid: true},
		},
		previous:  &GSc{Id: 10, Name: "2026-07-28-11-58", ScTime: 100},
		next:      &GSc{Id: 30, Name: "2026-07-28-12-03", ScTime: 300, Cooldown: 100},
		movies:    []*GList{{Id: 1, Name: "jav-row"}, {Id: 2, Name: "fc2-row"}},
		personIDs: []int64{7, 7, 0, 9},
	}
	conn := &scEventWriteTestConn{session: session}
	gLists := &scEventWriteTestGListModel{}
	events := &scEventWriteTestGScModel{}
	model := NewScEventWriteModel(conn, gLists, events)

	result, err := model.DeleteEvent(context.Background(), session.event.Name, 999)
	if err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}
	if !conn.committed || conn.rolledBack {
		t.Fatalf("transaction flags: committed=%v rolledBack=%v", conn.committed, conn.rolledBack)
	}
	if result == nil || result.Event == nil || result.Event.Id != 20 {
		t.Fatalf("deleted event = %#v", result)
	}
	if result.NextEvent == nil || result.NextEvent.Id != 30 || result.NextEvent.Cooldown != 200 || result.NextEvent.UpdatedOn != 999 {
		t.Fatalf("next event = %#v", result.NextEvent)
	}
	if len(result.Movies) != 2 {
		t.Fatalf("deleted movies = %d", len(result.Movies))
	}
	if len(result.PersonIDs) != 2 || result.PersonIDs[0] != 7 || result.PersonIDs[1] != 9 {
		t.Fatalf("person ids = %#v", result.PersonIDs)
	}
	if session.execCalls != 7 {
		t.Fatalf("exec calls = %d, queries = %#v", session.execCalls, session.queries)
	}
	if gLists.invalidated != 2 || events.invalidated != 2 {
		t.Fatalf("cache invalidation: g_list=%d g_sc=%d", gLists.invalidated, events.invalidated)
	}
	wantFragments := []string{
		"UPDATE `G_SC` SET `COOLDOWN`",
		"DELETE FROM `C_PERSON_SC`",
		"DELETE FROM `G_LIST`",
		"DELETE FROM `G_SC`",
	}
	for index, fragment := range wantFragments {
		if !strings.Contains(strings.ToUpper(session.queries[index]), fragment) {
			t.Fatalf("query %d = %q, want fragment %q", index, session.queries[index], fragment)
		}
	}
	for index, table := range []string{"`G_SC_OPTION_KIND`", "`G_SC_OPTION_FG`", "`G_SC_OPTION_VESSL`"} {
		queryIndex := index + len(wantFragments)
		if !strings.Contains(strings.ToUpper(session.queries[queryIndex]), "UPDATE "+table) {
			t.Fatalf("aggregate query %d = %q", queryIndex, session.queries[queryIndex])
		}
	}
}

func TestScEventWriteModelDeleteFirstEventResetsNextCooldown(t *testing.T) {
	session := &scEventWriteTestSession{
		event: &GSc{Id: 10, Name: "2026-07-28-11-58", ScTime: 100},
		next:  &GSc{Id: 20, Name: "2026-07-28-12-00", ScTime: 200, Cooldown: 100},
	}
	conn := &scEventWriteTestConn{session: session}
	model := NewScEventWriteModel(conn, &scEventWriteTestGListModel{}, &scEventWriteTestGScModel{})

	result, err := model.DeleteEvent(context.Background(), session.event.Name, 999)
	if err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}
	if result.NextEvent == nil || result.NextEvent.Cooldown != 0 || result.NextEvent.UpdatedOn != 999 {
		t.Fatalf("next event = %#v", result.NextEvent)
	}
	if session.execCalls != 4 {
		t.Fatalf("exec calls = %d, queries = %#v", session.execCalls, session.queries)
	}
}

func TestScEventWriteModelDeleteLastEventLeavesNoCooldownTarget(t *testing.T) {
	session := &scEventWriteTestSession{
		event:    &GSc{Id: 20, Name: "2026-07-28-12-00", ScTime: 200},
		previous: &GSc{Id: 10, Name: "2026-07-28-11-58", ScTime: 100},
	}
	conn := &scEventWriteTestConn{session: session}
	model := NewScEventWriteModel(conn, &scEventWriteTestGListModel{}, &scEventWriteTestGScModel{})

	result, err := model.DeleteEvent(context.Background(), session.event.Name, 999)
	if err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}
	if result.NextEvent != nil {
		t.Fatalf("next event = %#v", result.NextEvent)
	}
	if session.execCalls != 3 {
		t.Fatalf("exec calls = %d, queries = %#v", session.execCalls, session.queries)
	}
}

func TestScEventWriteModelDeleteEventRollsBackOnDeleteFailure(t *testing.T) {
	session := &scEventWriteTestSession{
		event:  &GSc{Id: 20, Name: "2026-07-28-12-00", ScTime: 200},
		movies: []*GList{{Id: 1, Name: "jav-row"}},
		failAt: 2,
	}
	conn := &scEventWriteTestConn{session: session}
	gLists := &scEventWriteTestGListModel{}
	events := &scEventWriteTestGScModel{}
	model := NewScEventWriteModel(conn, gLists, events)

	result, err := model.DeleteEvent(context.Background(), session.event.Name, 999)
	if err == nil || !strings.Contains(err.Error(), "forced transaction failure") {
		t.Fatalf("DeleteEvent error = %v", err)
	}
	if result != nil {
		t.Fatalf("DeleteEvent result = %#v", result)
	}
	if conn.committed || !conn.rolledBack {
		t.Fatalf("transaction flags: committed=%v rolledBack=%v", conn.committed, conn.rolledBack)
	}
	if gLists.invalidated != 0 || events.invalidated != 0 {
		t.Fatalf("cache invalidated after rollback: g_list=%d g_sc=%d", gLists.invalidated, events.invalidated)
	}
}

func TestScEventWriteModelDeleteEventReturnsNotFound(t *testing.T) {
	session := &scEventWriteTestSession{}
	conn := &scEventWriteTestConn{session: session}
	model := NewScEventWriteModel(conn, &scEventWriteTestGListModel{}, &scEventWriteTestGScModel{})

	result, err := model.DeleteEvent(context.Background(), "missing", 999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("DeleteEvent error = %v", err)
	}
	if result != nil {
		t.Fatalf("DeleteEvent result = %#v", result)
	}
}
