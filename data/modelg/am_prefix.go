package modelg

type Prefix struct {
	Id               int64  `gorm:"primary_key"`
	Name             string `gorm:"not null;unique"`
	MovieNumber      int64  `gorm:"not null"`
	OwnedMovieNumber int64  `gorm:"not null"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

const prefixTableName = "am_prefix"

func (i *Prefix) TableName() string {
	return prefixTableName
}
