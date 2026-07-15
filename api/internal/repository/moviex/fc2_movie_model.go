package moviex

import (
	"context"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type FC2MovieModel interface {
	Upsert(ctx context.Context, data *FC2Movie) (*FC2Movie, error)
	FindOne(ctx context.Context, id int64) (*FC2Movie, error)
	FindOneByKey(ctx context.Context, key string) (*FC2Movie, error)
	ListByIDs(ctx context.Context, ids []int64) ([]*FC2Movie, error)
	ListByCode(ctx context.Context, code string) ([]*FC2Movie, error)
	ListPage(ctx context.Context, offset, limit int64, keyword, status, sortField, sortOrder string) ([]*FC2Movie, error)
	CountAll(ctx context.Context, keyword, status string) (int64, error)
	ListCards(ctx context.Context, offset, limit int64, ownedOnly, randomOrder bool) ([]*FC2Movie, error)
	CountCards(ctx context.Context, ownedOnly bool) (int64, error)
	UpdateMediaSnapshot(ctx context.Context, id int64, present int64, birthTime, updatedOn int64) error
	UpdateCoverImage(ctx context.Context, id, coverImageID, updatedOn int64) error
}

type FC2Movie struct {
	ID                int64  `db:"id"`
	FC2Key            string `db:"fc2_key"`
	Code              string `db:"code"`
	CodeSeq           int64  `db:"code_seq"`
	PPVNo             int64  `db:"ppv_no"`
	Title             string `db:"title"`
	PerformerText     string `db:"performer_text"`
	MaskType          string `db:"mask_type"`
	MetadataThreadID  int64  `db:"metadata_thread_id"`
	MetadataThreadTID string `db:"metadata_thread_tid"`
	PostTime          int64  `db:"post_time"`
	MediaPresent      int64  `db:"media_present"`
	MediaBirthTime    int64  `db:"media_birth_time"`
	CoverImageID      int64  `db:"cover_image_id"`
	DataStatus        string `db:"data_status"`
	CreatedOn         int64  `db:"created_on"`
	UpdatedOn         int64  `db:"updated_on"`
}

type defaultFC2MovieModel struct {
	conn sqlx.SqlConn
}

const fc2MovieRows = "`id`,`fc2_key`,`code`,`code_seq`,`ppv_no`,`title`,`performer_text`,`mask_type`,`metadata_thread_id`,'' AS `metadata_thread_tid`,`post_time`,`media_present`,`media_birth_time`,`cover_image_id`,`data_status`,`created_on`,`updated_on`"

const fc2MovieAliasedRows = "`fm`.`id`,`fm`.`fc2_key`,`fm`.`code`,`fm`.`code_seq`,`fm`.`ppv_no`,`fm`.`title`,`fm`.`performer_text`,`fm`.`mask_type`,`fm`.`metadata_thread_id`,'' AS `metadata_thread_tid`,`fm`.`post_time`,`fm`.`media_present`,`fm`.`media_birth_time`,`fm`.`cover_image_id`,`fm`.`data_status`,`fm`.`created_on`,`fm`.`updated_on`"

func NewFC2MovieModel(conn sqlx.SqlConn) FC2MovieModel {
	return &defaultFC2MovieModel{conn: conn}
}

func (m *defaultFC2MovieModel) Upsert(ctx context.Context, data *FC2Movie) (*FC2Movie, error) {
	query := `INSERT INTO fc2_movie
(` + "`fc2_key`,`code`,`code_seq`,`ppv_no`,`title`,`performer_text`,`mask_type`,`metadata_thread_id`,`post_time`,`data_status`,`created_on`,`updated_on`" + `)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
fc2_key = VALUES(fc2_key),
code = VALUES(code),
code_seq = VALUES(code_seq),
ppv_no = VALUES(ppv_no),
title = VALUES(title),
performer_text = VALUES(performer_text),
mask_type = VALUES(mask_type),
metadata_thread_id = VALUES(metadata_thread_id),
post_time = VALUES(post_time),
data_status = VALUES(data_status),
updated_on = VALUES(updated_on)`
	if _, err := m.conn.ExecCtx(ctx, query, data.FC2Key, data.Code, data.CodeSeq, data.PPVNo, data.Title, data.PerformerText, data.MaskType, data.MetadataThreadID, data.PostTime, data.DataStatus, data.CreatedOn, data.UpdatedOn); err != nil {
		return nil, err
	}
	return m.FindOneByKey(ctx, data.FC2Key)
}

func (m *defaultFC2MovieModel) FindOne(ctx context.Context, id int64) (*FC2Movie, error) {
	var row FC2Movie
	query := "SELECT " + fc2MovieRows + " FROM `fc2_movie` WHERE `id` = ? LIMIT 1"
	if err := m.conn.QueryRowCtx(ctx, &row, query, id); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultFC2MovieModel) FindOneByKey(ctx context.Context, key string) (*FC2Movie, error) {
	var row FC2Movie
	query := "SELECT " + fc2MovieRows + " FROM `fc2_movie` WHERE `fc2_key` = ? LIMIT 1"
	if err := m.conn.QueryRowCtx(ctx, &row, query, strings.TrimSpace(key)); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultFC2MovieModel) ListByIDs(ctx context.Context, ids []int64) ([]*FC2Movie, error) {
	values := uniquePositiveInt64s(ids)
	if len(values) == 0 {
		return []*FC2Movie{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(values)), ",")
	args := make([]any, len(values))
	for index, id := range values {
		args[index] = id
	}
	var rows []*FC2Movie
	if err := m.conn.QueryRowsCtx(ctx, &rows, "SELECT "+fc2MovieRows+" FROM `fc2_movie` WHERE `id` IN ("+placeholders+")", args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*FC2Movie{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultFC2MovieModel) ListByCode(ctx context.Context, code string) ([]*FC2Movie, error) {
	var rows []*FC2Movie
	query := "SELECT " + fc2MovieRows + " FROM `fc2_movie` WHERE `code` = ? ORDER BY `code_seq` ASC, `id` ASC"
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, strings.TrimSpace(code)); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*FC2Movie{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultFC2MovieModel) ListPage(ctx context.Context, offset, limit int64, keyword, status, sortField, sortOrder string) ([]*FC2Movie, error) {
	builder := buildFC2MovieListQuery(offset, limit, keyword, status, sortField, sortOrder)
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*FC2Movie
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*FC2Movie{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func buildFC2MovieListQuery(offset, limit int64, keyword, status, sortField, sortOrder string) squirrel.SelectBuilder {
	var builder squirrel.SelectBuilder
	if isFC2ScStatSort(sortField) {
		builder = applyAliasedFC2MovieFilter(
			squirrel.Select(fc2MovieAliasedRows).
				From("`fc2_movie` AS `fm`").
				LeftJoin("`fc2_sc_stat` AS `fss` ON `fss`.`fc2_movie_id` = `fm`.`id`"),
			keyword,
			status,
		)
	} else {
		builder = applyFC2MovieFilter(squirrel.Select(fc2MovieRows).From("`fc2_movie`"), keyword, status)
	}
	builder = builder.OrderBy(fc2MovieOrderBy(sortField, sortOrder)...)
	if offset > 0 {
		builder = builder.Offset(uint64(offset))
	}
	if limit > 0 {
		builder = builder.Limit(uint64(limit))
	}
	return builder
}

func fc2MovieOrderBy(sortField, sortOrder string) []string {
	direction := "DESC"
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		direction = "ASC"
	}
	var primary string
	switch strings.ToLower(strings.TrimSpace(sortField)) {
	case "title":
		primary = "`title` " + direction
	case "performer_text":
		primary = "`performer_text` " + direction
	case "post_time":
		primary = "`post_time` " + direction
	case "play":
		primary = "`media_present` " + direction
	case "birth_time":
		primary = "`media_birth_time` " + direction
	case "sc_times":
		return fc2ScStatOrderBy("`fss`.`sc_times`", direction)
	case "come_times":
		return fc2ScStatOrderBy("`fss`.`come_times`", direction)
	case "last_selected_sc_time":
		return fc2ScStatOrderBy("`fss`.`last_selected_sc_time`", direction)
	default:
		return []string{"`ppv_no` " + direction, "`code_seq` " + direction, "`id` " + direction}
	}
	return []string{primary, "`ppv_no` DESC", "`code_seq` ASC", "`id` DESC"}
}

func isFC2ScStatSort(sortField string) bool {
	switch strings.ToLower(strings.TrimSpace(sortField)) {
	case "sc_times", "come_times", "last_selected_sc_time":
		return true
	default:
		return false
	}
}

func fc2ScStatOrderBy(column, direction string) []string {
	return []string{
		"COALESCE(" + column + ", 0) " + direction,
		"`fm`.`ppv_no` DESC",
		"`fm`.`code_seq` ASC",
		"`fm`.`id` DESC",
	}
}

func (m *defaultFC2MovieModel) UpdateMediaSnapshot(ctx context.Context, id int64, present int64, birthTime, updatedOn int64) error {
	if id <= 0 {
		return nil
	}
	if present != 0 {
		present = 1
	}
	_, err := m.conn.ExecCtx(ctx, "UPDATE `fc2_movie` SET `media_present` = ?, `media_birth_time` = ?, `updated_on` = ? WHERE `id` = ?", present, birthTime, updatedOn, id)
	return err
}

func (m *defaultFC2MovieModel) UpdateCoverImage(ctx context.Context, id, coverImageID, updatedOn int64) error {
	if id <= 0 || coverImageID < 0 {
		return nil
	}
	_, err := m.conn.ExecCtx(ctx, "UPDATE `fc2_movie` SET `cover_image_id` = ?, `updated_on` = ? WHERE `id` = ?", coverImageID, updatedOn, id)
	return err
}

func (m *defaultFC2MovieModel) CountAll(ctx context.Context, keyword, status string) (int64, error) {
	builder := applyFC2MovieFilter(squirrel.Select("COUNT(*) AS total").From("`fc2_movie`"), keyword, status)
	query, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}
	var row struct {
		Total int64 `db:"total"`
	}
	if err := m.conn.QueryRowCtx(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return row.Total, nil
}

func (m *defaultFC2MovieModel) ListCards(ctx context.Context, offset, limit int64, ownedOnly, randomOrder bool) ([]*FC2Movie, error) {
	builder := applyFC2CardFilter(squirrel.Select(fc2MovieRows).From("`fc2_movie`"), ownedOnly)
	if randomOrder {
		builder = builder.OrderBy("RAND()")
	} else {
		builder = builder.OrderBy("`post_time` DESC", "`ppv_no` DESC", "`code_seq` ASC", "`id` DESC")
	}
	if offset > 0 {
		builder = builder.Offset(uint64(offset))
	}
	if limit > 0 {
		builder = builder.Limit(uint64(limit))
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*FC2Movie
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*FC2Movie{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultFC2MovieModel) CountCards(ctx context.Context, ownedOnly bool) (int64, error) {
	builder := applyFC2CardFilter(squirrel.Select("COUNT(*) AS total").From("`fc2_movie`"), ownedOnly)
	query, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}
	var row struct {
		Total int64 `db:"total"`
	}
	if err := m.conn.QueryRowCtx(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return row.Total, nil
}

func applyFC2CardFilter(builder squirrel.SelectBuilder, ownedOnly bool) squirrel.SelectBuilder {
	if ownedOnly {
		builder = builder.Where("`media_present` = 1")
	}
	return builder
}

func applyFC2MovieFilter(builder squirrel.SelectBuilder, keyword, status string) squirrel.SelectBuilder {
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		builder = builder.Where(squirrel.Or{squirrel.Like{"fc2_key": like}, squirrel.Like{"code": like}, squirrel.Like{"title": like}, squirrel.Like{"performer_text": like}})
	}
	if status = strings.TrimSpace(status); status != "" {
		builder = builder.Where("`data_status` = ?", status)
	}
	return builder
}

func applyAliasedFC2MovieFilter(builder squirrel.SelectBuilder, keyword, status string) squirrel.SelectBuilder {
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		builder = builder.Where(squirrel.Or{
			squirrel.Like{"fm.fc2_key": like},
			squirrel.Like{"fm.code": like},
			squirrel.Like{"fm.title": like},
			squirrel.Like{"fm.performer_text": like},
		})
	}
	if status = strings.TrimSpace(status); status != "" {
		builder = builder.Where("`fm`.`data_status` = ?", status)
	}
	return builder
}
