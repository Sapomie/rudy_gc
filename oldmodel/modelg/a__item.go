package modelg

type Item struct {
	Id     int64  `gorm:"primary_key"`
	Name   string `gorm:"not null;index"`
	JavId  string `gorm:"not null;unique"`
	Prefix string `gorm:"not null;type:varchar(191)"`

	HasDetail  int64  `gorm:"not null;type:tinyint"`
	HasBus     int64  `gorm:"not null;type:tinyint"`
	SearchType int64  `gorm:"not null;type:tinyint"`
	CoverUrl   string `gorm:"not null;type:varchar(300)"`

	SearchBy string `gorm:"not null;type:varchar(191)"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const itemTableName = "a__item"

func (i *Item) TableName() string {
	return itemTableName
}
