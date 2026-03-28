package modelg

type JavbusMagnetFetch struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	MovieJavID  string `gorm:"column:movie_jav_id;type:varchar(191);not null;uniqueIndex"`
	MovieCode   string `gorm:"column:movie_code;type:varchar(191);not null"`
	ReleaseDate int64  `gorm:"column:release_date;not null"`

	FetchStatus int8  `gorm:"column:fetch_status;type:tinyint;not null;index"`
	TryCount    int64 `gorm:"column:try_count;type:int;not null"`

	LastFetchTime     int64  `gorm:"column:last_fetch_time;not null;index"`
	LastSuccessTime   int64  `gorm:"column:last_success_time;not null"`
	LastError         string `gorm:"column:last_error;type:varchar(500);not null"`
	LastResultCount   int64  `gorm:"column:last_result_count;type:int;not null"`
	TorrentHashCount  int64  `gorm:"column:torrent_hash_count;type:int;not null"`
	LatestPublishTime int64  `gorm:"column:latest_publish_time;not null"`

	SourceURL string `gorm:"column:source_url;type:varchar(255);not null"`

	CreatedOn int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64 `gorm:"column:updated_on;not null;default:0"`
}

const javbusMagnetFetchTableName = "t_javbus_magnet_fetch"

func (JavbusMagnetFetch) TableName() string { return javbusMagnetFetchTableName }
