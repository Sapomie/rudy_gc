// internal/model/modelg/inventory.go
package modelg

// Inventory 表示已抓取的原始库存页（HTML 原文存档）
type Inventory struct {
	Id            int64  `gorm:"primaryKey;autoIncrement"`
	Name          string `gorm:"type:varchar(191);not null;unique;comment:'唯一名'"`
	NeedScan      int64  `gorm:"type:tinyint;not null;default:1;index;comment:'是否需要扫描'"`
	Keyword       string `gorm:"type:varchar(191);not null;default:'';comment:'来源关键字'"`
	Parent        string `gorm:"type:varchar(191);not null;default:'';comment:'上级查询名'"`
	Page          int64  `gorm:"type:mediumint;not null;default:1;comment:'页码'"`
	Content       string `gorm:"type:longtext;not null;comment:'页面HTML内容'"`
	Category      int64  `gorm:"type:tinyint;not null;comment:'类别'"`
	LastQueryTime int64  `gorm:"not null;default:0;comment:'最后抓取时间'"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

func (Inventory) TableName() string {
	return "d_inventory"
}
