package modelg

type MovieCast struct {
	Id        int64 `gorm:"primary_key"`
	MovieId   int64 `gorm:"not null;uniqueIndex:m_c"`
	CastId    int64 `gorm:"not null;index:idx_cid;uniqueIndex:m_c"`
	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const movieCastTableName = "amr_movie_cast"

func (i *MovieCast) TableName() string {
	return movieCastTableName
}
