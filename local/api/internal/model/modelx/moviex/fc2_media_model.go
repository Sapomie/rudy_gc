package moviex

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/local/api/internal/consts"
)

type FC2MediaModel interface {
	Insert(ctx context.Context, data *FC2Media) (*FC2Media, error)
	FindOne(ctx context.Context, id int64) (*FC2Media, error)
	FindOneByMovieID(ctx context.Context, movieID int64) (*FC2Media, error)
	FindOneByKey(ctx context.Context, key string) (*FC2Media, error)
	FindOneByFileName(ctx context.Context, fileName string) (*FC2Media, error)
	ListAll(ctx context.Context) ([]*FC2Media, error)
	ListByFileNames(ctx context.Context, fileNames []string) ([]*FC2Media, error)
	ListByKeys(ctx context.Context, keys []string) ([]*FC2Media, error)
	ListByMovieIDs(ctx context.Context, movieIDs []int64) ([]*FC2Media, error)
	CountByDirectoryID(ctx context.Context, directoryID int64) (int64, error)
	CountActiveByDirectoryIDs(ctx context.Context, directoryIDs []int64) (int64, error)
	ListByDirectoryIDs(ctx context.Context, directoryIDs []int64) ([]*FC2Media, error)
	ListPageByDirectoryIDs(ctx context.Context, directoryIDs []int64, offset, limit int64) ([]*FC2Media, error)
	ListByFullDirPrefixes(ctx context.Context, prefixes []string) ([]*FC2Media, error)
	ListSmartPickCandidates(ctx context.Context, filter FC2SmartPickMediaFilter) ([]*FC2Media, error)
	Update(ctx context.Context, data *FC2Media) error
	Delete(ctx context.Context, id int64) error
}

type FC2SmartPickMediaFilter struct {
	BirthTimeStart int64
	BirthTimeEnd   int64
	MediaDir1      string
	MediaDir2      string
	MediaDir3      string
	MediaDir4      string
	HasSub         *int64
	SelfMake       *int64
	HasMask        *int64
}

type FC2Media struct {
	ID                int64   `db:"id"`
	FC2MovieID        int64   `db:"fc2_movie_id"`
	FC2Key            string  `db:"fc2_key"`
	FC2Code           string  `db:"fc2_code"`
	SourceThreadID    int64   `db:"source_thread_id"`
	SourceTorrentHash string  `db:"source_torrent_hash"`
	FileName          string  `db:"file_name"`
	DirectoryID       int64   `db:"directory_id"`
	RootDir           string  `db:"root_dir"`
	FullDir           string  `db:"full_dir"`
	Alias             string  `db:"alias"`
	Size              int64   `db:"size"`
	Width             int64   `db:"width"`
	Height            int64   `db:"height"`
	BitRate           int64   `db:"bit_rate"`
	Duration          int64   `db:"duration"`
	FrameAverage      float64 `db:"frame_average"`
	HasSub            int64   `db:"has_sub"`
	SelfMake          int64   `db:"self_make"`
	HasMask           int64   `db:"has_mask"`
	NeedScanMeta      int64   `db:"need_scan_meta"`
	IsRemoved         int64   `db:"is_removed"`
	RemoveTime        int64   `db:"remove_time"`
	BirthTime         int64   `db:"birth_time"`
	CreatedOn         int64   `db:"created_on"`
	UpdatedOn         int64   `db:"updated_on"`
}

type defaultFC2MediaModel struct {
	conn sqlx.SqlConn
}

const fc2MediaRows = "`id`,`fc2_movie_id`,`fc2_key`,`fc2_code`,`source_thread_id`,`source_torrent_hash`,`file_name`,`directory_id`,`root_dir`,`full_dir`,`alias`,`size`,`width`,`height`,`bit_rate`,`duration`,`frame_average`,`has_sub`,`self_make`,`has_mask`,`need_scan_meta`,`is_removed`,`remove_time`,`birth_time`,`created_on`,`updated_on`"

func NewFC2MediaModel(conn sqlx.SqlConn) FC2MediaModel {
	return &defaultFC2MediaModel{conn: conn}
}

func (m *defaultFC2MediaModel) Insert(ctx context.Context, data *FC2Media) (*FC2Media, error) {
	result, err := m.conn.ExecCtx(ctx, `INSERT INTO fc2_media
(`+"`fc2_movie_id`,`fc2_key`,`fc2_code`,`source_thread_id`,`source_torrent_hash`,`file_name`,`directory_id`,`root_dir`,`full_dir`,`alias`,`size`,`width`,`height`,`bit_rate`,`duration`,`frame_average`,`has_sub`,`self_make`,`has_mask`,`need_scan_meta`,`is_removed`,`remove_time`,`birth_time`,`created_on`,`updated_on`"+`)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		data.FC2MovieID, data.FC2Key, data.FC2Code, data.SourceThreadID, data.SourceTorrentHash, data.FileName,
		data.DirectoryID, data.RootDir, data.FullDir, data.Alias, data.Size, data.Width, data.Height,
		data.BitRate, data.Duration, data.FrameAverage, data.HasSub, data.SelfMake, data.HasMask,
		data.NeedScanMeta, data.IsRemoved, data.RemoveTime, data.BirthTime, data.CreatedOn, data.UpdatedOn)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return m.FindOne(ctx, id)
}

func (m *defaultFC2MediaModel) FindOne(ctx context.Context, id int64) (*FC2Media, error) {
	var row FC2Media
	if err := m.conn.QueryRowCtx(ctx, &row, "SELECT "+fc2MediaRows+" FROM `fc2_media` WHERE `id` = ? LIMIT 1", id); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultFC2MediaModel) FindOneByMovieID(ctx context.Context, movieID int64) (*FC2Media, error) {
	var row FC2Media
	if err := m.conn.QueryRowCtx(ctx, &row, "SELECT "+fc2MediaRows+" FROM `fc2_media` WHERE `fc2_movie_id` = ? LIMIT 1", movieID); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultFC2MediaModel) FindOneByKey(ctx context.Context, key string) (*FC2Media, error) {
	var row FC2Media
	if err := m.conn.QueryRowCtx(ctx, &row, "SELECT "+fc2MediaRows+" FROM `fc2_media` WHERE `fc2_key` = ? LIMIT 1", strings.TrimSpace(key)); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultFC2MediaModel) FindOneByFileName(ctx context.Context, fileName string) (*FC2Media, error) {
	var row FC2Media
	if err := m.conn.QueryRowCtx(ctx, &row, "SELECT "+fc2MediaRows+" FROM `fc2_media` WHERE `file_name` = ? LIMIT 1", strings.TrimSpace(fileName)); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultFC2MediaModel) ListAll(ctx context.Context) ([]*FC2Media, error) {
	rows := make([]*FC2Media, 0)
	if err := m.conn.QueryRowsCtx(ctx, &rows, "SELECT "+fc2MediaRows+" FROM `fc2_media` ORDER BY `id` ASC"); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*FC2Media{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultFC2MediaModel) ListByFileNames(ctx context.Context, fileNames []string) ([]*FC2Media, error) {
	return m.listByStringValues(ctx, "file_name", fileNames)
}

func (m *defaultFC2MediaModel) ListByKeys(ctx context.Context, keys []string) ([]*FC2Media, error) {
	return m.listByStringValues(ctx, "fc2_key", keys)
}

func (m *defaultFC2MediaModel) listByStringValues(ctx context.Context, column string, values []string) ([]*FC2Media, error) {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	if len(normalized) == 0 {
		return []*FC2Media{}, nil
	}
	if column != "file_name" && column != "fc2_key" {
		return nil, errors.New("invalid fc2_media string column")
	}
	rows := make([]*FC2Media, 0, len(normalized))
	for _, batch := range sqlValueBatches(normalized, 0) {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(batch)), ",")
		args := make([]any, len(batch))
		for index, value := range batch {
			args[index] = value
		}
		var batchRows []*FC2Media
		if err := m.conn.QueryRowsCtx(ctx, &batchRows, "SELECT "+fc2MediaRows+" FROM `fc2_media` WHERE `"+column+"` IN ("+placeholders+")", args...); err != nil {
			if errors.Is(err, sqlx.ErrNotFound) {
				continue
			}
			return nil, err
		}
		rows = append(rows, batchRows...)
	}
	return rows, nil
}

func (m *defaultFC2MediaModel) ListByMovieIDs(ctx context.Context, movieIDs []int64) ([]*FC2Media, error) {
	ids := make([]int64, 0, len(movieIDs))
	seen := make(map[int64]struct{}, len(movieIDs))
	for _, movieID := range movieIDs {
		if movieID <= 0 {
			continue
		}
		if _, exists := seen[movieID]; exists {
			continue
		}
		seen[movieID] = struct{}{}
		ids = append(ids, movieID)
	}
	if len(ids) == 0 {
		return []*FC2Media{}, nil
	}
	rows := make([]*FC2Media, 0, len(ids))
	for _, batch := range sqlValueBatches(ids, 0) {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(batch)), ",")
		args := make([]any, 0, len(batch))
		for _, movieID := range batch {
			args = append(args, movieID)
		}
		var batchRows []*FC2Media
		if err := m.conn.QueryRowsCtx(ctx, &batchRows, "SELECT "+fc2MediaRows+" FROM `fc2_media` WHERE `fc2_movie_id` IN ("+placeholders+")", args...); err != nil {
			if errors.Is(err, sqlx.ErrNotFound) {
				continue
			}
			return nil, err
		}
		rows = append(rows, batchRows...)
	}
	return rows, nil
}

func (m *defaultFC2MediaModel) CountByDirectoryID(ctx context.Context, directoryID int64) (int64, error) {
	if directoryID <= 0 {
		return 0, nil
	}
	var count int64
	if err := m.conn.QueryRowCtx(ctx, &count, "SELECT COUNT(*) FROM `fc2_media` WHERE `directory_id` = ?", directoryID); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *defaultFC2MediaModel) CountActiveByDirectoryIDs(ctx context.Context, directoryIDs []int64) (int64, error) {
	ids := uniquePositiveInt64s(directoryIDs)
	if len(ids) == 0 {
		return 0, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, 0, len(ids)+1)
	args = append(args, int64(consts.FilmIsNotRemoved))
	for _, directoryID := range ids {
		args = append(args, directoryID)
	}
	var count int64
	query := "SELECT COUNT(*) FROM `fc2_media` WHERE `is_removed` = ? AND `directory_id` IN (" + placeholders + ")"
	if err := m.conn.QueryRowCtx(ctx, &count, query, args...); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *defaultFC2MediaModel) ListByDirectoryIDs(ctx context.Context, directoryIDs []int64) ([]*FC2Media, error) {
	ids := uniquePositiveInt64s(directoryIDs)
	if len(ids) == 0 {
		return []*FC2Media{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, 0, len(ids)+1)
	args = append(args, int64(consts.FilmIsNotRemoved))
	for _, directoryID := range ids {
		args = append(args, directoryID)
	}
	rows := make([]*FC2Media, 0)
	query := "SELECT " + fc2MediaRows + " FROM `fc2_media` WHERE `is_removed` = ? AND `directory_id` IN (" + placeholders + ") ORDER BY `birth_time` DESC, `id` DESC"
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*FC2Media{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultFC2MediaModel) ListPageByDirectoryIDs(ctx context.Context, directoryIDs []int64, offset, limit int64) ([]*FC2Media, error) {
	ids := uniquePositiveInt64s(directoryIDs)
	if len(ids) == 0 || limit <= 0 {
		return []*FC2Media{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, 0, len(ids)+3)
	args = append(args, int64(consts.FilmIsNotRemoved))
	for _, directoryID := range ids {
		args = append(args, directoryID)
	}
	args = append(args, limit)
	query := "SELECT " + fc2MediaRows + " FROM `fc2_media` WHERE `is_removed` = ? AND `directory_id` IN (" + placeholders + ") ORDER BY `birth_time` DESC, `id` DESC LIMIT ?"
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}
	rows := make([]*FC2Media, 0)
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*FC2Media{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultFC2MediaModel) ListByFullDirPrefixes(ctx context.Context, prefixes []string) ([]*FC2Media, error) {
	clauses := make([]string, 0, len(prefixes))
	args := make([]any, 0, len(prefixes)*2)
	for _, prefix := range prefixes {
		prefix = filepath.Clean(strings.TrimSpace(prefix))
		if prefix == "." {
			continue
		}
		clauses = append(clauses, "(`full_dir` = ? OR `full_dir` LIKE ?)")
		args = append(args, prefix, prefix+string(filepath.Separator)+"%")
	}
	if len(clauses) == 0 {
		return []*FC2Media{}, nil
	}
	var rows []*FC2Media
	query := "SELECT " + fc2MediaRows + " FROM `fc2_media` WHERE " + strings.Join(clauses, " OR ")
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*FC2Media{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultFC2MediaModel) ListSmartPickCandidates(ctx context.Context, filter FC2SmartPickMediaFilter) ([]*FC2Media, error) {
	builder := buildFC2SmartPickMediaQuery(filter)
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	rows := make([]*FC2Media, 0)
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*FC2Media{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func buildFC2SmartPickMediaQuery(filter FC2SmartPickMediaFilter) squirrel.SelectBuilder {
	builder := squirrel.
		Select(fc2MediaRows).
		From("`fc2_media`").
		Where("`fc2_movie_id` > 0").
		Where("`is_removed` = ?", consts.FilmIsNotRemoved).
		Where("`full_dir` <> ''").
		Where("`file_name` <> ''")
	if filter.BirthTimeStart > 0 {
		builder = builder.Where("`birth_time` >= ?", filter.BirthTimeStart)
	}
	if filter.BirthTimeEnd > 0 {
		builder = builder.Where("`birth_time` <= ?", filter.BirthTimeEnd)
	}
	for _, mediaDir := range []string{filter.MediaDir1, filter.MediaDir2, filter.MediaDir3, filter.MediaDir4} {
		mediaDir = strings.Trim(strings.TrimSpace(mediaDir), "/\\")
		if mediaDir != "" {
			builder = builder.Where("CONCAT('/', `full_dir`, '/') LIKE ?", "%/"+mediaDir+"/%")
		}
	}
	if filter.HasSub != nil {
		builder = builder.Where("`has_sub` = ?", *filter.HasSub)
	}
	if filter.SelfMake != nil {
		builder = builder.Where("`self_make` = ?", *filter.SelfMake)
	}
	if filter.HasMask != nil {
		builder = builder.Where("`has_mask` = ?", *filter.HasMask)
	}
	return builder.OrderBy("`birth_time` DESC", "`id` DESC")
}

func (m *defaultFC2MediaModel) Update(ctx context.Context, data *FC2Media) error {
	_, err := m.conn.ExecCtx(ctx, `UPDATE fc2_media SET
fc2_movie_id = ?, fc2_key = ?, fc2_code = ?, source_thread_id = ?, source_torrent_hash = ?, file_name = ?,
directory_id = ?, root_dir = ?, full_dir = ?, alias = ?, size = ?, width = ?, height = ?,
bit_rate = ?, duration = ?, frame_average = ?, has_sub = ?, self_make = ?, has_mask = ?,
need_scan_meta = ?, is_removed = ?, remove_time = ?, birth_time = ?, updated_on = ?
WHERE id = ?`, data.FC2MovieID, data.FC2Key, data.FC2Code, data.SourceThreadID, data.SourceTorrentHash,
		data.FileName, data.DirectoryID, data.RootDir, data.FullDir, data.Alias, data.Size,
		data.Width, data.Height, data.BitRate, data.Duration, data.FrameAverage, data.HasSub,
		data.SelfMake, data.HasMask, data.NeedScanMeta, data.IsRemoved, data.RemoveTime,
		data.BirthTime, data.UpdatedOn, data.ID)
	return err
}

func (m *defaultFC2MediaModel) Delete(ctx context.Context, id int64) error {
	_, err := m.conn.ExecCtx(ctx, "DELETE FROM `fc2_media` WHERE `id` = ?", id)
	return err
}
