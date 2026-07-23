package modelg

type DailyBestFetchLog struct {
	ID            int64  `gorm:"primaryKey"`
	FetchRunID    int64  `gorm:"not null;index:idx_dailybest_fetch_log_run_page,priority:1"`
	PageNumber    int64  `gorm:"not null;type:int;index:idx_dailybest_fetch_log_run_page,priority:2"`
	AttemptNumber int64  `gorm:"not null;type:int;default:1"`
	RequestURL    string `gorm:"not null;type:varchar(2048)"`
	Status        string `gorm:"not null;type:varchar(32);index"`
	HTTPStatus    int64  `gorm:"not null;type:int;default:0"`
	DurationMS    int64  `gorm:"not null;default:0"`
	ErrorMessage  string `gorm:"not null;type:text"`
	CreatedOn     int64  `gorm:"not null"`
	UpdatedOn     int64  `gorm:"not null"`
}

func (DailyBestFetchLog) TableName() string { return "d_dailybest_fetch_log" }
