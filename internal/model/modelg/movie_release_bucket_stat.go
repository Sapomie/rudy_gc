package modelg

type MovieReleaseBucketStat struct {
	ID                  int64  `gorm:"primaryKey;autoIncrement;column:id"`
	ScopeKey            string `gorm:"column:scope_key;type:varchar(32);not null;uniqueIndex:uk_mrbs_scope_key"`
	Level               string `gorm:"column:level;type:varchar(16);not null;index:idx_mrbs_level_sort,priority:1"`
	Year                int64  `gorm:"column:year;type:int;not null;index:idx_mrbs_level_sort,priority:2,sort:desc"`
	Quarter             int64  `gorm:"column:quarter;type:tinyint;not null;index:idx_mrbs_level_sort,priority:3,sort:desc"`
	Month               int64  `gorm:"column:month;type:tinyint;not null;index:idx_mrbs_level_sort,priority:4,sort:desc"`
	Day                 int64  `gorm:"column:day;type:tinyint;not null;index:idx_mrbs_level_sort,priority:5,sort:desc"`
	CountAll            int64  `gorm:"column:count_all;not null"`
	CountOwned          int64  `gorm:"column:count_owned;not null"`
	SizeBytes           int64  `gorm:"column:size_bytes;not null"`
	LatestReleasingDate int64  `gorm:"column:latest_releasing_date;not null;index:idx_mrbs_latest_release,sort:desc"`
	CreatedOn           int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn           int64  `gorm:"column:updated_on;not null;default:0"`
}

func (MovieReleaseBucketStat) TableName() string { return "movie_release_bucket_stat" }
