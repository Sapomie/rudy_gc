package modelg

type Cast struct {
	Id               int64  `gorm:"primary_key"`
	Name             string `gorm:"not null;unique"`
	JavId            string `gorm:"not null"`
	MovieNumber      int64  `gorm:"not null"`
	OwnedMovieNumber int64  `gorm:"not null"`

	ScTimes    int64 `gorm:"not null;type:MEDIUMINT;index"`
	ComeTimes  int64 `gorm:"not null;type:MEDIUMINT;index"`
	LastScTime int64 `gorm:"not null;index"`

	Rank500MovieNumber int64 `gorm:"not null;type:MEDIUMINT;index"`
	Rank20MovieNumber  int64 `gorm:"not null;type:MEDIUMINT;index"`
	Rank1MovieNumber   int64 `gorm:"not null;type:MEDIUMINT;index"`
	HighestRank        int64 `gorm:"not null;type:MEDIUMINT;index"`
	RankTimes          int64 `gorm:"not null;type:MEDIUMINT;index"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const castTableName = "bm_cast"

func (i *Cast) TableName() string {
	return castTableName
}
