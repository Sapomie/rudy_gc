package modelg

type SehuatangFetchItem struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	RunSource     string `gorm:"column:run_source;type:varchar(32);not null;default:'fetch';uniqueIndex:uk_run_thread,priority:1;index"`
	RunID         int64  `gorm:"column:run_id;not null;uniqueIndex:uk_run_thread,priority:2;index"`
	FetchMode     string `gorm:"column:fetch_mode;type:varchar(32);not null;index"`
	SourceForumID string `gorm:"column:source_forum_id;type:varchar(32);not null;index"`
	SourcePage    int64  `gorm:"column:source_page;not null;index"`
	SourcePageURL string `gorm:"column:source_page_url;type:varchar(500);not null"`

	ThreadTID   string `gorm:"column:thread_tid;type:varchar(64);not null;uniqueIndex:uk_run_thread,priority:3;index"`
	ThreadURL   string `gorm:"column:thread_url;type:varchar(500);not null"`
	ThreadTitle string `gorm:"column:thread_title;type:varchar(500);not null"`
	AuthorName  string `gorm:"column:author_name;type:varchar(191);not null"`

	Status      string `gorm:"column:status;type:varchar(32);not null;index"`
	ThreadRowID int64  `gorm:"column:thread_row_id;not null;default:0;index"`
	Error       string `gorm:"column:error;type:text;not null"`

	CreatedOn int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64 `gorm:"column:updated_on;not null;default:0"`
}

const sehuatangFetchItemTableName = "sh_item"

func (SehuatangFetchItem) TableName() string { return sehuatangFetchItemTableName }
