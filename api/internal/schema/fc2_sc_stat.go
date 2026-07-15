package schema

type FC2ScStat struct {
	ID int64 `gorm:"column:id;primaryKey;autoIncrement"`

	FC2MovieID int64  `gorm:"column:fc2_movie_id;not null;uniqueIndex:uk_fc2_sc_stat_movie_id"`
	FC2Key     string `gorm:"column:fc2_key;type:varchar(64);not null;default:'';uniqueIndex:uk_fc2_sc_stat_key"`
	MovieName  string `gorm:"column:movie_name;type:varchar(500);not null;default:'';index:idx_fc2ss_sc_selected_name,priority:3,sort:desc;index:idx_fc2ss_come_selected_name,priority:3,sort:desc;index:idx_fc2ss_selected_name,priority:2,sort:desc;index:idx_fc2ss_post_name,priority:2,sort:desc;index:idx_fc2ss_media_birth_name,priority:2,sort:desc"`

	ScEventTimes           int64 `gorm:"column:sc_event_times;type:mediumint;not null;default:0"`
	ScTimes                int64 `gorm:"column:sc_times;type:mediumint;not null;default:0;index:idx_fc2ss_sc_selected_name,priority:1,sort:desc"`
	ComeTimes              int64 `gorm:"column:come_times;type:mediumint;not null;default:0;index:idx_fc2ss_come_selected_name,priority:1,sort:desc"`
	LastSelectedScTime     int64 `gorm:"column:last_selected_sc_time;not null;default:0;index:idx_fc2ss_sc_selected_name,priority:2,sort:desc;index:idx_fc2ss_come_selected_name,priority:2,sort:desc;index:idx_fc2ss_selected_name,priority:1,sort:desc"`
	LastParticipatedScTime int64 `gorm:"column:last_participated_sc_time;not null;default:0"`
	PostTime               int64 `gorm:"column:post_time;not null;default:0;index:idx_fc2ss_post_name,priority:1,sort:desc"`
	MediaBirthTime         int64 `gorm:"column:media_birth_time;not null;default:0;index:idx_fc2ss_media_birth_name,priority:1,sort:desc"`

	CreatedOn int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64 `gorm:"column:updated_on;not null;default:0"`
}

const FC2ScStatTableName = "fc2_sc_stat"

func (FC2ScStat) TableName() string { return FC2ScStatTableName }
