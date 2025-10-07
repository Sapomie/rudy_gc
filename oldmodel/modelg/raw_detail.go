package modelg

type Detail struct {
	Id            int64  `gorm:"primary_key"`
	Name          string `gorm:"not null;type:varchar(191)"`
	JavId         string `gorm:"not null;unique"`
	Prefix        string `gorm:"not null;type:varchar(191)"`
	LastQueryTime int64  `gorm:"not null"`
	QueryUrl      string `gorm:"not null;type:varchar(300)"`
	NeedScan      int64  `gorm:"not null;type:tinyint;index"`
	Content       string `gorm:"not null"`
	CreatedOn     int64  `gorm:"not null"`
	UpdatedOn     int64  `gorm:"not null"`
}

const detailTableName = "raw_detail"

func (i *Detail) TableName() string {
	return detailTableName
}
