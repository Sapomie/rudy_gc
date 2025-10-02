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

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

func (Seed) TableName() string { return "d_seed" }
