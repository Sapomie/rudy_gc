package modelg

type Kv struct {
	ID          int64  `gorm:"primaryKey;autoIncrement;column:id"`
	ItemKey     string `gorm:"column:item_key;type:varchar(191);not null;uniqueIndex:uk_w_kv_item_key"`
	ItemValue   string `gorm:"column:item_value;type:varchar(191);not null"`
	CreatedTime int64  `gorm:"column:created_time;not null"`
	UpdatedTime int64  `gorm:"column:updated_time;not null"`
}

func (Kv) TableName() string { return "w_kv" }
