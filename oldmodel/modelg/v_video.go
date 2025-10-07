package modelg

type Video struct {
	Id       int64  `gorm:"not null"`
	FilePath string `gorm:"not null;unique;type:varchar(500)"`
	BaseName string `gorm:"not null;unique"`

	DirRoot   string `gorm:"not null;type:varchar(191)"`
	Dir1      string `gorm:"not null;type:varchar(191)"`
	Dir2      string `gorm:"not null;type:varchar(191)"`
	Dir3      string `gorm:"not null;type:varchar(191)"`
	Size      int64  `gorm:"not null"`
	BirthTime int64  `gorm:"not null"`
	Label     string `gorm:"not null;type:varchar(191)"`
	Code      string `gorm:"not null;type:varchar(191)"`
	Alias     string `gorm:"not null;type:varchar(191)"`

	Width        int64   `gorm:"not null;type:smallint"`
	Height       int64   `gorm:"not null;type:smallint"`
	BitRate      int64   `gorm:"not null;type:int"`
	Duration     int64   `gorm:"not null;type:int"`
	FrameAverage float64 `gorm:"not null"`
	RawProbe     string  `gorm:"not null"`

	HasSub   int64 `gorm:"not null;type:tinyint"`
	SelfMake int64 `gorm:"not null;type:tinyint"`

	NeedScanBase int64 `gorm:"not null;type:tinyint"`
	NeedScanMeta int64 `gorm:"not null;type:tinyint"`

	IsRemoved  int64 `gorm:"not null;type:tinyint"`
	RemoveTime int64 `gorm:"not null"`
	ScTimes    int64 `gorm:"not null;type:MEDIUMINT;index"`
	ComeTimes  int64 `gorm:"not null;type:MEDIUMINT;index"`
	LastScTime int64 `gorm:"not null;index"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const videoTableName = "v_video"

func (i *Video) TableName() string {
	return videoTableName
}
