package modelg

type DailyBestMovie struct {
	Id         int64  `gorm:"not null;primaryKey"`
	MovieJavID string `gorm:"not null;type:varchar(191);uniqueIndex;comment:DailyBest 影片唯一标识"`
	MovieName  string `gorm:"not null;type:varchar(191);default:'';comment:源站影片名称"`
	MovieCode  string `gorm:"not null;type:varchar(191);default:'';comment:影片番号"`
	Title      string `gorm:"not null;type:text;comment:完整标题"`
	CoverURL   string `gorm:"not null;type:text;comment:封面地址"`
	CreatedOn  int64  `gorm:"not null;default:0"`
	UpdatedOn  int64  `gorm:"not null;default:0"`
}

const DailyBestMovieTableName = "d_dailybest_movie"

func (DailyBestMovie) TableName() string { return DailyBestMovieTableName }
