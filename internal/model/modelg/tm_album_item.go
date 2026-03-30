package modelg

type AlbumItem struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	AlbumId int64 `gorm:"column:album_id;not null;index;uniqueIndex:uk_album_source"`

	SourceType  string `gorm:"column:source_type;type:varchar(32);not null;uniqueIndex:uk_album_source"`
	SourceRowId int64  `gorm:"column:source_row_id;not null;uniqueIndex:uk_album_source"`

	MovieJavId  string `gorm:"column:movie_jav_id;type:varchar(191);not null;index"`
	MovieName   string `gorm:"column:movie_name;type:varchar(191);not null;index"`
	InfoHash    string `gorm:"column:info_hash;type:varchar(128);not null;index"`
	Size        string `gorm:"column:size;type:varchar(64);not null"`
	PublishTime int64  `gorm:"column:publish_time;not null;index"`

	CreatedOn int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64 `gorm:"column:updated_on;not null;default:0"`
}

const albumItemTableName = "tm_album_item"

func (AlbumItem) TableName() string { return albumItemTableName }
