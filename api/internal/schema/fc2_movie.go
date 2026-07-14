package schema

type FC2Movie struct {
	ID               int64  `gorm:"column:id;primaryKey;autoIncrement"`
	FC2Key           string `gorm:"column:fc2_key;type:varchar(64);not null;default:''"`
	Code             string `gorm:"column:code;type:varchar(32);not null;index:idx_fc2_movie_code"`
	CodeSeq          int64  `gorm:"column:code_seq;not null;default:1"`
	PPVNo            int64  `gorm:"column:ppv_no;not null;index:idx_fc2_movie_ppv_no"`
	Title            string `gorm:"column:title;type:varchar(500);not null;default:''"`
	PerformerText    string `gorm:"column:performer_text;type:varchar(500);not null;default:''"`
	MaskType         string `gorm:"column:mask_type;type:varchar(32);not null;default:''"`
	MetadataThreadID int64  `gorm:"column:metadata_thread_id;not null;default:0;index"`
	PostTime         int64  `gorm:"column:post_time;not null;default:0;index"`
	MediaPresent     int8   `gorm:"column:media_present;type:tinyint;not null;default:0;index:idx_fc2_movie_media_present"`
	MediaBirthTime   int64  `gorm:"column:media_birth_time;not null;default:0;index:idx_fc2_movie_media_birth_time"`
	CoverImageID     int64  `gorm:"column:cover_image_id;not null;default:0;index:idx_fc2_movie_cover_image_id"`
	DataStatus       string `gorm:"column:data_status;type:varchar(32);not null;default:'ready';index"`
	CreatedOn        int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn        int64  `gorm:"column:updated_on;not null;default:0"`
}

func (FC2Movie) TableName() string { return "fc2_movie" }
