package modelg

type MovieReleaseTopStat struct {
	ID         int64  `gorm:"primaryKey;autoIncrement;column:id"`
	ScopeKey   string `gorm:"column:scope_key;type:varchar(32);not null;index:idx_mrts_scope_type_rank,priority:1;uniqueIndex:uk_mrts_scope_type_key,priority:1"`
	Level      string `gorm:"column:level;type:varchar(16);not null;index:idx_mrts_level_sort,priority:1"`
	Year       int64  `gorm:"column:year;type:int;not null;index:idx_mrts_level_sort,priority:2,sort:desc"`
	Quarter    int64  `gorm:"column:quarter;type:tinyint;not null;index:idx_mrts_level_sort,priority:3,sort:desc"`
	Month      int64  `gorm:"column:month;type:tinyint;not null;index:idx_mrts_level_sort,priority:4,sort:desc"`
	AggType    string `gorm:"column:agg_type;type:varchar(16);not null;index:idx_mrts_scope_type_rank,priority:2;uniqueIndex:uk_mrts_scope_type_key,priority:2"`
	AggKey     string `gorm:"column:agg_key;type:varchar(191);not null;uniqueIndex:uk_mrts_scope_type_key,priority:3"`
	AggID      int64  `gorm:"column:agg_id;not null"`
	AggName    string `gorm:"column:agg_name;type:varchar(191);not null"`
	CountAll   int64  `gorm:"column:count_all;not null"`
	CountOwned int64  `gorm:"column:count_owned;not null"`
	RankNo     int64  `gorm:"column:rank_no;not null;index:idx_mrts_scope_type_rank,priority:3"`
	CreatedOn  int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn  int64  `gorm:"column:updated_on;not null;default:0"`
}

func (MovieReleaseTopStat) TableName() string { return "movie_release_top_stat" }
