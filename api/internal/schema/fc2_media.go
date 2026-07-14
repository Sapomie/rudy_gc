package schema

type FC2Media struct {
	ID                int64   `gorm:"column:id;primaryKey;autoIncrement"`
	FC2MovieID        int64   `gorm:"column:fc2_movie_id;not null;uniqueIndex:uk_fc2_media_movie_id"`
	FC2Key            string  `gorm:"column:fc2_key;type:varchar(64);not null;default:''"`
	FC2Code           string  `gorm:"column:fc2_code;type:varchar(32);not null;index:idx_fc2_media_code;index:idx_fc2_media_birth_code,priority:2,sort:desc"`
	SourceThreadID    int64   `gorm:"column:source_thread_id;not null;default:0;index:idx_fc2_media_source_thread_id"`
	SourceTorrentHash string  `gorm:"column:source_torrent_hash;type:char(40);not null;default:'';index:idx_fc2_media_source_torrent_hash"`
	FileName          string  `gorm:"column:file_name;type:varchar(500);not null;uniqueIndex:uk_fc2_media_file_name"`
	DirectoryID       int64   `gorm:"column:directory_id;not null;index:idx_fc2_media_directory_id"`
	RootDir           string  `gorm:"column:root_dir;type:varchar(191);not null"`
	FullDir           string  `gorm:"column:full_dir;type:varchar(191);not null"`
	Alias             string  `gorm:"column:alias;type:varchar(191);not null;default:''"`
	Size              int64   `gorm:"column:size;not null;default:0"`
	Width             int64   `gorm:"column:width;type:smallint;not null;default:0"`
	Height            int64   `gorm:"column:height;type:smallint;not null;default:0"`
	BitRate           int64   `gorm:"column:bit_rate;type:bigint;not null;default:0"`
	Duration          int64   `gorm:"column:duration;type:bigint;not null;default:0"`
	FrameAverage      float64 `gorm:"column:frame_average;type:double;not null;default:0"`
	HasSub            int8    `gorm:"column:has_sub;type:tinyint;not null"`
	SelfMake          int8    `gorm:"column:self_make;type:tinyint;not null"`
	HasMask           int8    `gorm:"column:has_mask;type:tinyint;not null"`
	NeedScanMeta      int8    `gorm:"column:need_scan_meta;type:tinyint;not null"`
	IsRemoved         int8    `gorm:"column:is_removed;type:tinyint;not null;index:idx_fc2_media_is_removed"`
	RemoveTime        int64   `gorm:"column:remove_time;not null;default:0"`
	BirthTime         int64   `gorm:"column:birth_time;not null;default:0;index:idx_fc2_media_birth_code,priority:1,sort:desc"`
	CreatedOn         int64   `gorm:"column:created_on;not null;default:0"`
	UpdatedOn         int64   `gorm:"column:updated_on;not null;default:0"`
}

func (FC2Media) TableName() string { return "fc2_media" }
