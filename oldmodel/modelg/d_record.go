package modelg

type Record struct {
	Id           int64  `gorm:"not null"`
	Name         string `gorm:"not null;unique"`
	StartTime    int64  `gorm:"not null"`
	EndTime      int64  `gorm:"not null"`
	Type         string `gorm:"not null;type:varchar(191)"`
	DetailNumber int64  `gorm:"not null"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const recordTableName = "d_record"

const (
	RecordTypeAll = 1 + iota
	RecordTypeBest
	RecordTypeRefresh
	RecordTypeRefBusMag
)

const (
	RecordQueryByTimer = 1 + iota
	RecordQueryByManual
)

const (
	RecordSuccess = 1 + iota
	RecordUnSuccess
)

func (i *Record) TableName() string {
	return recordTableName
}
