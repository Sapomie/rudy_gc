package modelg

type Query struct {
	Id         int64  `gorm:"primary_key"`
	Name       string `gorm:"not null;unique"`
	Active     int64  `gorm:"not null;type:tinyint"`
	SearchType int64  `gorm:"not null;type:tinyint"`
	NameType   int64  `gorm:"not null;type:tinyint"`

	PageNow int64 `gorm:"not null;type:MEDIUMINT"`
	Offset  int64 `gorm:"not null;type:MEDIUMINT"`

	StartPage int64 `gorm:"not null;type:MEDIUMINT"`
	EndPage   int64 `gorm:"not null;type:MEDIUMINT"`

	LastQueryTime int64 `gorm:"not null"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const queryTableName = "d_query"

func (i *Query) TableName() string {
	return queryTableName
}
