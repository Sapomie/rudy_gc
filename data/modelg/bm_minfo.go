package modelg

type Minfo struct {
	// base info
	Id    int64  `gorm:"primaryKey"`
	JavId string `gorm:"not null;unique"`
	Name  string `gorm:"not null;type:varchar(191);index:idx_name_only;index:idx_first_rank_name,priority:2;index:idx_highest_rank_name,priority:2;index:idx_days_in_rank_name,priority:2"`

	// other
	Chinese    string `gorm:"not null;type:varchar(300)"`
	EncodeName string `gorm:"not null;type:varchar(191);index"`

	// ORDER BY first_rank_day_number desc, name desc
	FirstRankDayNumber int64 `gorm:"not null;index:idx_first_rank_name,priority:1"`
	HighestRank        int64 `gorm:"not null;index:idx_highest_rank_name,priority:1"`
	DaysInRank         int64 `gorm:"not null;index:idx_days_in_rank_name,priority:1"`
	NeedDownload       int64 `gorm:"not null;type:tinyint;index"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

const minfoTableName = "bm_minfo"

func (i *Minfo) TableName() string {
	return minfoTableName
}
