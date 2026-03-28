package modelg

type JavbusMagnet struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	MovieJavID string `gorm:"column:movie_jav_id;type:varchar(191);not null;uniqueIndex:uk_movie_hash,priority:1;index"`
	PageURL    string `gorm:"column:page_url;type:varchar(255);not null"`

	MagnetName string `gorm:"column:magnet_name;type:varchar(500);not null"`
	MagnetURL  string `gorm:"column:magnet_url;type:text;not null"`
	InfoHash   string `gorm:"column:info_hash;type:char(40);not null;uniqueIndex:uk_movie_hash,priority:2;index"`
	Dn         string `gorm:"column:dn;type:varchar(500);not null"`

	SizeBytes int64  `gorm:"column:size_bytes;not null"`
	SizeText  string `gorm:"column:size_text;type:varchar(64);not null"`
	ShareDate int64  `gorm:"column:share_date;not null;index"`

	HasHD       int8  `gorm:"column:has_hd;type:tinyint;not null"`
	HasSubtitle int8  `gorm:"column:has_subtitle;type:tinyint;not null"`
	RowSort     int64 `gorm:"column:row_sort;type:smallint;not null"`

	LastSeenTime int64 `gorm:"column:last_seen_time;not null;index"`
	CreatedOn    int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn    int64 `gorm:"column:updated_on;not null;default:0"`
}

const javbusMagnetTableName = "t_javbus_magnet"

func (JavbusMagnet) TableName() string { return javbusMagnetTableName }
