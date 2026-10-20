package modelg

type ShMagnet struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	ThreadTID     string `gorm:"column:thread_tid;type:varchar(64);not null;default:'';index"`
	MovieJavID    string `gorm:"column:movie_jav_id;type:varchar(191);not null;default:'';index"`
	MovieName     string `gorm:"column:movie_name;type:varchar(191);not null;default:'';index;index:idx_sh_magnet_type_movie_name,priority:2"`
	MovieType     string `gorm:"column:movie_type;type:varchar(16);not null;default:'';index;index:idx_sh_magnet_type_movie_name,priority:1"`
	SourceForumID string `gorm:"column:source_forum_id;type:varchar(32);not null;default:'';index"`
	PostTime      int64  `gorm:"column:post_time;not null;default:0;index"`
	MagnetURL     string `gorm:"column:magnet_url;type:text;not null"`
	InfoHash      string `gorm:"column:info_hash;type:char(40);not null;uniqueIndex:uk_info_hash"`
	DownloadState int8   `gorm:"column:download_state;type:tinyint;not null;default:0;index:idx_sh_magnet_download_state"`
	CreatedOn     int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn     int64  `gorm:"column:updated_on;not null;default:0"`
}

func (ShMagnet) TableName() string { return "sh_magnet" }
