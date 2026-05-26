package modelg

type Seed struct {
	Id         int64  `gorm:"primaryKey;autoIncrement"`
	Name       string `gorm:"type:varchar(128);not null;unique;default:'';comment:'查询名'"`
	Active     int64  `gorm:"type:tinyint;not null;default:0;comment:'状态'"`
	SearchType int64  `gorm:"type:tinyint;not null;default:0;comment:'查询类型'"`
	NameType   int64  `gorm:"type:tinyint;not null;default:0;comment:'名称类型'"`

	// 断点相关
	PageNow   int64 `gorm:"type:MEDIUMINT;not null;default:0;comment:'当前页'"`
	Offset    int64 `gorm:"type:MEDIUMINT;not null;default:0;comment:'偏移量'"`
	StartPage int64 `gorm:"type:MEDIUMINT;not null;default:0;comment:'起始页'"`
	EndPage   int64 `gorm:"type:MEDIUMINT;not null;default:0;comment:'结束页'"`

	// 运行状态
	LastQueryTime int64  `gorm:"not null;default:0;comment:'最后一次抓取时间'"`
	LastStatus    int64  `gorm:"type:tinyint;not null;default:0;comment:'最后状态'"`
	LastError     string `gorm:"type:varchar(255);not null;default:'';comment:'最后错误摘要'"`
	MovieTotal    int64  `gorm:"type:bigint;not null;default:0;comment:'当前 seed 对应 movie 总量'"`

	MovieLatestReleasingMovieJavId string `gorm:"column:movie_latest_releasing_movie_jav_id;type:varchar(64);not null;default:'';comment:'最新上映 movie 的 jav_id'"`
	MovieLatestReleasingMovieName  string `gorm:"column:movie_latest_releasing_movie_name;type:varchar(191);not null;default:'';comment:'最新上映 movie 的番号 name'"`

	MovieLastAddedTime       int64 `gorm:"not null;default:0;comment:'最后一次 movie 增加时间'"`
	LastInsertCount          int64 `gorm:"type:bigint;not null;default:0;comment:'本轮新增 movie 数'"`
	MovieLatestReleasingDate int64 `gorm:"not null;default:0;comment:'当前 seed 最新 movie 的上映时间'"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

func (Seed) TableName() string { return "d_seed" }
