package modelg

type Movie struct {
	Id    int64  `gorm:"primaryKey"`
	JavId string `gorm:"not null;unique"`

	// 为了配合多种 order by，在 Name 上声明多个复合索引的第二列，并保留单列索引
	Name string `gorm:"not null;type:varchar(191);index:idx_name_only;index:idx_reldate_name,priority:2;index:idx_vnw_name,priority:2;index:idx_castage_name,priority:2;index:idx_score_name,priority:2;index:idx_detailupd_name,priority:2"`

	Title              string `gorm:"not null;type:varchar(500)"`
	Length             int64  `gorm:"not null;type:smallint"`
	ViewersNumberWant  int64  `gorm:"not null;type:MEDIUMINT"`
	ViewersNumberOwned int64  `gorm:"not null;type:MEDIUMINT"`

	ViewersNumberWatched int64 `gorm:"not null;type:MEDIUMINT;index:idx_vnw_name,priority:1"`
	DetailUpdateTime     int64 `gorm:"not null;default:0;index:idx_detailupd_name,priority:1"`
	ReleasingDate        int64 `gorm:"not null;index:idx_reldate_name,priority:1"`
	Score                int64 `gorm:"not null;type:tinyint;index:idx_score_name,priority:1"`

	PrefixId   int64 `gorm:"not null;index"`
	MakerId    int64 `gorm:"not null;index"`
	LabelId    int64 `gorm:"not null;index"`
	DirectorId int64 `gorm:"not null;index"`

	CastNumber     int64 `gorm:"not null;type:smallint"`
	CastAverageAge int64 `gorm:"not null;type:smallint;index:idx_castage_name,priority:1;comment:'平均年龄*10，例如237表示23.7岁'"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

const movieTableName = "a_movie"

func (i *Movie) TableName() string {
	return movieTableName
}
