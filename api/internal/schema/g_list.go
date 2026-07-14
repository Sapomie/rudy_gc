package schema

type List struct {
	Id         int64  `gorm:"primary_key"`
	Name       string `gorm:"not null;unique"`
	ScName     string `gorm:"not null;type:varchar(191)"`
	MovieType  string `gorm:"column:movie_type;not null;type:varchar(16);default:'jav'"`
	MovieJavId string `gorm:"not null;type:varchar(191)"`
	FC2MovieID int64  `gorm:"column:fc2_movie_id;not null;default:0"`
	IsCome     int64  `gorm:"not null;type:tinyint"`
	IsSc       int64  `gorm:"not null;type:tinyint"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

const ListTableName = "g_list"

func (i *List) TableName() string {
	return ListTableName
}
