package modelg

type Director struct {
	Id               int64  `gorm:"primary_key"`
	Name             string `gorm:"not null;unique"`
	JavId            string `gorm:"not null;index"`
	MovieNumber      int64  `gorm:"not null"`
	OwnedMovieNumber int64  `gorm:"not null"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

const directorTableName = "am_director"

func (i *Director) TableName() string {
	return directorTableName
}
