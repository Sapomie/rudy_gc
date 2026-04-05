package modelg

type MovieReleaseAggDirty struct {
	ID          int64  `gorm:"primaryKey;autoIncrement;column:id"`
	BucketMonth int64  `gorm:"column:bucket_month;not null;uniqueIndex:uk_mrad_bucket_month"`
	ScopeKey    string `gorm:"column:scope_key;type:varchar(32);not null;index:idx_mrad_scope_key"`
	CreatedOn   int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn   int64  `gorm:"column:updated_on;not null;default:0"`
}

func (MovieReleaseAggDirty) TableName() string { return "movie_release_agg_dirty" }
