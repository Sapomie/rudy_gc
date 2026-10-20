package modelg

type DailyBestSyncRank struct {
	ID               int64  `gorm:"primaryKey"`
	RemoteFetchRunID int64  `gorm:"not null;uniqueIndex"`
	BusinessDate     string `gorm:"not null;type:char(10);index:idx_dailybest_sync_date_category,priority:1"`
	DayNumber        int64  `gorm:"not null;type:int;uniqueIndex:uk_dailybest_sync_day_category,priority:1"`
	Category         int64  `gorm:"not null;type:tinyint;index:idx_dailybest_sync_date_category,priority:2;uniqueIndex:uk_dailybest_sync_day_category,priority:2"`
	PageCount        int64  `gorm:"not null"`
	ItemCount        int64  `gorm:"not null"`
	Checksum         string `gorm:"not null;type:char(64)"`
	SyncStatus       string `gorm:"not null;type:varchar(32);index"`
	ProcessStatus    string `gorm:"not null;type:varchar(32);index"`
	RemoteFinishedAt int64  `gorm:"not null;default:0;comment:远端抓取完成时间"`
	SyncedAt         int64  `gorm:"not null;default:0"`
	ProcessedAt      int64  `gorm:"not null;default:0"`
	LastError        string `gorm:"not null;type:text"`
	CreatedOn        int64  `gorm:"not null"`
	UpdatedOn        int64  `gorm:"not null"`
}

func (DailyBestSyncRank) TableName() string { return "d_dailybest_sync_rank" }
