package modelg

type Sc struct {
	Id            int64  `gorm:"primary_key"`
	Name          string `gorm:"not null;unique"`
	MovieNumber   int64  `gorm:"not null;type:smallint"`
	ScTime        int64  `gorm:"not null"`
	ComeMovieName string `gorm:"not null;type:varchar(191)"`

	Cooldown          int64  `gorm:"not null;default:0"`
	Duration          int64  `gorm:"not null;default:0"`
	MovieCastPersonID *int64 `gorm:"column:movie_cast_person_id;index"`
	Remarks           string `gorm:"not null;type:varchar(500)"`
	ImagePath         string `gorm:"not null;type:varchar(500)"`

	KindOptionID   *int64 `gorm:"column:kind_option_id;index"`
	FGOptionID     *int64 `gorm:"column:fg_option_id;index"`
	VesselOptionID *int64 `gorm:"column:vessel_option_id;index"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

const ScTableName = "g_sc"

func (i *Sc) TableName() string {
	return ScTableName
}
