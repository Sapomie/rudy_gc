package modelx

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlc"

	"rudy_gc/pkg/orm"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ VAlbumModel = (*customVAlbumModel)(nil)
var _ orm.DataModel = (VAlbumModel)(nil) // VAlbumModel implements orm.DataModel
var _ orm.DataStruct = (*VAlbum)(nil)    // VAlbum implements orm.DataStruct

type (
	// VAlbumModel is an interface that extends vAlbumModel and allows additional methods
	VAlbumModel interface {
		vAlbumModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error

		FindAll(ctx context.Context) ([]*VAlbum, error)
	}

	customVAlbumModel struct {
		*defaultVAlbumModel
	}
)

const (
	AlbumHasNotUploaded = 1 + iota
	AlbumHasUploaded
)

// NewVAlbumModel returns a model for the database table.
func NewVAlbumModel(conn sqlx.SqlConn) VAlbumModel {
	return &customVAlbumModel{
		defaultVAlbumModel: newVAlbumModel(conn),
	}
}

// Description returns the VAlbum's description, to be customized based on your field.
func (v *VAlbum) Description() interface{} {
	return v.Name // Replace with actual field
}

func (m *defaultVAlbumModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	name, ok := description.(string) // or other type based on VAlbum struct
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByName(ctx, name) // Replace with actual method
}

func (m *defaultVAlbumModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

func (m *defaultVAlbumModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*VAlbum); ok {
		insertData.CreatedOn = time.Now().Unix() // Set creation time
		_, err := m.Insert(ctx, insertData)      // Actual insert operation
		return err
	}
	return ErrInvalidData
}

func (m *defaultVAlbumModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*VAlbum); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*VAlbum); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix() // Set update time
			return m.Update(ctx, newData)         // Actual update operation
		}
		return ErrInvalidData
	}
	return ErrInvalidData
}

func (m *customVAlbumModel) FindAll(ctx context.Context) ([]*VAlbum, error) {
	query, values, err := squirrel.Select("*").
		From(m.tableName()).
		Where("`film_number` > 0").
		Limit(100000).
		ToSql()

	if err != nil {
		return nil, err
	}

	var result []*VAlbum
	if err := m.conn.QueryRowsCtx(ctx, &result, query, values...); err != nil {
		return nil, err
	}

	return result, nil
}
