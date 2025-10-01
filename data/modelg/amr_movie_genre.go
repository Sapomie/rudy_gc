package modelg

type MovieGenre struct {
	Id        int64 `gorm:"primary_key"`
	MovieId   int64 `gorm:"not null;uniqueIndex:m_c"`
	GenreId   int64 `gorm:"not null;index:idx_cid;uniqueIndex:m_c"`
	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const movieGenreTableName = "amr_movie_genre"

func (i *MovieGenre) TableName() string {
	return movieGenreTableName
}
