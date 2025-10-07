package modelg

import (
	"time"

	"gorm.io/gorm"
)

type Cafo struct {
	Id       int64   `gorm:"primary_key"`
	Name     string  `gorm:"not null;unique"`
	Alias    string  `gorm:"not null"`
	Chinese  string  `gorm:"not null;unique"`
	BirthDay int64   `gorm:"not null"`
	Height   float64 `gorm:"not null"`
	Cup      string  `gorm:"not null;type:varchar(191)"`
	BWH      string  `gorm:"not null;type:varchar(191)"`
	Avtar    string  `gorm:"not null;type:varchar(191)"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const CafoTableName = "c_cafo"

func (i *Cafo) Discrimination() string {
	return i.Name
}

func (i *Cafo) TableName() string {
	return CafoTableName
}

func (i *Cafo) BeforeCreate(tx *gorm.DB) (err error) {
	i.CreatedOn = time.Now().Unix()
	return
}

func (i *Cafo) BeforeUpdate(tx *gorm.DB) (err error) {
	i.UpdatedOn = time.Now().Unix()
	return
}
