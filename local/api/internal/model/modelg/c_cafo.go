package modelg

type Cafo struct {
	Id        int64  `gorm:"primary_key"`
	PersonId  *int64 `gorm:"index:idx_c_cafo_person_id"`
	Name      string `gorm:"not null;unique"`
	Alias     string `gorm:"not null"`
	Chinese   string `gorm:"not null;unique"`
	BirthDay  int64  `gorm:"not null"`
	Height    int64  `gorm:"not null"`
	Cup       string `gorm:"not null;type:varchar(191)"`
	BWH       string `gorm:"not null;type:varchar(191)"`
	Avtar     string `gorm:"not null;type:varchar(191)"`
	Biography string `gorm:"not null;type:text"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const CafoTableName = "c_cafo"

func (i *Cafo) TableName() string {
	return CafoTableName
}
