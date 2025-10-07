package modelg

type Murl struct {
	Id       int64  `gorm:"primary_key"`
	MurlName string `gorm:"not null;index"`
	JavId    string `gorm:"not null;unique"`

	JacketImg      string `gorm:"not null;type:varchar(300)"`
	JacketImgLocal string `gorm:"not null;type:varchar(300)"`
	SmallImg       string `gorm:"not null;type:varchar(300)"`
	SmallImgLocal  string `gorm:"not null;type:varchar(300)"`
	FilmUrl        string `gorm:"not null;type:varchar(300)"`
	EncodeName     string `gorm:"not null;type:varchar(191)"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const murlTableName = "bm_murl"

func (i *Murl) TableName() string {
	return murlTableName
}
