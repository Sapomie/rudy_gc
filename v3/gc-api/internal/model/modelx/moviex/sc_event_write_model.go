package moviex

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ScEventWriteModel interface {
	CreateEvent(ctx context.Context, event *GSc, movies []*GList) error
	UpdateMovieStatus(ctx context.Context, scName string, gListID, isSc, now int64) (*GList, *GSc, error)
	UpdateEventMeta(ctx context.Context, event *GSc, comeGListID, now int64) ([]*GList, *GSc, error)
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
		result, err := session.ExecCtx(ctx, "INSERT INTO `g_sc` (`name`,`movie_number`,`sc_time`,`kind`,`come_movie_name`,`cooldown`,`duration`,`fg`,`vessel`,`movie_cast`,`remarks`,`created_on`,`updated_on`,`image_path`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			event.Name, event.MovieNumber, event.ScTime, event.Kind, event.ComeMovieName, event.Cooldown, event.Duration, event.Fg, event.Vessel, event.MovieCast, event.Remarks, event.CreatedOn, event.UpdatedOn, event.ImagePath)
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
		return nil
	})
	if err != nil {
		return err
	}
	if err := m.events.Invalidate(ctx, event); err != nil {
		return err
	}
	return m.gLists.Invalidate(ctx, movies...)
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
			event.MovieCast = ""
		}
		event.MovieNumber = countRow.Total
		event.UpdatedOn = now
		_, err := session.ExecCtx(ctx, "UPDATE `g_sc` SET `movie_number` = ?, `come_movie_name` = ?, `movie_cast` = ?, `updated_on` = ? WHERE `id` = ?", event.MovieNumber, event.ComeMovieName, event.MovieCast, event.UpdatedOn, event.Id)
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

func (m *defaultScEventWriteModel) UpdateEventMeta(ctx context.Context, event *GSc, comeGListID, now int64) ([]*GList, *GSc, error) {
	if event == nil || strings.TrimSpace(event.Name) == "" {
		return nil, nil, ErrNotFound
	}
	var lockedEvent GSc
	var rows []*GList
	err := m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := session.QueryRowCtx(ctx, &lockedEvent, "SELECT "+gScRows+" FROM `g_sc` WHERE `name` = ? LIMIT 1 FOR UPDATE", event.Name); err != nil {
			return err
		}
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
		lockedEvent.Kind = event.Kind
		lockedEvent.Duration = event.Duration
		lockedEvent.Fg = event.Fg
		lockedEvent.Vessel = event.Vessel
		lockedEvent.MovieCast = event.MovieCast
		lockedEvent.ComeMovieName = event.ComeMovieName
		lockedEvent.Remarks = event.Remarks
		lockedEvent.UpdatedOn = now
		_, err := session.ExecCtx(ctx, "UPDATE `g_sc` SET `kind` = ?, `duration` = ?, `fg` = ?, `vessel` = ?, `movie_cast` = ?, `come_movie_name` = ?, `remarks` = ?, `updated_on` = ? WHERE `id` = ?",
			lockedEvent.Kind, lockedEvent.Duration, lockedEvent.Fg, lockedEvent.Vessel, lockedEvent.MovieCast, lockedEvent.ComeMovieName, lockedEvent.Remarks, lockedEvent.UpdatedOn, lockedEvent.Id)
		return err
	})
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}
	if err := m.gLists.Invalidate(ctx, rows...); err != nil {
		return nil, nil, err
	}
	if err := m.events.Invalidate(ctx, &lockedEvent); err != nil {
		return nil, nil, err
	}
	return rows, &lockedEvent, nil
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
