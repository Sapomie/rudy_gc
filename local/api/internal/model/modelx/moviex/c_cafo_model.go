package moviex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CCafoModel = (*customCCafoModel)(nil)

type (
	// CCafoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCCafoModel.
	CCafoModel interface {
		cCafoModel
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		ListAll(ctx context.Context) ([]*CCafo, error)
		ListByPersonID(ctx context.Context, personID int64) ([]*CCafoProfile, error)
		UpdateProfileForPerson(ctx context.Context, personID, cafoID int64, data *CCafoProfileUpdate) (*CCafoProfile, *CPerson, error)
		ListPersonLinks(ctx context.Context) (map[int64]int64, error)
		SetPersonID(ctx context.Context, cafoID, personID int64) error
		ReassignPersonIDs(ctx context.Context, keepPersonID int64, sourcePersonIDs []int64) error
	}

	customCCafoModel struct {
		*defaultCCafoModel
	}

	CCafoProfile struct {
		Id        int64  `db:"id"`
		PersonId  int64  `db:"person_id"`
		Name      string `db:"name"`
		Alias     string `db:"alias"`
		Chinese   string `db:"chinese"`
		BirthDay  int64  `db:"birth_day"`
		Height    int64  `db:"height"`
		Cup       string `db:"cup"`
		Bwh       string `db:"bwh"`
		Avatar    string `db:"avtar"`
		Biography string `db:"biography"`
		CreatedOn int64  `db:"created_on"`
		UpdatedOn int64  `db:"updated_on"`
	}

	CCafoProfileUpdate struct {
		Name      string
		Alias     string
		Chinese   string
		BirthDay  int64
		Height    int64
		Cup       string
		Bwh       string
		Avatar    string
		Biography string
		UpdatedOn int64
	}
)

const cCafoProfileRows = "`id`,`person_id`,`name`,`alias`,`chinese`,`birth_day`,`height`,`cup`,`bwh`,`avtar`,`biography`,`created_on`,`updated_on`"

// NewCCafoModel returns a model for the database table.
func NewCCafoModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CCafoModel {
	return &customCCafoModel{
		defaultCCafoModel: newCCafoModel(conn, c, opts...),
	}
}

func (m *customCCafoModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customCCafoModel) ListAll(ctx context.Context) ([]*CCafo, error) {
	var rows []*CCafo
	query := "select " + cCafoRows + " from " + m.table + " order by id asc"
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*CCafo{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customCCafoModel) ListByPersonID(ctx context.Context, personID int64) ([]*CCafoProfile, error) {
	if personID <= 0 {
		return []*CCafoProfile{}, nil
	}
	var rows []*CCafoProfile
	query := "select " + cCafoProfileRows + " from " + m.table + " where `person_id` = ? order by `id` asc"
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, personID); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*CCafoProfile{}, nil
		}
		return nil, err
	}
	if rows == nil {
		return []*CCafoProfile{}, nil
	}
	return rows, nil
}

func (m *customCCafoModel) UpdateProfileForPerson(
	ctx context.Context,
	personID, cafoID int64,
	data *CCafoProfileUpdate,
) (*CCafoProfile, *CPerson, error) {
	if personID <= 0 || cafoID <= 0 || data == nil {
		return nil, nil, ErrNotFound
	}

	var current CCafoProfile
	var person CPerson
	err := m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		if err := session.QueryRowCtx(
			ctx,
			&current,
			"select "+cCafoProfileRows+" from "+m.table+" where `id` = ? and `person_id` = ? limit 1 for update",
			cafoID,
			personID,
		); err != nil {
			return err
		}
		if err := session.QueryRowCtx(
			ctx,
			&person,
			"select "+cPersonRows+" from `c_person` where `id` = ? limit 1 for update",
			personID,
		); err != nil {
			return err
		}
		if _, err := session.ExecCtx(
			ctx,
			"update "+m.table+" set `name` = ?, `alias` = ?, `chinese` = ?, `birth_day` = ?, `height` = ?, `cup` = ?, `bwh` = ?, `avtar` = ?, `biography` = ?, `updated_on` = ? where `id` = ? and `person_id` = ?",
			data.Name,
			data.Alias,
			data.Chinese,
			data.BirthDay,
			data.Height,
			data.Cup,
			data.Bwh,
			data.Avatar,
			data.Biography,
			data.UpdatedOn,
			cafoID,
			personID,
		); err != nil {
			return err
		}
		if _, err := session.ExecCtx(
			ctx,
			"update `c_person` set `name` = ?, `alias` = ?, `chinese` = ?, `birth_day` = ?, `height` = ?, `cup` = ?, `bwh` = ?, `avatar` = ?, `updated_on` = ? where `id` = ?",
			data.Name,
			data.Alias,
			data.Chinese,
			data.BirthDay,
			data.Height,
			data.Cup,
			data.Bwh,
			data.Avatar,
			data.UpdatedOn,
			personID,
		); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}

	updated := &CCafoProfile{
		Id:        current.Id,
		PersonId:  current.PersonId,
		Name:      data.Name,
		Alias:     data.Alias,
		Chinese:   data.Chinese,
		BirthDay:  data.BirthDay,
		Height:    data.Height,
		Cup:       data.Cup,
		Bwh:       data.Bwh,
		Avatar:    data.Avatar,
		Biography: data.Biography,
		CreatedOn: current.CreatedOn,
		UpdatedOn: data.UpdatedOn,
	}
	person.Name = data.Name
	person.Alias = data.Alias
	person.Chinese = data.Chinese
	person.BirthDay = data.BirthDay
	person.Height = data.Height
	person.Cup = data.Cup
	person.Bwh = data.Bwh
	person.Avatar = data.Avatar
	person.UpdatedOn = data.UpdatedOn

	cacheKeys := map[string]struct{}{
		fmt.Sprintf("%s%v", cacheRudyGcCCafoIdPrefix, current.Id):           {},
		fmt.Sprintf("%s%v", cacheRudyGcCCafoNamePrefix, current.Name):       {},
		fmt.Sprintf("%s%v", cacheRudyGcCCafoChinesePrefix, current.Chinese): {},
		fmt.Sprintf("%s%v", cacheRudyGcCCafoNamePrefix, updated.Name):       {},
		fmt.Sprintf("%s%v", cacheRudyGcCCafoChinesePrefix, updated.Chinese): {},
		fmt.Sprintf("%s%v", cacheRudyGcCPersonIdPrefix, personID):           {},
	}
	keys := make([]string, 0, len(cacheKeys))
	for key := range cacheKeys {
		keys = append(keys, key)
	}
	if err := m.DelCacheCtx(ctx, keys...); err != nil {
		return nil, nil, err
	}
	return updated, &person, nil
}

func (m *customCCafoModel) ListPersonLinks(ctx context.Context) (map[int64]int64, error) {
	type linkRow struct {
		CafoID   int64         `db:"cafo_id"`
		PersonID sql.NullInt64 `db:"person_id"`
	}

	var rows []*linkRow
	query := "select `id` as `cafo_id`, `person_id` from " + m.table + " where `person_id` is not null and `person_id` > 0 order by `id` asc"
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[int64]int64{}, nil
		}
		return nil, err
	}

	links := make(map[int64]int64, len(rows))
	for _, row := range rows {
		if row == nil || row.CafoID <= 0 || !row.PersonID.Valid || row.PersonID.Int64 <= 0 {
			continue
		}
		links[row.CafoID] = row.PersonID.Int64
	}
	return links, nil
}

func (m *customCCafoModel) SetPersonID(ctx context.Context, cafoID, personID int64) error {
	if cafoID <= 0 {
		return nil
	}
	var value any
	if personID > 0 {
		value = personID
	}
	_, err := m.ExecNoCacheCtx(ctx, "update "+m.table+" set `person_id` = ? where `id` = ?", value, cafoID)
	return err
}

func (m *customCCafoModel) ReassignPersonIDs(ctx context.Context, keepPersonID int64, sourcePersonIDs []int64) error {
	if keepPersonID <= 0 {
		return nil
	}
	sourceIDs := uniqueInt64s(sourcePersonIDs)
	filtered := sourceIDs[:0]
	for _, personID := range sourceIDs {
		if personID > 0 && personID != keepPersonID {
			filtered = append(filtered, personID)
		}
	}
	if len(filtered) == 0 {
		return nil
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(filtered)), ",")
	args := make([]any, 0, len(filtered)+1)
	args = append(args, keepPersonID)
	for _, personID := range filtered {
		args = append(args, personID)
	}
	_, err := m.ExecNoCacheCtx(ctx, "update "+m.table+" set `person_id` = ? where `person_id` in ("+placeholders+")", args...)
	return err
}
