package moviex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ScEventWriteModel interface {
	CreateEvent(ctx context.Context, event *GSc, movies []*GList) error
	DeleteEvent(ctx context.Context, scName string, now int64) (*ScEventDeleteResult, error)
	UpdateMovieStatus(ctx context.Context, scName string, gListID, isSc, now int64) (*GList, *GSc, error)
	UpdateEventMeta(ctx context.Context, event *GSc, comeGListID, now int64, imagePath *string) ([]*GList, *GSc, error)
}

type ScEventDeleteResult struct {
	Event     *GSc
	Movies    []*GList
	NextEvent *GSc
	PersonIDs []int64
}

type defaultScEventWriteModel struct {
	conn   sqlx.SqlConn
	gLists GListModel
	events GScModel
}

func NewScEventWriteModel(conn sqlx.SqlConn, gLists GListModel, events GScModel) ScEventWriteModel {
	return &defaultScEventWriteModel{conn: conn, gLists: gLists, events: events}
}

func (m *defaultScEventWriteModel) CreateEvent(ctx context.Context, event *GSc, movies []*GList) error {
	if event == nil || strings.TrimSpace(event.Name) == "" {
		return errors.New("invalid sc event")
	}
	comeCount := 0
	scCount := int64(0)
	for _, movie := range movies {
		if err := validateScEventMovieIdentity(movie); err != nil {
			return err
		}
		if movie.CreatedOn <= 0 {
			movie.CreatedOn = event.CreatedOn
		}
		if movie.UpdatedOn <= 0 {
			movie.UpdatedOn = event.UpdatedOn
		}
		if movie.IsCome == 2 {
			comeCount++
			if movie.IsSc != 2 {
				return errors.New("come movie must be selected from is_sc=2 movies")
			}
		}
		if movie.IsSc == 2 {
			scCount++
		}
	}
	if comeCount > 1 {
		return errors.New("sc event contains multiple come movies")
	}
	event.MovieNumber = scCount

	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		result, err := session.ExecCtx(ctx, "INSERT INTO `g_sc` (`name`,`movie_number`,`sc_time`,`come_movie_name`,`cooldown`,`duration`,`remarks`,`created_on`,`updated_on`,`image_path`,`kind_option_id`,`fg_option_id`,`vessel_option_id`,`movie_cast_person_id`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			event.Name, event.MovieNumber, event.ScTime, event.ComeMovieName, event.Cooldown, event.Duration, event.Remarks, event.CreatedOn, event.UpdatedOn, event.ImagePath, event.KindOptionId, event.FgOptionId, event.VesselOptionId, event.MovieCastPersonId)
		if err != nil {
			return err
		}
		event.Id, err = result.LastInsertId()
		if err != nil {
			return err
		}
		for _, movie := range movies {
			result, err := session.ExecCtx(ctx, "INSERT INTO `g_list` (`name`,`sc_name`,`movie_type`,`movie_jav_id`,`fc2_movie_id`,`is_come`,`is_sc`,`created_on`,`updated_on`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
				movie.Name, movie.ScName, movie.MovieType, movie.MovieJavId, movie.FC2MovieId, movie.IsCome, movie.IsSc, movie.CreatedOn, movie.UpdatedOn)
			if err != nil {
				return err
			}
			movie.Id, err = result.LastInsertId()
			if err != nil {
				return err
			}
		}
		return refreshScOptionAggregates(ctx, session, event)
	})
	if err != nil {
		return err
	}
	if err := m.events.Invalidate(ctx, event); err != nil {
		return err
	}
	return m.gLists.Invalidate(ctx, movies...)
}

func (m *defaultScEventWriteModel) DeleteEvent(ctx context.Context, scName string, now int64) (*ScEventDeleteResult, error) {
	scName = strings.TrimSpace(scName)
	if scName == "" {
		return nil, ErrNotFound
	}

	result := &ScEventDeleteResult{}
	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		var event GSc
		if err := session.QueryRowCtx(ctx, &event, "SELECT "+gScRows+" FROM `g_sc` WHERE `name` = ? LIMIT 1 FOR UPDATE", scName); err != nil {
			return err
		}
		result.Event = &event

		var movies []*GList
		if err := session.QueryRowsCtx(ctx, &movies, "SELECT "+gListRows+" FROM `g_list` WHERE `sc_name` = ? ORDER BY `id` ASC FOR UPDATE", scName); err != nil && !errors.Is(err, sqlx.ErrNotFound) {
			return err
		}
		result.Movies = movies

		var personIDs []int64
		if err := session.QueryRowsCtx(ctx, &personIDs, "SELECT DISTINCT `person_id` FROM `c_person_sc` WHERE `sc_name` = ? AND `person_id` > 0", scName); err != nil && !errors.Is(err, sqlx.ErrNotFound) {
			return err
		}
		result.PersonIDs = uniquePositiveInt64s(personIDs)

		var previousEvent GSc
		previousErr := session.QueryRowCtx(
			ctx,
			&previousEvent,
			"SELECT "+gScRows+" FROM `g_sc` WHERE (`sc_time` < ? OR (`sc_time` = ? AND `id` < ?)) ORDER BY `sc_time` DESC, `id` DESC LIMIT 1 FOR UPDATE",
			event.ScTime,
			event.ScTime,
			event.Id,
		)
		if previousErr != nil && !errors.Is(previousErr, sqlx.ErrNotFound) {
			return previousErr
		}

		var nextEvent GSc
		nextErr := session.QueryRowCtx(
			ctx,
			&nextEvent,
			"SELECT "+gScRows+" FROM `g_sc` WHERE (`sc_time` > ? OR (`sc_time` = ? AND `id` > ?)) ORDER BY `sc_time` ASC, `id` ASC LIMIT 1 FOR UPDATE",
			event.ScTime,
			event.ScTime,
			event.Id,
		)
		if nextErr != nil && !errors.Is(nextErr, sqlx.ErrNotFound) {
			return nextErr
		}
		if nextErr == nil {
			nextCooldown := int64(0)
			if previousErr == nil {
				nextCooldown = nextEvent.ScTime - previousEvent.ScTime
				if nextCooldown < 0 {
					nextCooldown = 0
				}
			}
			if nextEvent.Cooldown != nextCooldown {
				nextEvent.Cooldown = nextCooldown
				nextEvent.UpdatedOn = now
				if _, err := session.ExecCtx(ctx, "UPDATE `g_sc` SET `cooldown` = ?, `updated_on` = ? WHERE `id` = ?", nextEvent.Cooldown, nextEvent.UpdatedOn, nextEvent.Id); err != nil {
					return err
				}
			}
			result.NextEvent = &nextEvent
		}

		if _, err := session.ExecCtx(ctx, "DELETE FROM `c_person_sc` WHERE `sc_name` = ?", scName); err != nil {
			return err
		}
		if _, err := session.ExecCtx(ctx, "DELETE FROM `g_list` WHERE `sc_name` = ?", scName); err != nil {
			return err
		}
		if _, err := session.ExecCtx(ctx, "DELETE FROM `g_sc` WHERE `id` = ?", event.Id); err != nil {
			return err
		}
		return refreshScOptionAggregates(ctx, session, &event)
	})
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	cacheErrors := make([]error, 0, 2)
	if err := m.gLists.Invalidate(ctx, result.Movies...); err != nil {
		cacheErrors = append(cacheErrors, fmt.Errorf("invalidate g_list cache: %w", err))
	}
	events := []*GSc{result.Event}
	if result.NextEvent != nil {
		events = append(events, result.NextEvent)
	}
	if err := m.events.Invalidate(ctx, events...); err != nil {
		cacheErrors = append(cacheErrors, fmt.Errorf("invalidate g_sc cache: %w", err))
	}
	if len(cacheErrors) > 0 {
		return result, errors.Join(cacheErrors...)
	}
	return result, nil
}

func (m *defaultScEventWriteModel) UpdateMovieStatus(ctx context.Context, scName string, gListID, isSc, now int64) (*GList, *GSc, error) {
	scName = strings.TrimSpace(scName)
	if scName == "" || gListID <= 0 {
		return nil, nil, ErrNotFound
	}
	if isSc != 1 && isSc != 2 {
		return nil, nil, errors.New("invalid sc movie status")
	}
	var target GList
	var event GSc
	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := session.QueryRowCtx(ctx, &event, "SELECT "+gScRows+" FROM `g_sc` WHERE `name` = ? LIMIT 1 FOR UPDATE", scName); err != nil {
			return err
		}
		if err := session.QueryRowCtx(ctx, &target, "SELECT "+gListRows+" FROM `g_list` WHERE `id` = ? AND `sc_name` = ? LIMIT 1 FOR UPDATE", gListID, scName); err != nil {
			return err
		}
		wasCome := target.IsCome == 2
		isCome := target.IsCome
		if isSc == 1 {
			isCome = 1
		}
		if _, err := session.ExecCtx(ctx, "UPDATE `g_list` SET `is_sc` = ?, `is_come` = ?, `updated_on` = ? WHERE `id` = ?", isSc, isCome, now, target.Id); err != nil {
			return err
		}
		target.IsSc = isSc
		target.IsCome = isCome
		target.UpdatedOn = now

		var countRow struct {
			Total int64 `db:"total"`
		}
		if err := session.QueryRowCtx(ctx, &countRow, "SELECT COUNT(*) AS total FROM `g_list` WHERE `sc_name` = ? AND `is_sc` = ?", scName, int64(2)); err != nil {
			return err
		}
		if wasCome && isSc == 1 {
			event.ComeMovieName = ""
			event.MovieCastPersonId = sql.NullInt64{}
		}
		event.MovieNumber = countRow.Total
		event.UpdatedOn = now
		_, err := session.ExecCtx(ctx, "UPDATE `g_sc` SET `movie_number` = ?, `come_movie_name` = ?, `movie_cast_person_id` = ?, `updated_on` = ? WHERE `id` = ?", event.MovieNumber, event.ComeMovieName, event.MovieCastPersonId, event.UpdatedOn, event.Id)
		return err
	})
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}
	if err := m.gLists.Invalidate(ctx, &target); err != nil {
		return nil, nil, err
	}
	if err := m.events.Invalidate(ctx, &event); err != nil {
		return nil, nil, err
	}
	return &target, &event, nil
}

func (m *defaultScEventWriteModel) UpdateEventMeta(ctx context.Context, event *GSc, comeGListID, now int64, imagePath *string) ([]*GList, *GSc, error) {
	if event == nil || strings.TrimSpace(event.Name) == "" {
		return nil, nil, ErrNotFound
	}
	var lockedEvent GSc
	var rows []*GList
	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := session.QueryRowCtx(ctx, &lockedEvent, "SELECT "+gScRows+" FROM `g_sc` WHERE `name` = ? LIMIT 1 FOR UPDATE", event.Name); err != nil {
			return err
		}
		previousEvent := lockedEvent
		if err := session.QueryRowsCtx(ctx, &rows, "SELECT "+gListRows+" FROM `g_list` WHERE `sc_name` = ? ORDER BY `id` ASC FOR UPDATE", event.Name); err != nil && !errors.Is(err, sqlx.ErrNotFound) {
			return err
		}
		if comeGListID > 0 {
			found := false
			for _, row := range rows {
				if row != nil && row.Id == comeGListID && row.IsSc == 2 {
					found = true
					break
				}
			}
			if !found {
				return errors.New("come movie must be selected from is_sc=2 movies")
			}
		}
		for _, row := range rows {
			if row == nil {
				continue
			}
			want := int64(1)
			if comeGListID > 0 && row.Id == comeGListID {
				want = 2
			}
			if row.IsCome == want {
				continue
			}
			row.IsCome = want
			row.UpdatedOn = now
			if _, err := session.ExecCtx(ctx, "UPDATE `g_list` SET `is_come` = ?, `updated_on` = ? WHERE `id` = ?", row.IsCome, row.UpdatedOn, row.Id); err != nil {
				return err
			}
		}
		lockedEvent.KindOptionId = event.KindOptionId
		lockedEvent.Duration = event.Duration
		lockedEvent.FgOptionId = event.FgOptionId
		lockedEvent.VesselOptionId = event.VesselOptionId
		lockedEvent.MovieCastPersonId = event.MovieCastPersonId
		lockedEvent.ComeMovieName = event.ComeMovieName
		lockedEvent.Remarks = event.Remarks
		lockedEvent.UpdatedOn = now
		query := "UPDATE `g_sc` SET `kind_option_id` = ?, `duration` = ?, `fg_option_id` = ?, `vessel_option_id` = ?, `movie_cast_person_id` = ?, `come_movie_name` = ?, `remarks` = ?, `updated_on` = ?"
		args := []any{lockedEvent.KindOptionId, lockedEvent.Duration, lockedEvent.FgOptionId, lockedEvent.VesselOptionId, lockedEvent.MovieCastPersonId, lockedEvent.ComeMovieName, lockedEvent.Remarks, lockedEvent.UpdatedOn}
		if imagePath != nil {
			lockedEvent.ImagePath = *imagePath
			query += ", `image_path` = ?"
			args = append(args, lockedEvent.ImagePath)
		}
		query += " WHERE `id` = ?"
		args = append(args, lockedEvent.Id)
		if _, err := session.ExecCtx(ctx, query, args...); err != nil {
			return err
		}
		return refreshScOptionAggregates(ctx, session, &previousEvent, &lockedEvent)
	})
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}
	if err := errors.Join(m.gLists.Invalidate(ctx, rows...), m.events.Invalidate(ctx, &lockedEvent)); err != nil {
		// 非空事件表示事务已提交，调用方必须保留已被数据库引用的新图片。
		return rows, &lockedEvent, err
	}
	return rows, &lockedEvent, nil
}

type scOptionAggregateTarget struct {
	table    string
	idColumn string
	id       int64
}

func refreshScOptionAggregates(ctx context.Context, session sqlx.Session, events ...*GSc) error {
	targets := make([]scOptionAggregateTarget, 0, len(events)*3)
	seen := make(map[scOptionAggregateTarget]struct{}, len(events)*3)
	appendTarget := func(table, idColumn string, id int64, valid bool) {
		if !valid || id <= 0 {
			return
		}
		target := scOptionAggregateTarget{table: table, idColumn: idColumn, id: id}
		if _, ok := seen[target]; ok {
			return
		}
		seen[target] = struct{}{}
		targets = append(targets, target)
	}

	for _, event := range events {
		if event == nil {
			continue
		}
		appendTarget(gScKindOptionTable, "kind_option_id", event.KindOptionId.Int64, event.KindOptionId.Valid)
		appendTarget(gScFGOptionTable, "fg_option_id", event.FgOptionId.Int64, event.FgOptionId.Valid)
		appendTarget(gScVesselOptionTable, "vessel_option_id", event.VesselOptionId.Int64, event.VesselOptionId.Valid)
	}

	for _, target := range targets {
		query := "UPDATE " + target.table +
			" SET `sc_event_times` = (SELECT COUNT(*) FROM `g_sc` WHERE `" + target.idColumn + "` = ?)," +
			" `last_sc_event_time` = COALESCE((SELECT MAX(`sc_time`) FROM `g_sc` WHERE `" + target.idColumn + "` = ?), 0)" +
			" WHERE `id` = ?"
		if _, err := session.ExecCtx(ctx, query, target.id, target.id, target.id); err != nil {
			return fmt.Errorf("refresh %s aggregate(%d): %w", target.table, target.id, err)
		}
	}
	return nil
}

func validateScEventMovieIdentity(movie *GList) error {
	if movie == nil || strings.TrimSpace(movie.Name) == "" || strings.TrimSpace(movie.ScName) == "" {
		return errors.New("invalid sc event movie")
	}
	switch strings.TrimSpace(movie.MovieType) {
	case "jav":
		if strings.TrimSpace(movie.MovieJavId) == "" || movie.FC2MovieId != 0 {
			return fmt.Errorf("invalid jav sc movie identity: %s", movie.Name)
		}
	case "fc2ppv":
		if strings.TrimSpace(movie.MovieJavId) != "" || movie.FC2MovieId <= 0 {
			return fmt.Errorf("invalid fc2 sc movie identity: %s", movie.Name)
		}
	default:
		return fmt.Errorf("invalid sc movie type: %s", movie.MovieType)
	}
	return nil
}
