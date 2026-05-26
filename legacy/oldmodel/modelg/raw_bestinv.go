package modelg

type Inventory struct {
	Id            int64  `gorm:"primary_key"`
	Name          string `gorm:"unique;not null;unique"`
	NeedScan      int64  `gorm:"not null;type:tinyint;index"`
	Keyword       string `gorm:"not null;type:varchar(191)"`
	Parent        string `gorm:"not null;type:varchar(191)"`
	Page          int64  `gorm:"not null;type:smallint"`
	Content       string `gorm:"not null"`
	Category      int64  `gorm:"not null;type:tinyint"`
	LastQueryTime int64  `gorm:"not null"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const inventoryTableName = "raw_inventory"

func (i *Inventory) TableName() string {
	return inventoryTableName
}
