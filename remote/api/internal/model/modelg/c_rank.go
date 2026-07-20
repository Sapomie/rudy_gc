package modelg

type Rank struct {
	Id      int64  `gorm:"not null;primaryKey"`
	RankKey string `gorm:"not null;type:varchar(191);uniqueIndex;comment:日期、分类和名次组成的唯一键"`

	MovieJavID string `gorm:"not null;index:idx_movie_day"`
	DayNumber  int64  `gorm:"not null;type:smallint;index:idx_movie_day;index:idx_day_number;index:idx_day_rankpos,priority:1;uniqueIndex:uk_rank_day_category_pos,priority:1;comment:第几天"`
	RankPos    int64  `gorm:"not null;type:smallint;index:idx_movie_day;index:idx_day_rankpos,priority:2;uniqueIndex:uk_rank_day_category_pos,priority:3;comment:当天排名"`
	Category   int64  `gorm:"not null;type:tinyint;uniqueIndex:uk_rank_day_category_pos,priority:2;comment:榜单类别(月榜/总榜等)"`
	CreatedOn  int64  `gorm:"not null;default:0"`
	UpdatedOn  int64  `gorm:"not null;default:0"`
}

const RankTableName = "c_rank"

func (Rank) TableName() string { return RankTableName }
