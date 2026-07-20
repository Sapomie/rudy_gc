package modelg

type RankPeriodStat struct {
	Id int64 `gorm:"not null;primaryKey"`

	PeriodID     int64 `gorm:"not null;uniqueIndex:uk_rank_period_stat_period_id;comment:周期榜ID"`
	PrevPeriodID int64 `gorm:"not null;default:0;index:idx_rank_period_stat_prev_period_id;comment:同类型同分类的上一周期榜ID"`

	SourceRankRowCount  int64 `gorm:"not null;type:bigint;default:0;comment:周期内原始排行明细总数"`
	SourceRankDayCount  int64 `gorm:"not null;type:bigint;default:0;comment:周期内实际包含排行数据的天数"`
	SourceMovieCount    int64 `gorm:"not null;type:bigint;default:0;comment:周期内原始排行去重影片数"`
	MissingRankDayCount int64 `gorm:"not null;type:bigint;default:0;comment:周期内缺失排行数据的天数"`
	PeriodMovieCount    int64 `gorm:"not null;type:bigint;default:0;comment:最终周期榜单影片数"`
	ContinuedMovieCount int64 `gorm:"not null;type:bigint;default:0;comment:上期和本期均在榜的影片数"`
	FirstSeenMovieCount int64 `gorm:"not null;type:bigint;default:0;comment:历史首次进入同类型同分类周期榜的影片数"`
	ReturningMovieCount int64 `gorm:"not null;type:bigint;default:0;comment:上期缺席后重新进入周期榜的影片数"`
	DroppedMovieCount   int64 `gorm:"not null;type:bigint;default:0;comment:上期在榜且本期退出的影片数"`
	RiseMovieCount      int64 `gorm:"not null;type:bigint;default:0;comment:相对上期名次上升的影片数"`
	FallMovieCount      int64 `gorm:"not null;type:bigint;default:0;comment:相对上期名次下降的影片数"`
	UnchangedMovieCount int64 `gorm:"not null;type:bigint;default:0;comment:相对上期名次不变的影片数"`
	PeakMovieCount      int64 `gorm:"not null;type:bigint;default:0;comment:达到历史最佳周期名次的影片数"`

	RankOneMovieJavID    string  `gorm:"not null;type:varchar(191);default:'';comment:周期第一名影片ID"`
	AverageScore         float64 `gorm:"not null;type:decimal(10,4);default:0;comment:周期榜单平均得分"`
	AverageDaysInRank    float64 `gorm:"not null;type:decimal(10,4);default:0;comment:周期榜单平均上榜天数"`
	SourceFetchedStartOn int64   `gorm:"not null;default:0;comment:周期源数据最早抓取完成时间"`
	SourceFetchedEndOn   int64   `gorm:"not null;default:0;index:idx_rank_period_stat_source_fetched_end_on;comment:周期源数据最晚抓取完成时间"`
	SourceChecksum       string  `gorm:"not null;type:char(64);comment:周期排行源数据摘要"`
	CalculatedOn         int64   `gorm:"not null;default:0;comment:汇总计算完成时间"`
	CreatedOn            int64   `gorm:"not null;default:0"`
	UpdatedOn            int64   `gorm:"not null;default:0"`
}

const RankPeriodStatTableName = "c_rank_period_stat"

func (RankPeriodStat) TableName() string { return RankPeriodStatTableName }
