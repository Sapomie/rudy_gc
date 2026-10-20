package modelg

type RankPeriodStat struct {
	Id int64 `gorm:"not null;primaryKey"`

	PeriodType     int64  `gorm:"not null;type:tinyint;uniqueIndex:uk_rank_period_stat_period,priority:1;comment:统计周期类型(周/月/季/年)"`
	PeriodKey      string `gorm:"not null;type:varchar(32);uniqueIndex:uk_rank_period_stat_period,priority:2;comment:统计周期键"`
	StartDayNumber int64  `gorm:"not null;type:int;index:idx_rank_period_stat_range,priority:1;comment:统计周期开始日序号"`
	EndDayNumber   int64  `gorm:"not null;type:int;index:idx_rank_period_stat_range,priority:2;comment:统计周期结束日序号"`

	SourceRankRowCount  int64 `gorm:"not null;type:bigint;default:0;comment:周期内原始排行明细总数"`
	SourceRankDayCount  int64 `gorm:"not null;type:bigint;default:0;comment:周期内实际包含排行数据的天数"`
	SourceMovieCount    int64 `gorm:"not null;type:bigint;default:0;comment:周期内原始排行去重影片数"`
	MissingRankDayCount int64 `gorm:"not null;type:bigint;default:0;comment:周期内缺失排行数据的天数"`
	PeriodMovieCount    int64 `gorm:"not null;type:bigint;default:0;comment:周期内每日去重影片累计数"`
	NewEntryMovieCount  int64 `gorm:"not null;type:bigint;default:0;comment:周期内新上榜影片累计数"`

	SourceFetchedStartOn int64  `gorm:"not null;default:0;comment:周期源数据最早抓取完成时间"`
	SourceFetchedEndOn   int64  `gorm:"not null;default:0;index:idx_rank_period_stat_source_fetched_end_on;comment:周期源数据最晚抓取完成时间"`
	SourceChecksum       string `gorm:"not null;type:char(64);comment:周期排行源数据摘要"`
	CalculatedOn         int64  `gorm:"not null;default:0;comment:汇总计算完成时间"`
	CreatedOn            int64  `gorm:"not null;default:0"`
	UpdatedOn            int64  `gorm:"not null;default:0"`
}

const RankPeriodStatTableName = "c_rank_period_stat"

func (RankPeriodStat) TableName() string { return RankPeriodStatTableName }
