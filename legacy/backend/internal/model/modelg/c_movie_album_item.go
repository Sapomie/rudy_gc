package modelg

type CMovieAlbumItem struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	AlbumId int64 `gorm:"column:album_id;not null;index;uniqueIndex:uk_album_movie;index:idx_album_release_name,priority:1"`

	MovieJavId    string `gorm:"column:movie_jav_id;type:varchar(191);not null;index;uniqueIndex:uk_album_movie;index:idx_album_release_name,priority:4"`
	MovieName     string `gorm:"column:movie_name;type:varchar(191);not null;index;index:idx_album_release_name,priority:3"`
	ReleasingDate int64  `gorm:"column:releasing_date;not null;default:0;index:idx_album_release_name,priority:2"`
	SortNo        int64  `gorm:"column:sort_no;not null;index"`

	CreatedOn int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64 `gorm:"column:updated_on;not null;default:0"`
}

const cMovieAlbumItemTableName = "c_movie_album_item"

func (CMovieAlbumItem) TableName() string { return cMovieAlbumItemTableName }
