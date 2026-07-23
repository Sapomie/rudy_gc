package modelg

type DailyBestRuntimeEvent struct {
	ID            int64  `gorm:"primaryKey"`
	Component     string `gorm:"not null;type:varchar(32);index:idx_dailybest_runtime_event_component_time,priority:1"`
	EventType     string `gorm:"not null;type:varchar(64);index"`
	Level         string `gorm:"not null;type:varchar(16);index"`
	Message       string `gorm:"not null;type:text"`
	FetchRunID    int64  `gorm:"not null;default:0;index"`
	TargetProbeID int64  `gorm:"not null;default:0;index"`
	OccurredAt    int64  `gorm:"not null;index:idx_dailybest_runtime_event_component_time,priority:2"`
	CreatedOn     int64  `gorm:"not null"`
}

func (DailyBestRuntimeEvent) TableName() string { return "d_dailybest_runtime_event" }
