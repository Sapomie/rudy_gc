package modelg

type Bestinv struct {
	Id            int64  `gorm:"primary_key"`
	Name          string `gorm:"unique;not null;unique"`
	NeedScan      int64  `gorm:"not null;type:tinyint;index"`
	NeedRankCheck int64  `gorm:"not null;type:tinyint;index"`
	Category      int64  `gorm:"not null;type:tinyint"`
	Page          int64  `gorm:"not null;type:smallint"`
	DayNumber     int64  `gorm:"not null;type:smallint"`
	Content       string `gorm:"not null"`
	LastQueryTime int64  `gorm:"not null"`
	Date          string `gorm:"not null;type:varchar(191)"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const bestinvTableName = "raw_bestinv"

func (i *Bestinv) TableName() string {
	return bestinvTableName
}
