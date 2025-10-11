package modelg

type Record struct {
	Id           int64  `gorm:"primary_key"`
	Name         string `gorm:"not null;unique"`
	StartTime    int64  `gorm:"not null"`
	EndTime      int64  `gorm:"not null"`
	Type         string `gorm:"not null;type:varchar(191)"`
	DetailNumber int64  `gorm:"not null"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const recordTableName = "e_record"

func (i *Record) TableName() string {
	return recordTableName
}
