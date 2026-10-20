package modelg

type AggEvent struct {
	ID           int64  `gorm:"primaryKey;autoIncrement;column:id"`
	AggKey       string `gorm:"column:agg_key;type:varchar(64);not null;index:idx_w_agg_event_agg_key"`
	FlowKey      string `gorm:"column:flow_key;type:varchar(64);not null;index:idx_w_agg_event_flow_key"`
	Status       string `gorm:"column:status;type:varchar(32);not null;index:idx_w_agg_event_status"`
	ScopeCount   int64  `gorm:"column:scope_count;not null"`
	BucketCount  int64  `gorm:"column:bucket_count;not null"`
	TopCount     int64  `gorm:"column:top_count;not null"`
	StartedTime  int64  `gorm:"column:started_time;not null;index:idx_w_agg_event_started_time"`
	FinishedTime int64  `gorm:"column:finished_time;not null"`
	DurationMs   int64  `gorm:"column:duration_ms;not null"`
	ErrorMessage string `gorm:"column:error_message;type:text;not null"`
	CreatedTime  int64  `gorm:"column:created_time;not null"`
	UpdatedTime  int64  `gorm:"column:updated_time;not null"`
}

func (AggEvent) TableName() string { return "w_agg_event" }
