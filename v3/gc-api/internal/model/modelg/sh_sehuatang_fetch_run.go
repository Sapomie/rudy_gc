package modelg

type SehuatangFetchRun struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	FetchMode      string `gorm:"column:fetch_mode;type:varchar(32);not null;index"`
	Keyword        string `gorm:"column:keyword;type:varchar(191);not null"`
	TargetForumIDs string `gorm:"column:target_forum_ids;type:varchar(64);not null"`
	StartURL       string `gorm:"column:start_url;type:varchar(500);not null"`
	CurrentURL     string `gorm:"column:current_url;type:varchar(500);not null"`
	CurrentPage    int64  `gorm:"column:current_page;not null;index"`
	Status         string `gorm:"column:status;type:varchar(32);not null;index"`

	SeenCount            int64 `gorm:"column:seen_count;not null;default:0"`
	MatchedCount         int64 `gorm:"column:matched_count;not null;default:0"`
	SkippedExistingCount int64 `gorm:"column:skipped_existing_count;not null;default:0"`
	SkippedForumCount    int64 `gorm:"column:skipped_forum_count;not null;default:0"`
	FetchedCount         int64 `gorm:"column:fetched_count;not null;default:0"`
	FailedCount          int64 `gorm:"column:failed_count;not null;default:0"`

	StartedOn  int64  `gorm:"column:started_on;not null;default:0;index"`
	FinishedOn int64  `gorm:"column:finished_on;not null;default:0"`
	LastError  string `gorm:"column:last_error;type:text;not null"`
	CreatedOn  int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn  int64  `gorm:"column:updated_on;not null;default:0"`
}

const sehuatangFetchRunTableName = "sh_sehuatang_fetch_run"

func (SehuatangFetchRun) TableName() string { return sehuatangFetchRunTableName }
