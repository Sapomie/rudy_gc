package modelg

type DailyBestFetchRun struct {
	ID           int64  `gorm:"primaryKey"`
	BusinessDate string `gorm:"not null;type:char(10);index:idx_dailybest_run_date_category,priority:1"`
	Category     int64  `gorm:"not null;type:tinyint;index:idx_dailybest_run_date_category,priority:2"`
	TriggerType  string `gorm:"not null;type:varchar(32)"`
	Status       string `gorm:"not null;type:varchar(32);index"`
	StartedAt    int64  `gorm:"not null;default:0"`
	FinishedAt   int64  `gorm:"not null;default:0"`
	PageCount    int64  `gorm:"not null;default:0"`
	ItemCount    int64  `gorm:"not null;default:0"`
	Checksum     string `gorm:"not null;type:char(64);default:'';index"`
	ErrorMessage string `gorm:"not null;type:text"`
	CreatedOn    int64  `gorm:"not null"`
	UpdatedOn    int64  `gorm:"not null"`
}

func (DailyBestFetchRun) TableName() string { return "d_dailybest_fetch_run" }
