package modelg

type RankDayStat struct {
	Id int64 `gorm:"not null;primaryKey"`

	DayNumber     int64  `gorm:"not null;type:int;uniqueIndex:uk_rank_day_stat_day_number;comment:排行日期序号"`
	RankDate      string `gorm:"not null;type:char(10);index:idx_rank_day_stat_rank_date;comment:排行业务日期"`
	PrevDayNumber int64  `gorm:"not null;type:int;default:0;index:idx_rank_day_stat_prev_day_number;comment:上一个存在排行数据的日期序号"`

	RankRowCount        int64  `gorm:"not null;type:bigint;default:0;comment:当天排行明细总数"`
	MovieCount          int64  `gorm:"not null;type:bigint;default:0;comment:当天去重影片数"`
	ContinuedMovieCount int64  `gorm:"not null;type:bigint;default:0;comment:与上一排行日连续上榜的影片数"`
	FirstSeenMovieCount int64  `gorm:"not null;type:bigint;default:0;comment:历史首次上榜的影片数"`
	ReturningMovieCount int64  `gorm:"not null;type:bigint;default:0;comment:上一排行日缺席后重新上榜的影片数"`
	DroppedMovieCount   int64  `gorm:"not null;type:bigint;default:0;comment:上一排行日在榜且当天掉榜的影片数"`
	RiseMovieCount      int64  `gorm:"not null;type:bigint;default:0;comment:相对上一排行日名次上升的影片数"`
	FallMovieCount      int64  `gorm:"not null;type:bigint;default:0;comment:相对上一排行日名次下降的影片数"`
	UnchangedMovieCount int64  `gorm:"not null;type:bigint;default:0;comment:相对上一排行日名次不变的影片数"`
	SourceChecksum      string `gorm:"not null;type:char(64);comment:当天排行源数据摘要"`
	CalculatedOn        int64  `gorm:"not null;default:0;comment:汇总计算完成时间"`
	CreatedOn           int64  `gorm:"not null;default:0"`
	UpdatedOn           int64  `gorm:"not null;default:0"`
}

const RankDayStatTableName = "c_rank_day_stat"

func (RankDayStat) TableName() string { return RankDayStatTableName }
