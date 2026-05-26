package moviex

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/internal/consts"
)

type (
	CPersonScModel interface {
		Insert(ctx context.Context, data *CPersonSc) (sql.Result, error)
		Delete(ctx context.Context, id int64) error
		ListByPersonIDs(ctx context.Context, personIDs []int64) ([]*CPersonSc, error)
		ListRecentByPersonID(ctx context.Context, personID int64, limit int64) ([]*CPersonSc, error)
		DeleteByPersonIDs(ctx context.Context, personIDs []int64) error
		RebuildByPersonIDs(ctx context.Context, personIDs []int64, now int64) error
		RebuildByMovieJavIDs(ctx context.Context, movieJavIDs []string, now int64) error
		RebuildByScNames(ctx context.Context, scNames []string, now int64) error
	}

	customCPersonScModel struct {
		conn  sqlx.SqlConn
		table string
	}

	CPersonSc struct {
		Id          int64  `db:"id"`
		PersonId    int64  `db:"person_id"`
		ScName      string `db:"sc_name"`
		ScTime      int64  `db:"sc_time"`
		Cooldown    int64  `db:"cooldown"`
		MovieCount  int64  `db:"movie_count"`
		HasCome     int64  `db:"has_come"`
		MoviesJson  string `db:"movies_json"`
		CreatedTime int64  `db:"created_time"`
		UpdatedTime int64  `db:"updated_time"`
	}

	cPersonScSourceRow struct {
		PersonId   int64  `db:"person_id"`
		ScName     string `db:"sc_name"`
		ScTime     int64  `db:"sc_time"`
		Cooldown   int64  `db:"cooldown"`
		MovieJavId string `db:"movie_jav_id"`
		IsSc       int64  `db:"is_sc"`
		IsCome     int64  `db:"is_come"`
		MovieName  string `db:"movie_name"`
		GListName  string `db:"g_list_name"`
	}

	cPersonScMovieSnapshot struct {
		JavId  string `json:"javId"`
		Name   string `json:"name"`
		IsCome bool   `json:"isCome"`
	}
)

func NewCPersonScModel(conn sqlx.SqlConn, _ cache.CacheConf, _ ...cache.Option) CPersonScModel {
	return &customCPersonScModel{
		conn:  conn,
		table: "`c_person_sc`",
	}
}

func (m *customCPersonScModel) Insert(ctx context.Context, data *CPersonSc) (sql.Result, error) {
	query := "insert into " + m.table + " (`person_id`, `sc_name`, `sc_time`, `cooldown`, `movie_count`, `has_come`, `movies_json`, `created_time`, `updated_time`) values (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	return m.conn.ExecCtx(ctx, query, data.PersonId, data.ScName, data.ScTime, data.Cooldown, data.MovieCount, data.HasCome, data.MoviesJson, data.CreatedTime, data.UpdatedTime)
}

func (m *customCPersonScModel) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return nil
	}
	_, err := m.conn.ExecCtx(ctx, "delete from "+m.table+" where `id` = ?", id)
	return err
}

func (m *customCPersonScModel) ListByPersonIDs(ctx context.Context, personIDs []int64) ([]*CPersonSc, error) {
	uniq := uniqueInt64s(personIDs)
	if len(uniq) == 0 {
		return []*CPersonSc{}, nil
	}

	query, args := buildInQuery(
		"select `id`, `person_id`, `sc_name`, `sc_time`, `cooldown`, `movie_count`, `has_come`, `movies_json`, `created_time`, `updated_time` from "+m.table+" where `person_id` in (%s) order by `person_id` asc, `sc_time` desc, `sc_name` desc, `id` asc",
		int64SliceToAny(uniq),
	)
	var rows []*CPersonSc
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*CPersonSc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customCPersonScModel) ListRecentByPersonID(ctx context.Context, personID int64, limit int64) ([]*CPersonSc, error) {
	if personID <= 0 {
		return []*CPersonSc{}, nil
	}

	query := "select `id`, `person_id`, `sc_name`, `sc_time`, `cooldown`, `movie_count`, `has_come`, `movies_json`, `created_time`, `updated_time` from " + m.table + " where `person_id` = ? order by `sc_time` desc, `sc_name` desc, `id` asc"
	args := []any{personID}
	if limit > 0 {
		query += " limit ?"
		args = append(args, limit)
	}

	var rows []*CPersonSc
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*CPersonSc{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customCPersonScModel) DeleteByPersonIDs(ctx context.Context, personIDs []int64) error {
	uniq := uniqueInt64s(personIDs)
	if len(uniq) == 0 {
		return nil
	}

	rows, err := m.ListByPersonIDs(ctx, uniq)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, row := range rows {
			if row == nil || row.Id <= 0 {
				continue
			}
			if _, err := session.ExecCtx(ctx, "delete from "+m.table+" where `id` = ?", row.Id); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *customCPersonScModel) RebuildByMovieJavIDs(ctx context.Context, movieJavIDs []string, now int64) error {
	personIDs, err := m.listPersonIDsByMovieJavIDs(ctx, movieJavIDs)
	if err != nil {
		return err
	}
	return m.RebuildByPersonIDs(ctx, personIDs, now)
}

func (m *customCPersonScModel) RebuildByScNames(ctx context.Context, scNames []string, now int64) error {
	personIDs, err := m.listPersonIDsByScNames(ctx, scNames)
	if err != nil {
		return err
	}
	return m.RebuildByPersonIDs(ctx, personIDs, now)
}

func (m *customCPersonScModel) RebuildByPersonIDs(ctx context.Context, personIDs []int64, now int64) error {
	uniq := uniqueInt64s(personIDs)
	if len(uniq) == 0 {
		return nil
	}
	if now <= 0 {
		now = time.Now().Unix()
	}

	sourceRows, err := m.listSnapshotSourceRowsByPersonIDs(ctx, uniq)
	if err != nil {
		return err
	}
	grouped, err := buildCPersonScRows(sourceRows, now)
	if err != nil {
		return err
	}

	existingRows, err := m.ListByPersonIDs(ctx, uniq)
	if err != nil {
		return err
	}

	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, row := range existingRows {
			if row == nil || row.Id <= 0 {
				continue
			}
			if _, err := session.ExecCtx(ctx, "delete from "+m.table+" where `id` = ?", row.Id); err != nil {
				return err
			}
		}

		query := "insert into " + m.table + " (`person_id`, `sc_name`, `sc_time`, `cooldown`, `movie_count`, `has_come`, `movies_json`, `created_time`, `updated_time`) values (?, ?, ?, ?, ?, ?, ?, ?, ?)"
		for _, personID := range uniq {
			for _, row := range grouped[personID] {
				if row == nil {
					continue
				}
				if _, err := session.ExecCtx(ctx, query, row.PersonId, row.ScName, row.ScTime, row.Cooldown, row.MovieCount, row.HasCome, row.MoviesJson, row.CreatedTime, row.UpdatedTime); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (m *customCPersonScModel) listPersonIDsByMovieJavIDs(ctx context.Context, movieJavIDs []string) ([]int64, error) {
	uniq := uniqueNonEmptyStrings(movieJavIDs)
	if len(uniq) == 0 {
		return []int64{}, nil
	}

	query, args := buildInQuery(
		"select distinct ac.person_id from am_cast ac join amr_movie_cast amr on amr.cast_id = ac.id where ac.person_id > 0 and amr.movie_jav_id in (%s)",
		stringSliceToAny(uniq),
	)
	var ids []int64
	if err := m.conn.QueryRowsCtx(ctx, &ids, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []int64{}, nil
		}
		return nil, err
	}
	return uniqueInt64s(ids), nil
}

func (m *customCPersonScModel) listPersonIDsByScNames(ctx context.Context, scNames []string) ([]int64, error) {
	uniq := uniqueNonEmptyStrings(scNames)
	if len(uniq) == 0 {
		return []int64{}, nil
	}

	query, args := buildInQuery(
		"select distinct ac.person_id from am_cast ac join amr_movie_cast amr on amr.cast_id = ac.id join g_list gl on gl.movie_jav_id = amr.movie_jav_id where ac.person_id > 0 and gl.sc_name in (%s)",
		stringSliceToAny(uniq),
	)
	var ids []int64
	if err := m.conn.QueryRowsCtx(ctx, &ids, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []int64{}, nil
		}
		return nil, err
	}
	return uniqueInt64s(ids), nil
}

func (m *customCPersonScModel) listSnapshotSourceRowsByPersonIDs(ctx context.Context, personIDs []int64) ([]*cPersonScSourceRow, error) {
	query, args := buildInQuery(`
select
	ac.person_id as person_id,
	gl.sc_name as sc_name,
	coalesce(gs.sc_time, 0) as sc_time,
	coalesce(gs.cooldown, 0) as cooldown,
	gl.movie_jav_id as movie_jav_id,
	gl.is_sc as is_sc,
	gl.is_come as is_come,
	coalesce(am.name, '') as movie_name,
	coalesce(gl.name, '') as g_list_name
from am_cast ac
join amr_movie_cast amr on amr.cast_id = ac.id
join g_list gl on gl.movie_jav_id = amr.movie_jav_id
left join g_sc gs on gs.name = gl.sc_name
left join a_movie am on am.jav_id = gl.movie_jav_id
where ac.person_id in (%s)
	and ac.person_id > 0
	and gl.sc_name <> ''
	and gl.movie_jav_id <> ''
order by ac.person_id asc, gs.sc_time desc, gl.sc_name desc, gl.is_come desc, am.name asc, gl.movie_jav_id asc`,
		int64SliceToAny(personIDs),
	)

	var rows []*cPersonScSourceRow
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*cPersonScSourceRow{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func buildCPersonScRows(sourceRows []*cPersonScSourceRow, now int64) (map[int64][]*CPersonSc, error) {
	type eventAgg struct {
		personID int64
		scName   string
		scTime   int64
		cooldown int64
		hasCome  bool
		movies   map[string]*cPersonScMovieSnapshot
	}

	eventMap := make(map[string]*eventAgg)
	for _, row := range sourceRows {
		if row == nil || row.PersonId <= 0 {
			continue
		}
		if row.IsSc != consts.GListIsSc {
			continue
		}
		scName := strings.TrimSpace(row.ScName)
		movieJavID := strings.TrimSpace(row.MovieJavId)
		if scName == "" || movieJavID == "" {
			continue
		}

		key := buildPersonScEventKey(row.PersonId, scName)
		agg := eventMap[key]
		if agg == nil {
			agg = &eventAgg{
				personID: row.PersonId,
				scName:   scName,
				scTime:   row.ScTime,
				cooldown: row.Cooldown,
				movies:   make(map[string]*cPersonScMovieSnapshot),
			}
			eventMap[key] = agg
		}
		if row.ScTime > agg.scTime {
			agg.scTime = row.ScTime
		}

		name := strings.TrimSpace(row.MovieName)
		if name == "" {
			name = parseCPersonScMovieName(row.GListName)
		}
		if name == "" {
			name = movieJavID
		}

		movie := agg.movies[movieJavID]
		if movie == nil {
			movie = &cPersonScMovieSnapshot{
				JavId:  movieJavID,
				Name:   name,
				IsCome: row.IsCome == consts.GListIsCome,
			}
			agg.movies[movieJavID] = movie
		} else if row.IsCome == consts.GListIsCome {
			movie.IsCome = true
		}

		if row.IsCome == consts.GListIsCome {
			agg.hasCome = true
		}
	}

	grouped := make(map[int64][]*CPersonSc)
	for _, agg := range eventMap {
		if agg == nil || agg.personID <= 0 {
			continue
		}

		movies := make([]cPersonScMovieSnapshot, 0, len(agg.movies))
		for _, movie := range agg.movies {
			if movie == nil {
				continue
			}
			movies = append(movies, *movie)
		}
		sort.Slice(movies, func(i, j int) bool {
			if movies[i].IsCome != movies[j].IsCome {
				return movies[i].IsCome
			}
			if movies[i].Name != movies[j].Name {
				return movies[i].Name < movies[j].Name
			}
			return movies[i].JavId < movies[j].JavId
		})

		moviesJSON, err := json.Marshal(movies)
		if err != nil {
			return nil, err
		}

		grouped[agg.personID] = append(grouped[agg.personID], &CPersonSc{
			PersonId:    agg.personID,
			ScName:      agg.scName,
			ScTime:      agg.scTime,
			Cooldown:    agg.cooldown,
			MovieCount:  int64(len(movies)),
			HasCome:     boolToInt64(agg.hasCome),
			MoviesJson:  string(moviesJSON),
			CreatedTime: now,
			UpdatedTime: now,
		})
	}

	for personID := range grouped {
		sort.Slice(grouped[personID], func(i, j int) bool {
			if grouped[personID][i].ScTime != grouped[personID][j].ScTime {
				return grouped[personID][i].ScTime > grouped[personID][j].ScTime
			}
			return grouped[personID][i].ScName > grouped[personID][j].ScName
		})
	}
	return grouped, nil
}

func parseCPersonScMovieName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parts := strings.SplitN(raw, "__", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func buildPersonScEventKey(personID int64, scName string) string {
	return strconv.FormatInt(personID, 10) + "\x00" + strings.TrimSpace(scName)
}

func buildInQuery(format string, args []any) (string, []any) {
	placeholders := strings.TrimRight(strings.Repeat("?,", len(args)), ",")
	return strings.Replace(format, "%s", placeholders, 1), args
}

func int64SliceToAny(values []int64) []any {
	out := make([]any, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}

func stringSliceToAny(values []string) []any {
	out := make([]any, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}

func boolToInt64(v bool) int64 {
	if v {
		return 1
	}
	return 0
}
