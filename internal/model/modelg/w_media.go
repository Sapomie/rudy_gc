package modelg

type Media struct {
	ID                int64  `gorm:"primaryKey;autoIncrement;column:id"`
	MovieJavID        string `gorm:"column:movie_jav_id;type:varchar(191);not null;uniqueIndex"`
	MovieName         string `gorm:"column:movie_name;type:varchar(191);not null;uniqueIndex;index:idx_w_media_birth_name,priority:2,sort:desc;index:idx_w_media_reldate_name,priority:2,sort:desc"`
	FileName          string `gorm:"column:file_name;type:varchar(500);not null;uniqueIndex"`
	SourceTorrentHash string `gorm:"column:source_torrent_hash;type:char(40);not null;index:idx_w_media_source_torrent_hash"`

	DirectoryID int64  `gorm:"column:directory_id;not null;index"`
	RootDir     string `gorm:"column:root_dir;type:varchar(191);not null"`
	FullDir     string `gorm:"column:full_dir;type:varchar(191);not null"`

	Alias        string  `gorm:"column:alias;type:varchar(191);not null"`
	Size         int64   `gorm:"column:size;not null"`
	Width        int64   `gorm:"column:width;type:smallint;not null"`
	Height       int64   `gorm:"column:height;type:smallint;not null"`
	BitRate      int64   `gorm:"column:bit_rate;type:int;not null"`
	Duration     int64   `gorm:"column:duration;type:int;not null"`
	FrameAverage float64 `gorm:"column:frame_average;type:double;not null"`
	HasSub       int8    `gorm:"column:has_sub;type:tinyint;not null"`
	SelfMake     int8    `gorm:"column:self_make;type:tinyint;not null"`
	HasMask      int8    `gorm:"column:has_mask;type:tinyint;not null"`
	NeedScanMeta int8    `gorm:"column:need_scan_meta;type:tinyint;not null"`
	IsRemoved    int8    `gorm:"column:is_removed;type:tinyint;not null"`
	RemoveTime   int64   `gorm:"column:remove_time;not null"`

	BirthTime     int64 `gorm:"column:birth_time;not null;index:idx_w_media_birth_name,priority:1,sort:desc"`
	ReleasingDate int64 `gorm:"column:releasing_date;not null;index:idx_w_media_reldate_name,priority:1,sort:desc"`
	CreatedOn     int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn     int64 `gorm:"column:updated_on;not null;default:0"`
}

func (Media) TableName() string { return "w_media" }
