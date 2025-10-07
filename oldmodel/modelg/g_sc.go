package modelg

type Sc struct {
	Id            int64  `gorm:"not null"`
	Name          string `gorm:"not null;unique"`
	MovieNumber   int64  `gorm:"not null;type:smallint"`
	ScTime        int64  `gorm:"not null"`
	ComeMovieName string `gorm:"not null;type:varchar(191)"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const ScTableName = "g_sc"

func (i *Sc) TableName() string {
	return ScTableName
}
