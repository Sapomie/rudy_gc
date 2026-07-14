package schema

type FC2MediaImage struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement"`
	FC2MediaID int64  `gorm:"column:fc2_media_id;not null;index:idx_fc2_media_image_media_id"`
	ImageType  string `gorm:"column:image_type;type:varchar(32);not null"`
	FilePath   string `gorm:"column:file_path;type:text;not null"`
	CreatedOn  int64  `gorm:"column:created_on;not null;default:0"`
}

func (FC2MediaImage) TableName() string { return "fc2_media_image" }
