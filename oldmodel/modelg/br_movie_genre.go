package modelg

type MovieGenre struct {
	Id        int64 `gorm:"primaryKey;autoIncrement"`
	MovieId   int64 `gorm:"not null;index:idx_mid;uniqueIndex:m_c"`
	GenreId   int64 `gorm:"not null;index:idx_gid;uniqueIndex:m_c"`
	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const movieGenreTableName = "amr_movie_genre"

func (MovieGenre) TableName() string {
	return movieGenreTableName
}
