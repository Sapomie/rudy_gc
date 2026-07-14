package schema

type ShSehuatangThreadListFetchRun struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	JobID              string `gorm:"column:job_id;type:varchar(128);not null;uniqueIndex"`
	TargetForumIDs     string `gorm:"column:target_forum_ids;type:varchar(64);not null"`
	StartPage          int64  `gorm:"column:start_page;not null;default:0"`
	EndPage            int64  `gorm:"column:end_page;not null;default:0"`
	CurrentForumID     string `gorm:"column:current_forum_id;type:varchar(32);not null;index"`
	CurrentPage        int64  `gorm:"column:current_page;not null;default:0;index"`
	CurrentURL         string `gorm:"column:current_url;type:varchar(500);not null"`
	TotalPageCount     int64  `gorm:"column:total_page_count;not null;default:0"`
	CompletedPageCount int64  `gorm:"column:completed_page_count;not null;default:0"`
	SuccessPageCount   int64  `gorm:"column:success_page_count;not null;default:0"`
	FailedPageCount    int64  `gorm:"column:failed_page_count;not null;default:0"`

	SeenCount       int64 `gorm:"column:seen_count;not null;default:0"`
	InsertedCount   int64 `gorm:"column:inserted_count;not null;default:0"`
	ExistingCount   int64 `gorm:"column:existing_count;not null;default:0"`
	SkippedTopCount int64 `gorm:"column:skipped_top_count;not null;default:0"`
	FailedCount     int64 `gorm:"column:failed_count;not null;default:0"`

	CoveragePostTimeFrom int64  `gorm:"column:coverage_post_time_from;not null;default:0;index:idx_sh_thread_list_fetch_run_coverage,priority:1"`
	CoveragePostTimeTo   int64  `gorm:"column:coverage_post_time_to;not null;default:0;index:idx_sh_thread_list_fetch_run_coverage,priority:2"`
	CoverageComplete     bool   `gorm:"column:coverage_complete;not null;default:false"`
	PageCoverage         string `gorm:"column:page_coverage;type:json;not null"`

	Status     string `gorm:"column:status;type:varchar(32);not null;index"`
	StopReason string `gorm:"column:stop_reason;type:varchar(64);not null"`
	StartedOn  int64  `gorm:"column:started_on;not null;default:0;index"`
	FinishedOn int64  `gorm:"column:finished_on;not null;default:0"`
	LastError  string `gorm:"column:last_error;type:text;not null"`
	CreatedOn  int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn  int64  `gorm:"column:updated_on;not null;default:0"`
}

const shSehuatangThreadListFetchRunTableName = "sh_list_run"

func (ShSehuatangThreadListFetchRun) TableName() string {
	return shSehuatangThreadListFetchRunTableName
}
