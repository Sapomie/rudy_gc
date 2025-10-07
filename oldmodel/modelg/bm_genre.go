package modelg

type Genre struct {
	Id               int64  `gorm:"primary_key"`
	Name             string `gorm:"not null;unique"`
	JavId            string `gorm:"not null"`
	MovieNumber      int64  `gorm:"not null"`
	OwnedMovieNumber int64  `gorm:"not null"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const genreTableName = "bm_genre"

func (i *Genre) TableName() string {
	return genreTableName
}
