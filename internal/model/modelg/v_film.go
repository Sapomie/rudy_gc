package modelg

type Film struct {
	ID         int64  `gorm:"primaryKey;autoIncrement;column:id"`
	MovieJavID string `gorm:"column:movie_jav_id;type:varchar(191);not null;uniqueIndex"`
	// 参与所有联合排序索引的末列（降序）
	MovieName   string `gorm:"column:movie_name;type:varchar(191);not null;uniqueIndex;index:idx_birth_name,priority:2,sort:desc;index:idx_sc_last_name,priority:3,sort:desc;index:idx_come_last_name,priority:3,sort:desc;index:idx_lastsc_name,priority:2,sort:desc;index:idx_reldate_name,priority:2,sort:desc"`
	FileName    string `gorm:"column:file_name;type:varchar(500);not null;uniqueIndex"`
	DirectoryID int64  `gorm:"column:directory_id;not null;index"`
	RootDir     string `gorm:"column:root_dir;type:varchar(191);not null"`
	FullDir     string `gorm:"column:full_dir;type:varchar(191);not null"`

	Dir1ID int64 `gorm:"column:dir1_id;not null;default:0;index"`
	Dir2ID int64 `gorm:"column:dir2_id;not null;default:0;index"`
	Dir3ID int64 `gorm:"column:dir3_id;not null;default:0;index"`
	Dir4ID int64 `gorm:"column:dir4_id;not null;default:0;index"`

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

	// 支持：ORDER BY sc_times DESC, last_sc_time DESC, movie_name DESC
	ScTimes int64 `gorm:"column:sc_times;type:mediumint;not null;index:idx_sc_last_name,priority:1,sort:desc"`
	// 支持：ORDER BY come_times DESC, last_sc_time DESC, movie_name DESC
	ComeTimes int64 `gorm:"column:come_times;type:mediumint;not null;index:idx_come_last_name,priority:1,sort:desc"`
	// 在两条联合索引里都作为第二列；也支持单独：ORDER BY last_sc_time DESC, movie_name DESC
	LastScTime int64 `gorm:"column:last_sc_time;not null;index:idx_sc_last_name,priority:2,sort:desc;index:idx_come_last_name,priority:2,sort:desc;index:idx_lastsc_name,priority:1,sort:desc"`
	// 支持：ORDER BY birth_time DESC, movie_name DESC
	BirthTime int64 `gorm:"column:birth_time;not null;index:idx_birth_name,priority:1,sort:desc"`
	// 支持：ORDER BY releasing_date DESC, movie_name DESC
	ReleasingDate int64 `gorm:"column:releasing_date;not null;index:idx_reldate_name,priority:1,sort:desc"`

	CreatedOn int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64 `gorm:"column:updated_on;not null;default:0"`
}

func (Film) TableName() string { return "v_film" }
