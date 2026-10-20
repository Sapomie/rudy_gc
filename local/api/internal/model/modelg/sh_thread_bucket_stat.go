package modelg

type ShThreadBucketStat struct {
	ID                int64  `gorm:"primaryKey;autoIncrement;column:id"`
	AggScope          string `gorm:"column:agg_scope;type:varchar(64);not null;uniqueIndex:uk_shbs_scope_bucket,priority:1;index:idx_shbs_level_sort,priority:1;index:idx_shbs_latest_post,priority:1"`
	ScopeKey          string `gorm:"column:scope_key;type:varchar(32);not null;uniqueIndex:uk_shbs_scope_bucket,priority:2"`
	Level             string `gorm:"column:level;type:varchar(16);not null;index:idx_shbs_level_sort,priority:2"`
	Year              int64  `gorm:"column:year;type:int;not null;index:idx_shbs_level_sort,priority:3,sort:desc"`
	Quarter           int64  `gorm:"column:quarter;type:tinyint;not null;index:idx_shbs_level_sort,priority:4,sort:desc"`
	Month             int64  `gorm:"column:month;type:tinyint;not null;index:idx_shbs_level_sort,priority:5,sort:desc"`
	Week              int64  `gorm:"column:week;type:tinyint;not null;index:idx_shbs_level_sort,priority:6,sort:desc"`
	Day               int64  `gorm:"column:day;type:tinyint;not null;index:idx_shbs_level_sort,priority:7,sort:desc"`
	BucketStart       int64  `gorm:"column:bucket_start;not null"`
	BucketEnd         int64  `gorm:"column:bucket_end;not null"`
	CountAll          int64  `gorm:"column:count_all;not null"`
	CountPending      int64  `gorm:"column:count_pending;not null"`
	CountFetched      int64  `gorm:"column:count_fetched;not null"`
	CountFailed       int64  `gorm:"column:count_failed;not null"`
	CountWithMagnet   int64  `gorm:"column:count_with_magnet;not null"`
	CountWithInfoHash int64  `gorm:"column:count_with_info_hash;not null"`
	LatestPostTime    int64  `gorm:"column:latest_post_time;not null;index:idx_shbs_latest_post,priority:2,sort:desc"`
	CreatedOn         int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn         int64  `gorm:"column:updated_on;not null;default:0"`
}

func (ShThreadBucketStat) TableName() string { return "sh_bucket_stat" }
