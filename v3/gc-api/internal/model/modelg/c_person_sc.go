package modelg

type PersonSc struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	PersonId    int64  `gorm:"column:person_id;not null;uniqueIndex:uk_c_person_sc_person_sc,priority:1;index:idx_c_person_sc_person_time,priority:1"`
	ScName      string `gorm:"column:sc_name;type:varchar(191);not null;uniqueIndex:uk_c_person_sc_person_sc,priority:2"`
	ScTime      int64  `gorm:"column:sc_time;not null;index:idx_c_person_sc_person_time,priority:2"`
	Cooldown    int64  `gorm:"column:cooldown;not null"`
	MovieCount  int64  `gorm:"column:movie_count;not null"`
	HasCome     int64  `gorm:"column:has_come;type:tinyint;not null"`
	MoviesJson  string `gorm:"column:movies_json;type:longtext;not null"`
	CreatedTime int64  `gorm:"column:created_time;not null"`
	UpdatedTime int64  `gorm:"column:updated_time;not null"`
}

func (PersonSc) TableName() string { return "c_person_sc" }
