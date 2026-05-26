package modelg

type MovieCast struct {
	ID         int64  `gorm:"primaryKey;autoIncrement;column:id"`
	MovieJavID string `gorm:"column:movie_jav_id;type:varchar(191);not null;index:idx_mjav;uniqueIndex:m_c"`
	CastID     int64  `gorm:"column:cast_id;not null;index:idx_cid;uniqueIndex:m_c"`
	CreatedOn  int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn  int64  `gorm:"column:updated_on;not null;default:0"`
}

const movieCastTableName = "amr_movie_cast"

func (MovieCast) TableName() string { return movieCastTableName }
