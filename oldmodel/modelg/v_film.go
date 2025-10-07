package modelg

type Film struct {
	Id         int64  `gorm:"not null"`
	MovieJavId string `gorm:"not null;unique"`
	Name       string `gorm:"not null;unique"`

	FilePath string `gorm:"not null;unique;type:varchar(500)"`

	Dir  string `gorm:"not null;type:varchar(191)"`
	Dir1 string `gorm:"not null;type:varchar(191)"`
	Dir2 string `gorm:"not null;type:varchar(191)"`
	Dir3 string `gorm:"not null;type:varchar(191)"`
	Dir4 string `gorm:"not null;type:varchar(191)"`

	Size      int64  `gorm:"not null"`
	BirthTime int64  `gorm:"not null"`
	Prefix    string `gorm:"not null;type:varchar(191)"`
	Alias     string `gorm:"not null;type:varchar(191)"`
	AlbumId   int64  `gorm:"not null"`

	Width        int64   `gorm:"not null;type:smallint"`
	Height       int64   `gorm:"not null;type:smallint"`
	BitRate      int64   `gorm:"not null;type:int"`
	Duration     int64   `gorm:"not null;type:int"`
	FrameAverage float64 `gorm:"not null"`
	RawProbe     string  `gorm:"not null"`

	HasSub   int64 `gorm:"not null;type:tinyint"`
	SelfMake int64 `gorm:"not null;type:tinyint"`
	Erased   int64 `gorm:"not null;type:tinyint"`

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

const filmTableName = "v_film"

const (
	FilmMetaDataNeedScan = 1 + iota
	FilmMetaDataNoNeedScan
)

const (
	FilmNotErased = 1 + iota
	FilmErased
	FilmNoMosaic
)

func (i *Film) Discrimination() string {
	return i.Name
}

func (i *Film) TableName() string {
	return filmTableName
}
