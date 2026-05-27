package modelg

type RankPeriodItem struct {
	Id int64 `gorm:"not null;primaryKey"`

	PeriodID   int64  `gorm:"not null;index:idx_period_rank,priority:1;uniqueIndex:uk_period_movie,priority:1;comment:周期榜ID"`
	MovieJavID string `gorm:"type:varchar(191);not null;uniqueIndex:uk_period_movie,priority:2;comment:影片ID"`

	RankPos int64   `gorm:"not null;type:int;index:idx_period_rank,priority:2;comment:周期名次"`
	Score   float64 `gorm:"not null;type:decimal(10,4);comment:周期得分"`

	DaysInRank         int64 `gorm:"not null;type:smallint;comment:周期内上榜天数"`
	UsedPickDays       int64 `gorm:"not null;type:smallint;comment:实际参与计算的天数"`
	FirstRankDayNumber int64 `gorm:"not null;type:int;comment:周期内首次上榜日序号"`
	LastRankDayNumber  int64 `gorm:"not null;type:int;comment:周期内最后上榜日序号"`
	BestRank           int64 `gorm:"not null;type:smallint;comment:最佳日名次"`
	WorstPickedRank    int64 `gorm:"not null;type:smallint;comment:入选样本中的最差名次"`
	PrevRank           int64 `gorm:"not null;type:int;comment:上期名次"`
	RankChange         int64 `gorm:"not null;type:int;comment:相对上期名次变化"`
	PeakRank           int64 `gorm:"not null;type:int;comment:历史最佳周期名次"`
	CreatedOn          int64 `gorm:"not null;default:0"`
	UpdatedOn          int64 `gorm:"not null;default:0"`
}

const RankPeriodItemTableName = "c_rank_period_item"

func (RankPeriodItem) TableName() string { return RankPeriodItemTableName }
