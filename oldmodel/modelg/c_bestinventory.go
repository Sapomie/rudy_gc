package modelg

import (
	"time"

	"gorm.io/gorm"
)

const bestInvTableName = "bestinv"

const (
	BestCategoryMonth = 1 + iota
	BestCategoryAllTime
)

const (
	NeedRankCheck = 1 + iota
	NoNeedRankCheck
)

type BestInv struct {
	Id            int64  `gorm:"primary_key"`
	Name          string `gorm:"unique;not null"`
	Status        int8   `gorm:"not null"` //是否存入 item check
	RankCheck     int8   `gorm:"not null"` //是否扫描入 rank 表 check
	Page          int64  `gorm:"not null"`
	Date          string `gorm:"not null"`
	DayNumber     int64  `gorm:"not null"`
	Content       string `gorm:"not null"`
	Category      int64  `gorm:"not null"`
	LastQueryTime int64  `gorm:"not null"`
	CreatedOn     int64  `gorm:"not null"`
	UpdatedOn     int64  `gorm:"not null"`
}

func (i *BestInv) Discrimination() string {
	return i.Name
}

func (i *BestInv) TableName() string {
	return bestInvTableName
}

func (i *BestInv) BeforeCreate(tx *gorm.DB) (err error) {
	i.CreatedOn = time.Now().Unix()
	return
}

func (i *BestInv) BeforeUpdate(tx *gorm.DB) (err error) {
	i.UpdatedOn = time.Now().Unix()
	return
}
