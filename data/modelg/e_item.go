// data/modelg/item.go
package modelg

// Item 表：保存从 inventory 抽取出的条目
type Item struct {
	Id     int64  `gorm:"primary_key;autoIncrement"`
	JavId  string `gorm:"type:varchar(191);not null;uniqueIndex:uk_item_jav_id"`
	Name   string `gorm:"type:varchar(191);not null;index:idx_item_name"`
	Prefix string `gorm:"type:varchar(64);not null;index:idx_item_prefix"`

	SearchType int64  `gorm:"type:tinyint;not null;default:0"` // 来源方式: Prefix/Label
	CoverUrl   string `gorm:"type:varchar(512);not null"`
	SearchBy   string `gorm:"type:varchar(191);not null;index:idx_item_searchby"`

	HasDetail        int64 `gorm:"type:tinyint;not null;default:0"` // 明细状态
	HasDownloadCover int64 `gorm:"type:tinyint;not null;default:0"` // 封面下载状态
	HasChinese       int64 `gorm:"type:tinyint;not null;default:0"` // 中文字幕状态
	DetailNeedScan   int64 `gorm:"type:tinyint;not null;default:0"`
	DetailBirthTime  int64 `gorm:"not null;default:0"`
	DetailUpdateTime int64 `gorm:"not null;default:0"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

const itemTableName = "e_item"

func (Item) TableName() string { return itemTableName }
