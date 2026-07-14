package schema

type ShSehuatangThreadDetailFetchRun struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	JobID              string `gorm:"column:job_id;type:varchar(128);not null;uniqueIndex"`
	TriggerSource      string `gorm:"column:trigger_source;type:varchar(32);not null;index"`
	TargetForumIDs     string `gorm:"column:target_forum_ids;type:varchar(64);not null"`
	SelectionCriteria  string `gorm:"column:selection_criteria;type:json;not null"`
	SelectedThreadTIDs string `gorm:"column:selected_thread_tids;type:json;not null"`
	SelectedCount      int64  `gorm:"column:selected_count;not null;default:0"`
	CurrentIndex       int64  `gorm:"column:current_index;not null;default:0"`
	CurrentThreadTID   string `gorm:"column:current_thread_tid;type:varchar(64);not null;index"`
	CurrentThreadURL   string `gorm:"column:current_thread_url;type:varchar(500);not null"`
	CompletedCount     int64  `gorm:"column:completed_count;not null;default:0"`
	FetchedCount       int64  `gorm:"column:fetched_count;not null;default:0"`
	FailedCount        int64  `gorm:"column:failed_count;not null;default:0"`

	Status     string `gorm:"column:status;type:varchar(32);not null;index"`
	StopReason string `gorm:"column:stop_reason;type:varchar(64);not null"`
	StartedOn  int64  `gorm:"column:started_on;not null;default:0;index"`
	FinishedOn int64  `gorm:"column:finished_on;not null;default:0"`
	LastError  string `gorm:"column:last_error;type:text;not null"`
	CreatedOn  int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn  int64  `gorm:"column:updated_on;not null;default:0"`
}

const shSehuatangThreadDetailFetchRunTableName = "sh_detail_run"

func (ShSehuatangThreadDetailFetchRun) TableName() string {
	return shSehuatangThreadDetailFetchRunTableName
}
