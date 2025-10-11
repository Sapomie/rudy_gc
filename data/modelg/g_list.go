package modelg

type List struct {
	Id         int64  `gorm:"primary_key"`
	Name       string `gorm:"not null;unique"`
	ScName     string `gorm:"not null;type:varchar(191)"`
	MovieJavId string `gorm:"not null;type:varchar(191)"`
	IsCome     int64  `gorm:"not null;type:tinyint"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

const ListTableName = "g_list"

func (i *List) TableName() string {
	return ListTableName
}
