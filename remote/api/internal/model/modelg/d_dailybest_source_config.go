package modelg

// DailyBestSourceConfig 保存远端唯一的榜单来源配置。
type DailyBestSourceConfig struct {
	ID         int64  `gorm:"primaryKey;autoIncrement:false"`
	JavAddress string `gorm:"not null;type:text"`
	CreatedOn  int64  `gorm:"not null"`
	UpdatedOn  int64  `gorm:"not null"`
}

func (DailyBestSourceConfig) TableName() string { return "d_dailybest_source_config" }
