package modelg

type Rank struct {
	Id         int64  `gorm:"not null"`
	Name       string `gorm:"not null;unique"`
	MovieJavId string `gorm:"not null;index:idx_movie_number_day"`
	DayNumber  int64  `gorm:"not null;index;type:smallint;index:idx_movie_number_day"`
	Number     int64  `gorm:"not null;type:smallint;index:idx_movie_number_day"`
	Category   int64  `gorm:"not null;type:tinyint"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const RankTableName = "c_rank"

func (i *Rank) TableName() string {
	return RankTableName
}
