package modelg

type MediaAggDirty struct {
	ID        int64  `gorm:"primaryKey;autoIncrement;column:id"`
	BucketDay int64  `gorm:"column:bucket_day;not null;uniqueIndex:uk_wmad_bucket_day"`
	ScopeKey  string `gorm:"column:scope_key;type:varchar(32);not null;index:idx_wmad_scope_key"`
	CreatedOn int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64  `gorm:"column:updated_on;not null;default:0"`
}

func (MediaAggDirty) TableName() string { return "w_media_agg_dirty" }
