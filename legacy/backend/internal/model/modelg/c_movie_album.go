package modelg

type CMovieAlbum struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	Name   string `gorm:"column:name;type:varchar(128);not null;uniqueIndex"`
	Remark string `gorm:"column:remark;type:varchar(512);not null"`

	CreatedOn int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64 `gorm:"column:updated_on;not null;default:0"`
}

const cMovieAlbumTableName = "c_movie_album"

func (CMovieAlbum) TableName() string { return cMovieAlbumTableName }
