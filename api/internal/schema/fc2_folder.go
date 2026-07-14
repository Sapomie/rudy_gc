package schema

type FC2Folder struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement"`
	ParentID   int64  `gorm:"column:parent_id;not null;default:0;index:idx_fc2_folder_parent_id;uniqueIndex:uk_fc2_folder_parent_name_source_type,priority:1"`
	Name       string `gorm:"column:name;type:varchar(191);not null;uniqueIndex:uk_fc2_folder_parent_name_source_type,priority:2"`
	SourceType int64  `gorm:"column:source_type;type:tinyint;not null;default:2;index:idx_fc2_folder_source_type;uniqueIndex:uk_fc2_folder_parent_name_source_type,priority:3;uniqueIndex:uk_fc2_folder_path_source_type,priority:2"`
	Depth      int8   `gorm:"column:depth;type:tinyint;not null;index:idx_fc2_folder_depth"`
	Path       string `gorm:"column:path;type:varchar(512);not null;uniqueIndex:uk_fc2_folder_path_source_type,priority:1"`
	PathHash   []byte `gorm:"column:path_hash;type:binary(16);not null;index:idx_fc2_folder_path_hash"`
	CreatedOn  int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn  int64  `gorm:"column:updated_on;not null;default:0"`
}

func (FC2Folder) TableName() string { return "fc2_folder" }
