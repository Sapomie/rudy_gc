package modelg

type DeletedMovie struct {
	Id           int64  `gorm:"primary_key;autoIncrement"`
	JavId        string `gorm:"type:varchar(191);not null;uniqueIndex:uk_deleted_movie_jav_id"`
	Name         string `gorm:"type:varchar(191);not null;index:idx_deleted_movie_name"`
	DeleteSource string `gorm:"type:varchar(64);not null"`
	DeletedOn    int64  `gorm:"not null"`
	SnapshotJson string `gorm:"type:longtext;not null"`
	CreatedOn    int64  `gorm:"not null"`
	UpdatedOn    int64  `gorm:"not null"`
}

const deletedMovieTableName = "e_deleted_movie"

func (DeletedMovie) TableName() string { return deletedMovieTableName }
