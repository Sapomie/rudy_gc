// data/modelg/d_seed.go
package modelg

// NameType: 1=Prefix, 2=Label
// SearchType: 1=Offset, 2=StartEnd
type Seed struct {
	Id         int64  `gorm:"primary_key;autoIncrement"`
	Name       string `gorm:"type:varchar(128);not null;unique;default:'';comment:查询名(前缀/标签具体值)"`
	Active     int64  `gorm:"not null;type:tinyint;default:1;comment:状态(1=inactive,2=active)"`
	SearchType int64  `gorm:"not null;type:tinyint;default:1;comment:1=Offset,2=StartEnd"`
	NameType   int64  `gorm:"not null;type:tinyint;default:1;comment:1=Prefix,2=Label"`

	// 断点相关
	PageNow   int64 `gorm:"not null;type:MEDIUMINT;default:1;comment:当前断点页(最近成功处理到的页码)"`
	Offset    int64 `gorm:"not null;type:MEDIUMINT;default:2;comment:仅在SearchType=Offset时有效"`
	StartPage int64 `gorm:"not null;type:MEDIUMINT;default:1;comment:仅在SearchType=StartEnd时有效"`
	EndPage   int64 `gorm:"not null;type:MEDIUMINT;default:1;comment:仅在SearchType=StartEnd时有效"`

	// 运行状态
	LastQueryTime int64  `gorm:"not null;default:0;comment:最后一次开始抓取的Unix时间戳(秒,0=未运行)"`
	LastStatus    int64  `gorm:"not null;type:tinyint;default:0;comment:0=idle,1=ok,2=empty,3=error"`
	LastError     string `gorm:"type:varchar(255);not null;default:'';comment:最后一次错误摘要"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

func (Seed) TableName() string { return "d_seed" }
