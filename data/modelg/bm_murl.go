package modelg

type Murl struct {
	Id    int64  `gorm:"primary_key"`
	JavId string `gorm:"not null;unique"`
	Name  string `gorm:"not null;index"`

	JacketImg      string `gorm:"not null;type:varchar(300)"`
	JacketImgLocal string `gorm:"not null;type:varchar(300)"`
	SmallImg       string `gorm:"not null;type:varchar(300)"`
	SmallImgLocal  string `gorm:"not null;type:varchar(300)"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const murlTableName = "bm_murl"

func (i *Murl) TableName() string {
	return murlTableName
}
