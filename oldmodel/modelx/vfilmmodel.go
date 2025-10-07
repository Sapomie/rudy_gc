package modelx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"

	"rudy_gc/pkg/orm"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ VFilmModel = (*customVFilmModel)(nil)

type (
	// VFilmModel is an interface to be customized, add more methods here,
	// and implement the added methods in customVFilmModel.
	VFilmModel interface {
		vFilmModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error
		FindAll(ctx context.Context, existStatus int) ([]*VFilm, error)
		FindFilm2(params *FindFilmParams2) ([]*VFilm, int64, error)
		FindFilmByAlbumId(ctx context.Context, albumId int64) ([]*VFilm, error)
	}

	customVFilmModel struct {
		*defaultVFilmModel
	}
)

const (
	FilmHasSub     = 2
	FilmNoSub      = 1
	FilmSelfMake   = 2
	FilmNoSelfMake = 1
)

const (
	FilmMetaDataNeedScan = 1 + iota
	FilmMetaDataNoNeedScan
)

const (
	FilmNotErased = 1 + iota
	FilmErased
	FilmNoMosaic
)

const (
	FilmIsNotRemoved = 1 + iota
	FilmIsRemoved
)

// NewVFilmModel returns a model for the database table.
func NewVFilmModel(conn sqlx.SqlConn) VFilmModel {
	return &customVFilmModel{
		defaultVFilmModel: newVFilmModel(conn),
	}
}

func (f *VFilm) Description() interface{} {
	return f.MovieJavId
}

func (m *defaultVFilmModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	javId, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByMovieJavId(ctx, javId)
}

func (m *defaultVFilmModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if errors.Is(err, sqlc.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (m *defaultVFilmModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*VFilm); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultVFilmModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*VFilm); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*VFilm); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData)
		}
		return fmt.Errorf("invalid data")
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultVFilmModel) FindAll(ctx context.Context, existStatus int) ([]*VFilm, error) {
	// 使用 squirrel 构建 SQL 查询
	s := squirrel.Select("*").From(m.tableName())
	if existStatus != 0 {
		s = s.Where("is_removed = ?", existStatus)
	}

	query, args, err := s.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var result []*VFilm
	// 执行查询
	if err := m.conn.QueryRowsCtx(ctx, &result, query, args...); err != nil {
		return nil, err
	}

	return result, nil
}

func (m *defaultVFilmModel) FindFilmByAlbumId(ctx context.Context, albumId int64) ([]*VFilm, error) {
	// 使用 squirrel 构建 SQL 查询
	query, args, err := squirrel.Select("*").
		From(m.tableName()).
		Where("`album_id` = ?", albumId).ToSql()
	if err != nil {
		return nil, err
	}

	var result []*VFilm
	// 执行查询
	if err := m.conn.QueryRowsCtx(ctx, &result, query, args...); err != nil {
		if errors.Is(err, ErrNotFound) {
			return result, nil
		}
		return nil, err
	}

	return result, nil
}

type FindFilmParams2 struct {
	Ctx                context.Context
	Dir1               []string
	Dir2               []string
	Dir3               []string
	Dir4               []string
	HasSub             int64
	Page               int64
	PageSize           int64
	FilmBirthTimeStart int64
	FilmBirthTimeEnd   int64
	ComeTimesMin       int64
	ScTimesMin         int64
	ScTimesMax         int64
	LastScTimeMin      int64
}

func (m *defaultVFilmModel) FindFilm2(params *FindFilmParams2) ([]*VFilm, int64, error) {
	// 使用 squirrel 构建 SQL 查询
	db := squirrel.Select("*").
		From(m.tableName()).Where("is_removed = ?", FilmIsNotRemoved)

	// Helper function to add OR clauses for each Dir array
	addDirConditions := func(db squirrel.SelectBuilder, dirs []string, column string) squirrel.SelectBuilder {
		if len(dirs) > 0 {
			orClause := squirrel.Or{}
			for _, dir := range dirs {
				if dir != "" {
					orClause = append(orClause, squirrel.Eq{column: dir})
				}
			}
			if len(orClause) > 0 {
				db = db.Where(orClause)
			}
		}
		return db
	}

	// Add conditions for each Dir
	db = addDirConditions(db, params.Dir1, "dir1")
	db = addDirConditions(db, params.Dir2, "dir2")
	db = addDirConditions(db, params.Dir3, "dir3")
	db = addDirConditions(db, params.Dir4, "dir4")

	if params.HasSub != 0 {
		db = db.Where("has_sub = ?", params.HasSub)
	}

	if params.FilmBirthTimeStart > 0 {
		db = db.Where("`birth_time` >= ?", params.FilmBirthTimeStart)
	}
	if params.FilmBirthTimeEnd > 0 {
		db = db.Where("`birth_time` <= ?", params.FilmBirthTimeEnd)
	}
	if params.ScTimesMin > 0 {
		db = db.Where("`sc_times` >= ?", params.ScTimesMin)
	}

	db = db.Where("`sc_times` <= ?", params.ScTimesMax)

	if params.ComeTimesMin > 0 {
		db = db.Where("`come_times` >= ?", params.ComeTimesMin)
	}
	if params.LastScTimeMin > 0 {
		db = db.Where("`last_sc_time` >= ?", params.LastScTimeMin)
	}
	if params.Page == 0 {
		params.Page = 1
	}

	var total int64
	countQuery, countArgs, err := db.ToSql()
	if err != nil {
		return nil, 0, err
	}
	finalCountQuery := "SELECT COUNT(*) FROM (" + countQuery + ") AS count_query"
	if err := m.conn.QueryRowCtx(params.Ctx, &total, finalCountQuery, countArgs...); err != nil {
		return nil, 0, err
	}

	query, args, err := db.Limit(uint64(params.PageSize)).Offset(uint64((params.Page - 1) * params.PageSize)).OrderBy("birth_time desc").ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var result []*VFilm
	// 执行查询
	if err := m.conn.QueryRowsCtx(params.Ctx, &result, query, args...); err != nil {
		return nil, 0, err
	}

	return result, total, nil
}
