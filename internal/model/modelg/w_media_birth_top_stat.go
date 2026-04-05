package modelg

type MediaBirthTopStat struct {
	ID         int64  `gorm:"primaryKey;autoIncrement;column:id"`
	ScopeKey   string `gorm:"column:scope_key;type:varchar(32);not null;index:idx_wmbts_scope_type_rank,priority:1;uniqueIndex:uk_wmbts_scope_type_key,priority:1"`
	Level      string `gorm:"column:level;type:varchar(16);not null;index:idx_wmbts_level_sort,priority:1"`
	Year       int64  `gorm:"column:year;type:int;not null;index:idx_wmbts_level_sort,priority:2,sort:desc"`
	Quarter    int64  `gorm:"column:quarter;type:tinyint;not null;index:idx_wmbts_level_sort,priority:3,sort:desc"`
	Month      int64  `gorm:"column:month;type:tinyint;not null;index:idx_wmbts_level_sort,priority:4,sort:desc"`
	Day        int64  `gorm:"column:day;type:tinyint;not null;index:idx_wmbts_level_sort,priority:5,sort:desc"`
	AggType    string `gorm:"column:agg_type;type:varchar(16);not null;index:idx_wmbts_scope_type_rank,priority:2;uniqueIndex:uk_wmbts_scope_type_key,priority:2"`
	AggKey     string `gorm:"column:agg_key;type:varchar(191);not null;uniqueIndex:uk_wmbts_scope_type_key,priority:3"`
	AggID      int64  `gorm:"column:agg_id;not null"`
	AggName    string `gorm:"column:agg_name;type:varchar(191);not null"`
	MediaCount int64  `gorm:"column:media_count;not null"`
	SizeBytes  int64  `gorm:"column:size_bytes;not null"`
	RankNo     int64  `gorm:"column:rank_no;not null;index:idx_wmbts_scope_type_rank,priority:3"`
	CreatedOn  int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn  int64  `gorm:"column:updated_on;not null;default:0"`
}

func (MediaBirthTopStat) TableName() string { return "w_media_birth_top_stat" }
