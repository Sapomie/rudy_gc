package modelg

type Folder struct {
	ID         int64  `gorm:"primaryKey;autoIncrement;column:id"`
	ParentID   int64  `gorm:"column:parent_id;not null;default:0;index:idx_w_folder_parent_id;uniqueIndex:uk_w_folder_parent_name_source_type,priority:1"` // 根=0
	Name       string `gorm:"column:name;type:varchar(191);not null;uniqueIndex:uk_w_folder_parent_name_source_type,priority:2"`
	SourceType int64  `gorm:"column:source_type;type:tinyint;not null;default:2;index:idx_w_folder_source_type;uniqueIndex:uk_w_folder_parent_name_source_type,priority:3;uniqueIndex:uk_w_folder_path_source_type,priority:2"`
	Depth      int8   `gorm:"column:depth;type:tinyint;not null;index:idx_w_folder_depth"`
	Path       string `gorm:"column:path;type:varchar(512);not null;uniqueIndex:uk_w_folder_path_source_type,priority:1"` // 512 保证可建唯一索引
	PathHash   []byte `gorm:"column:path_hash;type:binary(16);not null;index:idx_w_folder_path_hash"`                     // md5 16B
	CreatedOn  int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn  int64  `gorm:"column:updated_on;not null;default:0"`
}

func (Folder) TableName() string { return "w_folder" }
