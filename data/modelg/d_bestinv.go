// data/modelg/d_bestinv.go
package modelg

// Bestinv 表示最佳影片排行的原始 HTML 存档
type Bestinv struct {
	Id   int64  `gorm:"primaryKey"`
	Name string `gorm:"not null;type:varchar(191);uniqueIndex;comment:唯一名(由query+page组合)"`

	NeedScan      int64 `gorm:"not null;type:tinyint;default:0;index;comment:是否需要进一步扫描(由业务常量定义)"`
	NeedRankCheck int64 `gorm:"not null;type:tinyint;default:0;index;comment:是否需要检查排名(由业务常量定义)"`

	Category  int64 `gorm:"not null;type:tinyint;default:0;comment:类别(取值由业务常量定义)"`
	Page      int64 `gorm:"not null;type:smallint;default:1;comment:页码"`
	DayNumber int64 `gorm:"not null;type:int;default:0;index;comment:排名对应的日序号"`

	Content       string `gorm:"not null;type:longtext;comment:页面HTML原文"`
	LastQueryTime int64  `gorm:"not null;default:0;comment:最后抓取时间(Unix秒)"`
	Date          string `gorm:"not null;type:varchar(32);default:'';comment:抓取日期字符串"`

	CreatedOn int64 `gorm:"not null;default:0;comment:创建时间(Unix秒)"`
	UpdatedOn int64 `gorm:"not null;default:0;comment:更新时间(Unix秒)"`
}

const bestinvTableName = "d_bestinv"

func (Bestinv) TableName() string {
	return bestinvTableName
}
