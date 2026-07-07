package modelg

type SehuatangThread struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	ThreadTID   string `gorm:"column:thread_tid;type:varchar(64);not null;uniqueIndex:uk_thread_tid"`
	ThreadURL   string `gorm:"column:thread_url;type:varchar(500);not null"`
	ThreadTitle string `gorm:"column:thread_title;type:varchar(500);not null"`
	AuthorName  string `gorm:"column:author_name;type:varchar(191);not null"`
	PostTime    int64  `gorm:"column:post_time;not null;index"`

	SourceForumID string `gorm:"column:source_forum_id;type:varchar(32);not null;default:'';index"`
	SourcePage    int64  `gorm:"column:source_page;not null;default:0;index"`
	SourcePageURL string `gorm:"column:source_page_url;type:varchar(500);not null;default:''"`

	DetailStatus    string `gorm:"column:detail_status;type:varchar(32);not null;default:'pending';index"`
	DetailFetchedOn int64  `gorm:"column:detail_fetched_on;not null;default:0;index"`
	DetailFailedOn  int64  `gorm:"column:detail_failed_on;not null;default:0"`
	DetailError     string `gorm:"column:detail_error;type:longtext"`

	RawContentText string `gorm:"column:raw_content_text;type:longtext;not null"`
	MagnetText     string `gorm:"column:magnet_text;type:text;not null"`
	InfoHashText   string `gorm:"column:info_hash_text;type:text;not null"`
	ContentHash    string `gorm:"column:content_hash;type:char(64);not null;index"`

	LastSeenTime int64 `gorm:"column:last_seen_time;not null;index"`
	CreatedOn    int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn    int64 `gorm:"column:updated_on;not null;default:0"`
}

const sehuatangThreadTableName = "sh_sehuatang_thread"

func (SehuatangThread) TableName() string { return sehuatangThreadTableName }
