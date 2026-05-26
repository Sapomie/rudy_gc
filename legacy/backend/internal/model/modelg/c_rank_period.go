package modelg

type RankPeriod struct {
	Id int64 `gorm:"not null;primaryKey"`

	PeriodType int64  `gorm:"not null;type:tinyint;uniqueIndex:uk_period_type_key_category,priority:1;index:idx_period_lookup,priority:1;comment:周期类型(周/月/季/年)"`
	PeriodKey  string `gorm:"not null;type:varchar(32);uniqueIndex:uk_period_type_key_category,priority:2;index:idx_period_lookup,priority:2;comment:周期键"`
	Category   int64  `gorm:"not null;type:tinyint;uniqueIndex:uk_period_type_key_category,priority:3;index:idx_period_lookup,priority:3;comment:榜单类别"`

	StartDayNumber int64 `gorm:"not null;type:int;index:idx_period_range,priority:1;comment:周期开始日序号"`
	EndDayNumber   int64 `gorm:"not null;type:int;index:idx_period_range,priority:2;comment:周期结束日序号"`

	PickDays int64 `gorm:"not null;type:smallint;comment:固定取样天数"`
	TopN     int64 `gorm:"not null;type:smallint;comment:固定榜单数量"`
	Status   int64 `gorm:"not null;type:tinyint;index:idx_period_status;comment:状态"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

const RankPeriodTableName = "c_rank_period"

func (RankPeriod) TableName() string { return RankPeriodTableName }
