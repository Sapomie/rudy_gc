package modelg

import (
	"context"

	"gorm.io/gorm"
)

type BestInvModel struct {
	db *gorm.DB
}

func NewBestInvModel(dbEngine *gorm.DB) *BestInvModel {
	return &BestInvModel{db: dbEngine.Table(bestInvTableName)}
}

func (i *BestInvModel) ExistByDiscrimination(ctx context.Context, name string) (bool, error) {
	var count int64
	err := i.db.WithContext(ctx).Where("`name` = ?", name).Count(&count).Error
	if err != nil {
		return false, err
	}
	exist := count != 0
	return exist, nil
}

func (i *BestInvModel) FindDataByDiscrimination(ctx context.Context, name string) (DataStruct, error) {
	return i.FindByName(ctx, name)
}

func (i *BestInvModel) InsertData(ctx context.Context, data DataStruct) error {
	return i.Insert(ctx, data.(*BestInv))
}

func (i *BestInvModel) UpdateDataByDiscrimination(ctx context.Context, data DataStruct) error {
	return i.UpdateByName(ctx, data.(*BestInv))
}

func (i *BestInvModel) FindByName(ctx context.Context, name string) (*BestInv, error) {
	result := new(BestInv)
	err := i.db.WithContext(ctx).Where("`name` = ?", name).First(result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (i *BestInvModel) Insert(ctx context.Context, data *BestInv) error {
	err := i.db.WithContext(ctx).Create(data).Error
	if err != nil {
		return err
	}
	return err
}

func (i *BestInvModel) UpdateByName(ctx context.Context, data *BestInv) error {
	err := i.db.WithContext(ctx).Where("`name` = ?", data.Name).Updates(data).Error
	if err != nil {
		return err
	}
	return err
}

//func (i *BestInvModel) NeedScan(ctx context.Context) ([]*BestInv, error) {
//	result := make([]*BestInv, 0)
//	err := i.db.WithContext(ctx).Where("`status` = ?", InventoryStatusNeedScan).Find(&result).Error
//	if err != nil {
//		return nil, err
//	}
//	return result, nil
//}

func (i *BestInvModel) NeedRankCheck(ctx context.Context, limit int) ([]*BestInv, error) {
	result := make([]*BestInv, 0)
	err := i.db.WithContext(ctx).Where("`rank_check` != ?", NoNeedRankCheck).Limit(limit).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (i *BestInvModel) FindByDayNumber(ctx context.Context, dauNumber int64) ([]*BestInv, error) {
	result := make([]*BestInv, 0)
	err := i.db.WithContext(ctx).Where("`day_number` = ?", dauNumber).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (i *BestInvModel) FindByDayNumberRankMonth(ctx context.Context, dauNumber int64) ([]*BestInv, error) {
	result := make([]*BestInv, 0)
	err := i.db.WithContext(ctx).
		Where("`day_number` = ?", dauNumber).
		Where("`category` = ?", 1).
		Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (i *BestInvModel) LatestDayNumber(ctx context.Context) (int64, error) {
	var latestDayNumber int64
	err := i.db.WithContext(ctx).Select("day_number").Order("day_number desc").Limit(1).Take(&latestDayNumber).Error
	if err != nil {
		return 0, err
	}

	return latestDayNumber, nil
}

func (i *BestInvModel) Count(ctx context.Context) (int64, error) {
	var count int64
	err := i.db.WithContext(ctx).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (i *BestInvModel) DayBestInvNumb(ctx context.Context, date string) (int64, error) {
	var result int64
	err := i.db.WithContext(ctx).Where("`date` = ?", date).Count(&result).Error
	if err != nil {
		return 0, err
	}
	return result, nil
}
