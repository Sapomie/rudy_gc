package modelg

type Film struct {
	ID          int64  `gorm:"primaryKey;autoIncrement;column:id"`
	MovieJavID  string `gorm:"column:movie_jav_id;type:varchar(191);not null;uniqueIndex"` // 扩展长度，防止长 jav_id 被截断
	MovieName   string `gorm:"column:movie_name;type:varchar(191);not null;uniqueIndex"`
	FileName    string `gorm:"column:file_name;type:varchar(500);not null;uniqueIndex"`
	DirectoryID int64  `gorm:"column:directory_id;not null;index"`
	RootDir     string `gorm:"type:varchar(191);not null"`

	// 新增的多层目录 ID（全部非空，默认 0）
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

	HasSub       int8  `gorm:"column:has_sub;type:tinyint;not null"`
	SelfMake     int8  `gorm:"column:self_make;type:tinyint;not null"`
	HasMask      int8  `gorm:"column:has_mask;type:tinyint;not null"`
	NeedScanMeta int8  `gorm:"column:need_scan_meta;type:tinyint;not null"`
	IsRemoved    int8  `gorm:"column:is_removed;type:tinyint;not null"`
	RemoveTime   int64 `gorm:"column:remove_time;not null"`

	ScTimes    int64 `gorm:"column:sc_times;type:mediumint;not null;index"`
	ComeTimes  int64 `gorm:"column:come_times;type:mediumint;not null;index"`
	LastScTime int64 `gorm:"column:last_sc_time;not null;index"`
	BirthTime  int64 `gorm:"column:birth_time;not null;index"`

	CreatedOn int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64 `gorm:"column:updated_on;not null;default:0"`
}

func (Film) TableName() string { return "v_film" }
