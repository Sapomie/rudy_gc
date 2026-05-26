package modelg

type Sc struct {
	Id            int64  `gorm:"primary_key"`
	Name          string `gorm:"not null;unique"`
	MovieNumber   int64  `gorm:"not null;type:smallint"`
	ScTime        int64  `gorm:"not null"`
	Kind          string `gorm:"not null;type:varchar(32);index"`
	ComeMovieName string `gorm:"not null;type:varchar(191)"`

	Cooldown  int64  `gorm:"not null;default:0"`
	Duration  int64  `gorm:"not null;default:0"`
	FG        string `gorm:"not null;type:varchar(191)"`
	Vessel    string `gorm:"not null;type:varchar(191)"`
	MovieCast string `gorm:"not null;type:varchar(191)"`
	Remarks   string `gorm:"not null;type:varchar(500)"`
	ImagePath string `gorm:"not null;type:varchar(500)"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

const ScTableName = "g_sc"

func (i *Sc) TableName() string {
	return ScTableName
}
