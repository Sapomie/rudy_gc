package schema

type ShForum struct {
	ForumID   string `gorm:"primaryKey;column:forum_id;type:varchar(32);not null"`
	ForumName string `gorm:"column:forum_name;type:varchar(100);not null"`
	SortOrder int64  `gorm:"column:sort_order;not null;default:0;index"`
	IsEnabled int64  `gorm:"column:is_enabled;type:tinyint;not null;default:1;index"`
	CreatedOn int64  `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64  `gorm:"column:updated_on;not null;default:0"`
}

func (ShForum) TableName() string { return "sh_forum" }
