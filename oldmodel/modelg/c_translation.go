package modelg

type Translation struct {
	Id         int64  `gorm:"not null"`
	MovieJavId string `gorm:"not null;unique"`
	Title      string `gorm:"not null;type:varchar(300)"`
	Chinese    string `gorm:"not null;type:varchar(300)"`
	HasChinese int64  `gorm:"not null;type:tinyint"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const translationTableName = "c_translation"

func (i *Translation) Discrimination() string {
	return i.MovieJavId
}

func (i *Translation) TableName() string {
	return translationTableName
}
