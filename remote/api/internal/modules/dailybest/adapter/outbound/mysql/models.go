package dailybestremotex

type FetchRun struct {
	ID           int64  `gorm:"primaryKey"`
	BusinessDate string `gorm:"not null;type:char(10);index:idx_dailybest_run_date_category,priority:1"`
	Category     int64  `gorm:"not null;type:tinyint;index:idx_dailybest_run_date_category,priority:2"`
	TriggerType  string `gorm:"not null;type:varchar(32)"`
	Status       string `gorm:"not null;type:varchar(32);index"`
	StartedAt    int64  `gorm:"not null;default:0"`
	FinishedAt   int64  `gorm:"not null;default:0"`
	PageCount    int64  `gorm:"not null;default:0"`
	ItemCount    int64  `gorm:"not null;default:0"`
	SnapshotID   int64  `gorm:"not null;default:0;index"`
	ErrorMessage string `gorm:"not null;type:text"`
	CreatedAt    int64  `gorm:"not null"`
	UpdatedAt    int64  `gorm:"not null"`
}

func (FetchRun) TableName() string { return "dailybest_fetch_run" }

type Snapshot struct {
	ID           int64  `gorm:"primaryKey"`
	FetchRunID   int64  `gorm:"not null;uniqueIndex"`
	BusinessDate string `gorm:"not null;type:char(10);index:idx_dailybest_snapshot_date_category_revision,priority:1;uniqueIndex:uk_dailybest_snapshot_content,priority:1"`
	Category     int64  `gorm:"not null;type:tinyint;index:idx_dailybest_snapshot_date_category_revision,priority:2;uniqueIndex:uk_dailybest_snapshot_content,priority:2"`
	PageCount    int64  `gorm:"not null"`
	ItemCount    int64  `gorm:"not null"`
	Checksum     string `gorm:"not null;type:char(64);uniqueIndex:uk_dailybest_snapshot_content,priority:3"`
	PublishedAt  int64  `gorm:"not null"`
	CreatedAt    int64  `gorm:"not null"`
}

func (Snapshot) TableName() string { return "dailybest_snapshot" }

type SnapshotItem struct {
	ID          int64  `gorm:"primaryKey"`
	SnapshotID  int64  `gorm:"not null;uniqueIndex:uk_dailybest_snapshot_rank,priority:1;uniqueIndex:uk_dailybest_snapshot_jav,priority:1;index"`
	RankPos     int64  `gorm:"not null;uniqueIndex:uk_dailybest_snapshot_rank,priority:2"`
	SourcePage  int64  `gorm:"not null"`
	SourceIndex int64  `gorm:"not null"`
	JavID       string `gorm:"not null;type:varchar(191);uniqueIndex:uk_dailybest_snapshot_jav,priority:2;index"`
	MovieName   string `gorm:"not null;type:varchar(191)"`
	MovieCode   string `gorm:"not null;type:varchar(191)"`
	Title       string `gorm:"not null;type:text"`
	CoverURL    string `gorm:"not null;type:text"`
	CreatedAt   int64  `gorm:"not null"`
}

func (SnapshotItem) TableName() string { return "dailybest_snapshot_item" }

type SnapshotPage struct {
	ID         int64  `gorm:"primaryKey"`
	SnapshotID int64  `gorm:"not null;uniqueIndex:uk_dailybest_snapshot_page,priority:1;index"`
	PageNumber int64  `gorm:"not null;uniqueIndex:uk_dailybest_snapshot_page,priority:2"`
	SourceURL  string `gorm:"not null;type:text"`
	Content    string `gorm:"not null;type:longtext"`
	Checksum   string `gorm:"not null;type:char(64)"`
	CreatedAt  int64  `gorm:"not null"`
}

func (SnapshotPage) TableName() string { return "dailybest_snapshot_page" }
