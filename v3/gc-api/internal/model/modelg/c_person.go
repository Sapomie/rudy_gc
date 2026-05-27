package modelg

type Person struct {
	Id      int64  `gorm:"primaryKey;autoIncrement"`
	Name    string `gorm:"not null;type:varchar(191);index"`
	Alias   string `gorm:"not null;type:text"`
	Chinese string `gorm:"not null;type:varchar(191);index"`

	BirthDay int64  `gorm:"not null;index"`
	Height   int64  `gorm:"not null;index"`
	Cup      string `gorm:"not null;type:varchar(191)"`
	Bwh      string `gorm:"not null;type:varchar(191)"`
	Avatar   string `gorm:"not null;type:varchar(191)"`

	MovieNumber       int64 `gorm:"not null;index"`
	OwnedMovieNumber  int64 `gorm:"not null;index"`
	OwnedWMediaNumber int64 `gorm:"not null;index"`
	OwnedWMediaRatio  int64 `gorm:"not null;default:0;index"`
	ScTimes           int64 `gorm:"not null;type:MEDIUMINT;index"`
	ComeTimes         int64 `gorm:"not null;type:MEDIUMINT;index"`
	LastScTime        int64 `gorm:"not null;index"`
	LastScEventTime   int64 `gorm:"not null;index"`
	HighestRank       int64 `gorm:"not null;type:MEDIUMINT;index"`
	RankTimes         int64 `gorm:"not null;type:MEDIUMINT;index"`

	CreatedOn int64 `gorm:"not null;default:0"`
	UpdatedOn int64 `gorm:"not null;default:0"`
}

const PersonTableName = "c_person"

func (i *Person) TableName() string {
	return PersonTableName
}
