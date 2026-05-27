package orm

import "context"

type DataModel interface {
	FindDataByDescription(ctx context.Context, description any) (DataStruct, error)
	ExistByDescription(ctx context.Context, description any) (bool, error)
	InsertData(ctx context.Context, data DataStruct) error
	UpdateDataByDescription(ctx context.Context, data DataStruct) error
}

type DataStruct interface {
	Description() interface{}
}

func InsertOrUpdate(ctx context.Context, dataModel DataModel, data DataStruct) (DataStruct, bool, error) {
	return upsertData(ctx, dataModel, data, true)
}

func TryInsert(ctx context.Context, dataModel DataModel, data DataStruct) (DataStruct, bool, error) {
	return upsertData(ctx, dataModel, data, false)
}

func upsertData(ctx context.Context, dataModel DataModel, data DataStruct, allowUpdate bool) (dataDb DataStruct, exist bool, err error) {
	exist, err = dataModel.ExistByDescription(ctx, data.Description())
	if err != nil {
		return nil, false, err
	}

	if !exist {
		err = dataModel.InsertData(ctx, data)
		if err != nil {
			return nil, false, err
		}
	} else if allowUpdate {
		err = dataModel.UpdateDataByDescription(ctx, data)
		if err != nil {
			return nil, false, err
		}
	}

	dataDb, err = dataModel.FindDataByDescription(ctx, data.Description())
	if err != nil {
		return nil, false, err
	}

	return dataDb, exist, nil
}
