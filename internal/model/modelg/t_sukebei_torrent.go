package modelg

type SukebeiTorrent struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	MovieJavID string `gorm:"column:movie_jav_id;type:varchar(191);not null;index"`
	QueryText  string `gorm:"column:query_text;type:varchar(191);not null"`
	SearchURL  string `gorm:"column:search_url;type:varchar(255);not null"`

	TorrentTitle string `gorm:"column:torrent_title;type:varchar(500);not null"`
	ViewID       int64  `gorm:"column:view_id;not null;uniqueIndex"`
	ViewURL      string `gorm:"column:view_url;type:varchar(255);not null"`
	TorrentURL   string `gorm:"column:torrent_url;type:varchar(255);not null"`
	MagnetURL    string `gorm:"column:magnet_url;type:text;not null"`
	InfoHash     string `gorm:"column:info_hash;type:char(40);not null;index"`
	Dn           string `gorm:"column:dn;type:varchar(500);not null"`

	SizeBytes int64  `gorm:"column:size_bytes;not null"`
	SizeText  string `gorm:"column:size_text;type:varchar(64);not null"`

	PublishTime int64 `gorm:"column:publish_time;not null;index"`
	Seeders     int64 `gorm:"column:seeders;type:mediumint;not null"`
	Leechers    int64 `gorm:"column:leechers;type:mediumint;not null"`
	Completed   int64 `gorm:"column:completed;type:int;not null"`

	LastSeenTime int64 `gorm:"column:last_seen_time;not null;index"`
	CreatedOn    int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn    int64 `gorm:"column:updated_on;not null;default:0"`
}

const sukebeiTorrentTableName = "t_sukebei_torrent"

func (SukebeiTorrent) TableName() string { return sukebeiTorrentTableName }
