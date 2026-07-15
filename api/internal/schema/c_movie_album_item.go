package schema

type CMovieAlbumItem struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	AlbumId int64 `gorm:"column:album_id;not null;index;uniqueIndex:uk_album_movie_identity,priority:1;index:idx_album_movie_jav,priority:1;index:idx_album_movie_fc2,priority:1;index:idx_album_release_identity,priority:1"`

	MovieType  string `gorm:"column:movie_type;type:varchar(16);not null;default:'jav';uniqueIndex:uk_album_movie_identity,priority:2;index:idx_album_movie_jav,priority:2;index:idx_album_movie_fc2,priority:2;index:idx_album_release_identity,priority:3"`
	MovieJavId string `gorm:"column:movie_jav_id;type:varchar(191);not null;index;uniqueIndex:uk_album_movie_identity,priority:3;index:idx_album_movie_jav,priority:3;index:idx_album_release_identity,priority:4"`
	FC2Key     string `gorm:"column:fc2_key;type:varchar(64);not null;default:'';uniqueIndex:uk_album_movie_identity,priority:4;index:idx_album_movie_fc2,priority:3;index:idx_album_release_identity,priority:5"`

	MovieName     string `gorm:"column:movie_name;type:varchar(500);not null"`
	ReleasingDate int64  `gorm:"column:releasing_date;not null;default:0;index:idx_album_release_identity,priority:2"`
	SortNo        int64  `gorm:"column:sort_no;not null;index"`

	CreatedOn int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64 `gorm:"column:updated_on;not null;default:0"`
}

const cMovieAlbumItemTableName = "c_movie_album_item"

func (CMovieAlbumItem) TableName() string { return cMovieAlbumItemTableName }
