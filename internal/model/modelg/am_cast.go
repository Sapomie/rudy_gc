package modelg

type Cast struct {
	Id       int64  `gorm:"primary_key"`
	PersonId int64  `gorm:"not null;index"`
	Name     string `gorm:"not null;unique"`
	JavId    string `gorm:"not null;index"`

	MovieNumber       int64 `gorm:"not null;index"`
	OwnedMovieNumber  int64 `gorm:"not null;index"`
	OwnedWMediaNumber int64 `gorm:"not null;index"`

	ScTimes         int64 `gorm:"not null;type:MEDIUMINT;index"`
	ComeTimes       int64 `gorm:"not null;type:MEDIUMINT;index"`
	LastScTime      int64 `gorm:"not null;index"`
	LastScEventTime int64 `gorm:"not null;index"`

	Rank500MovieNumber int64 `gorm:"not null;type:MEDIUMINT;index"`
	Rank20MovieNumber  int64 `gorm:"not null;type:MEDIUMINT;index"`
	Rank1MovieNumber   int64 `gorm:"not null;type:MEDIUMINT;index"`
	HighestRank        int64 `gorm:"not null;type:MEDIUMINT;index"`
	RankTimes          int64 `gorm:"not null;type:MEDIUMINT;index"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

const castTableName = "am_cast"

func (i *Cast) TableName() string {
	return castTableName
}
