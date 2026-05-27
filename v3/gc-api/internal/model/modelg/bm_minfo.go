package modelg

type Minfo struct {
	Id                 int64  `gorm:"primaryKey"`
	JavId              string `gorm:"not null;type:varchar(191);uniqueIndex:uniq_jav_id"`
	Name               string `gorm:"not null;type:varchar(191);index:idx_name_only;index:idx_first_rank_name,priority:2;index:idx_highest_rank_name,priority:2;index:idx_days_in_rank_name,priority:2;index:idx_reldate_name,priority:2"`
	Chinese            string `gorm:"not null;type:varchar(300)"`
	FirstRankDayNumber int64  `gorm:"not null;index:idx_first_rank_name,priority:1"`
	HighestRank        int64  `gorm:"not null;index:idx_highest_rank_name,priority:1"`
	DaysInRank         int64  `gorm:"not null;index:idx_days_in_rank_name,priority:1"`
	ReleasingDate      int64  `gorm:"not null;index:idx_reldate_name,priority:1"`
	CreatedOn          int64  `gorm:"not null;default:0"`
	UpdatedOn          int64  `gorm:"not null;default:0"`
}

const minfoTableName = "bm_minfo"

func (i *Minfo) TableName() string {
	return minfoTableName
}
