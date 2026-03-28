// internal/model/modelg/detail.go
package modelg

// Detail 表示抓取到的影片详情页（HTML 原文存档）
type Detail struct {
	Id    int64  `gorm:"primary_key;autoIncrement"`
	Name  string `gorm:"type:varchar(191);not null;comment:影片名称"`
	JavId string `gorm:"type:varchar(191);not null;unique;comment:JavId唯一标识"`

	Prefix   string `gorm:"type:varchar(64);not null;default:'';comment:前缀，如 ABC"`
	QueryUrl string `gorm:"type:varchar(300);not null;default:'';comment:来源请求URL"`
	Content  string `gorm:"type:longtext;not null;comment:详情页HTML内容"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

func (Detail) TableName() string {
	return "d_detail"
}
