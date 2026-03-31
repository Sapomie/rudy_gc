package modelg

type SehuatangMagnet struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	MovieJavID string `gorm:"column:movie_jav_id;type:varchar(191);not null;index"`
	MovieName  string `gorm:"column:movie_name;type:varchar(191);not null;index"`

	ThreadTitle string `gorm:"column:thread_title;type:varchar(500);not null"`
	ThreadURL   string `gorm:"column:thread_url;type:varchar(255);not null;index"`
	PostTime    int64  `gorm:"column:post_time;not null;index"`
	PostDate    int64  `gorm:"column:post_date;not null;index"`

	InfoHash string `gorm:"column:info_hash;type:char(40);not null;index;uniqueIndex:uk_info_hash"`

	LastSeenTime int64 `gorm:"column:last_seen_time;not null;index"`
	CreatedOn    int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn    int64 `gorm:"column:updated_on;not null;default:0"`
}

const sehuatangMagnetTableName = "t_sehuatang_magnet"

func (SehuatangMagnet) TableName() string { return sehuatangMagnetTableName }
