package modelg

type FetchSite struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	SiteCode string `gorm:"column:site_code;type:varchar(64);not null;uniqueIndex"`
	SiteName string `gorm:"column:site_name;type:varchar(128);not null"`
	BaseURL  string `gorm:"column:base_url;type:varchar(255);not null"`

	UserAgent string `gorm:"column:user_agent;type:varchar(500);not null"`
	Cookie    string `gorm:"column:cookie;type:text;not null"`
	Proxy     string `gorm:"column:proxy;type:varchar(255);not null"`

	TimeoutSeconds int64 `gorm:"column:timeout_seconds;type:int;not null"`
	RequestSleepMs int64 `gorm:"column:request_sleep_ms;type:int;not null"`
	RetrySleepMs   int64 `gorm:"column:retry_sleep_ms;type:int;not null"`
	MaxRetryTimes  int64 `gorm:"column:max_retry_times;type:tinyint;not null"`

	Status int8 `gorm:"column:status;type:tinyint;not null;index"`

	CreatedOn int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64 `gorm:"column:updated_on;not null;default:0"`
}

const fetchSiteTableName = "t_fetch_site"

func (FetchSite) TableName() string { return fetchSiteTableName }
