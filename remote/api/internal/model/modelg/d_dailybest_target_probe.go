package modelg

type DailyBestTargetProbe struct {
	ID            int64  `gorm:"primaryKey;index:idx_dailybest_target_probe_started_id,priority:2"`
	RequestID     string `gorm:"not null;type:char(32);uniqueIndex"`
	Status        string `gorm:"not null;type:varchar(32);index:idx_dailybest_target_probe_status_started,priority:1"`
	TargetURL     string `gorm:"not null;type:varchar(2048)"`
	FinalURL      string `gorm:"not null;type:varchar(2048)"`
	HTTPStatus    int64  `gorm:"not null;type:int;default:0"`
	DurationMS    int64  `gorm:"not null;default:0"`
	BrowserStatus string `gorm:"not null;type:varchar(32);default:''"`
	ErrorType     string `gorm:"not null;type:varchar(64);default:'';index"`
	ErrorMessage  string `gorm:"not null;type:text"`
	StartedAt     int64  `gorm:"not null;default:0;index:idx_dailybest_target_probe_started_id,priority:1;index:idx_dailybest_target_probe_status_started,priority:2"`
	FinishedAt    int64  `gorm:"not null;default:0"`
	CreatedOn     int64  `gorm:"not null"`
	UpdatedOn     int64  `gorm:"not null"`
}

func (DailyBestTargetProbe) TableName() string { return "d_dailybest_target_probe" }
