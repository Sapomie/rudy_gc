// data/modelg/inventory.go
package modelg

// Inventory 表示已抓取的原始库存页（HTML 原文存档）
type Inventory struct {
	Id            int64  `gorm:"primary_key;autoIncrement"`
	Name          string `gorm:"type:varchar(191);not null;unique;comment:唯一名(由query+page组合)"`
	NeedScan      int64  `gorm:"not null;type:tinyint;default:1;index;comment:是否需要扫描(0=不需要,1=需要)"`
	Keyword       string `gorm:"type:varchar(191);not null;default:'';comment:来源的关键字(前缀或标签)"`
	Parent        string `gorm:"type:varchar(191);not null;default:'';comment:上级queryBy"`
	Page          int64  `gorm:"type:mediumint;not null;default:1;comment:页码"`
	Content       string `gorm:"type:longtext;not null;comment:页面HTML内容"`
	Category      int64  `gorm:"not null;type:tinyint;comment:类别(1=Prefix,2=Label)"`
	LastQueryTime int64  `gorm:"not null;default:0;comment:最后抓取时间(Unix秒)"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

func (Inventory) TableName() string {
	return "d_inventory"
}
