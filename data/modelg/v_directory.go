package modelg

// v_directory
type Directory struct {
	ID        int64  `gorm:"primaryKey;autoIncrement;column:id"`
	ParentID  int64  `gorm:"column:parent_id;not null;default:0;index:idx_v_directory_parent_id;uniqueIndex:uk_parent_name"` // 根=0
	Name      string `gorm:"column:name;type:varchar(191);not null;uniqueIndex:uk_parent_name"`
	Depth     int8   `gorm:"column:depth;type:tinyint;not null"`
	Path      string `gorm:"column:path;type:varchar(512);not null;uniqueIndex:uk_path"`    // 512 保证可建唯一索引
	PathHash  []byte `gorm:"column:path_hash;type:binary(16);not null;index:idx_path_hash"` // md5 16B
	CreatedOn int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64  `gorm:"column:updated_on;not null;default:0"`
}

func (Directory) TableName() string { return "v_directory" }
