package modelg

type GScStat struct {
	Id int64 `gorm:"primaryKey;autoIncrement"`

	MovieJavID string `gorm:"column:movie_jav_id;type:varchar(191);not null;uniqueIndex"`
	MovieName  string `gorm:"column:movie_name;type:varchar(191);not null;index:idx_sc_last_name,priority:3,sort:desc;index:idx_come_last_name,priority:3,sort:desc;index:idx_lastsc_name,priority:2,sort:desc"`

	ScTimes        int64 `gorm:"column:sc_times;type:mediumint;not null;index:idx_sc_last_name,priority:1,sort:desc"`
	ComeTimes      int64 `gorm:"column:come_times;type:mediumint;not null;index:idx_come_last_name,priority:1,sort:desc"`
	LastScTime     int64 `gorm:"column:last_sc_time;not null;index:idx_sc_last_name,priority:2,sort:desc;index:idx_come_last_name,priority:2,sort:desc;index:idx_lastsc_name,priority:1,sort:desc"`
	ReleasingDate  int64 `gorm:"column:releasing_date;not null;default:0;index:idx_gss_release_name,priority:1,sort:desc"`
	MediaBirthTime int64 `gorm:"column:media_birth_time;not null;default:0;index:idx_gss_media_birth_name,priority:1,sort:desc"`

	CreatedOn int64 `gorm:"column:created_on;not null;default:0"`
	UpdatedOn int64 `gorm:"column:updated_on;not null;default:0"`
}

const GScStatTableName = "g_sc_stat"

func (GScStat) TableName() string { return GScStatTableName }
