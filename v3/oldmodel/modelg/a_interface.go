package modelg

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

var ErrNotFound = gorm.ErrRecordNotFound

type DataModel interface {
	FindDataByDiscrimination(ctx context.Context, name string) (DataStruct, error)
	ExistByDiscrimination(ctx context.Context, name string) (bool, error)
	InsertData(ctx context.Context, data DataStruct) error
	UpdateDataByDiscrimination(ctx context.Context, data DataStruct) error
}

type DataStruct interface {
	Discrimination() string
	TableName() string
}

func InsertOrUpdate(ctx context.Context, dataModel DataModel, data DataStruct) (dataDb DataStruct, info string, err error) {

	exist, err := dataModel.ExistByDiscrimination(ctx, data.Discrimination())
	if err != nil && err != ErrNotFound {
		return nil, "", err //todo err handle
	}

	if !exist {
		err = dataModel.InsertData(ctx, data)
		if err != nil {
			return nil, "", err
		}
		info = fmt.Sprintf("Add  %v to %v", data.Discrimination(), data.TableName())
	} else {
		err = dataModel.UpdateDataByDiscrimination(ctx, data)
		if err != nil {
			return nil, "", err
		}
	}

	dataDb, err = dataModel.FindDataByDiscrimination(ctx, data.Discrimination())
	if err != nil {
		return nil, "", err
	}

	return dataDb, info, nil
}

func TryInsert(ctx context.Context, dataModel DataModel, data DataStruct) (dataDb DataStruct, info string, err error) {

	exist, err := dataModel.ExistByDiscrimination(ctx, data.Discrimination())
	if err != nil && err != ErrNotFound {
		return nil, "", err
	}

	if !exist {
		err = dataModel.InsertData(ctx, data)
		if err != nil {
			return nil, "", err
		}
		info = fmt.Sprintf("Add  %v to %v", data.Discrimination(), data.TableName())
	}

	dataDb, err = dataModel.FindDataByDiscrimination(ctx, data.Discrimination())
	if err != nil {
		return nil, "", err
	}

	return dataDb, info, nil
}

func makeFakeNameMovieColumn(movieId, columnId int64) string {
	cIdStr := fmt.Sprintf("%04d", columnId)
	mIdStr := fmt.Sprintf("%06d", movieId)

	return cIdStr + mIdStr
}

func decodeFakeNameMovieColumn(name string) (movieId, columnId int64, err error) {
	if len(name) != 10 {
		return 0, 0, errors.New("illegal length of fake name")
	}
	cId, err := strconv.Atoi(name[:4])
	if err != nil {
		return 0, 0, err
	}
	mId, err := strconv.Atoi(name[4:])
	if err != nil {
		return 0, 0, err
	}
	return int64(mId), int64(cId), nil
}

func makeFakeNameDateRank(date string, rank int64) string {
	return fmt.Sprintf("%v:%03d", date, rank)
}

func decodeFakeNameDateRank(fakeName string) (date string, rank int64, err error) {
	if len(fakeName) != 14 {
		return "", 0, errors.New("illegal length of fake name")
	}
	date = fakeName[:10]

	rankInt, err := strconv.Atoi(fakeName[len(fakeName)-3:])
	if err != nil {
		return "", 0, err
	}
	rank = int64(rankInt)
	return
}
