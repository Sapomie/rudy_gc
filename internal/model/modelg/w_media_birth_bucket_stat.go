package modelg

type MediaBirthBucketStat struct {
	ID              int64  `gorm:"primaryKey;autoIncrement;column:id"`
	ScopeKey        string `gorm:"column:scope_key;type:varchar(32);not null;uniqueIndex:uk_wmbbs_scope_key"`
	Level           string `gorm:"column:level;type:varchar(16);not null;index:idx_wmbbs_level_sort,priority:1"`
	Year            int64  `gorm:"column:year;type:int;not null;index:idx_wmbbs_level_sort,priority:2,sort:desc"`
	Quarter         int64  `gorm:"column:quarter;type:tinyint;not null;index:idx_wmbbs_level_sort,priority:3,sort:desc"`
	Month           int64  `gorm:"column:month;type:tinyint;not null;index:idx_wmbbs_level_sort,priority:4,sort:desc"`
	Day             int64  `gorm:"column:day;type:tinyint;not null;index:idx_wmbbs_level_sort,priority:5,sort:desc"`
	MediaCount      int64  `gorm:"column:media_count;not null"`
	RemovedCount    int64  `gorm:"column:removed_count;not null"`
	SizeBytes       int64  `gorm:"column:size_bytes;not null"`
	HasSubCount     int64  `gorm:"column:has_sub_count;not null"`
	LatestBirthTime int64  `gorm:"column:latest_birth_time;not null;index:idx_wmbbs_latest_birth_time,sort:desc"`
	CreatedOn       int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn       int64  `gorm:"column:updated_on;not null;default:0"`
}

func (MediaBirthBucketStat) TableName() string { return "w_media_birth_bucket_stat" }
