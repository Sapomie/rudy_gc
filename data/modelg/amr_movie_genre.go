package modelg

type MovieGenre struct {
	ID         int64  `gorm:"primaryKey;autoIncrement;column:id"`
	MovieJavID string `gorm:"column:movie_jav_id;type:varchar(191);not null;index:idx_mjav;uniqueIndex:m_g"`
	GenreID    int64  `gorm:"column:genre_id;not null;index:idx_gid;uniqueIndex:m_g"`
	CreatedOn  int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn  int64  `gorm:"column:updated_on;not null;default:0"`
}

const movieGenreTableName = "amr_movie_genre"

func (MovieGenre) TableName() string { return movieGenreTableName }
