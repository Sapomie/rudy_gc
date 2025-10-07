package modelg

import (
	"time"

	"gorm.io/gorm"
)

type Album struct {
	Id          int64  `gorm:"primary_key"`
	Name        string `gorm:"not null;unique"`
	FilmNumber  int64  `gorm:"not null"`
	HasUploaded int64  `gorm:"not null;type:tinyint"`
	Size        int64  `gorm:"not null"`
	CreatedOn   int64  `gorm:"not null"`
	UpdatedOn   int64  `gorm:"not null"`
}

const albumTableName = "v_album"

const (
	AlbumHasNotUploaded = 1 + iota
	AlbumHasUploaded
)

func (i *Album) Discrimination() string {
	return i.Name
}

func (i *Album) TableName() string {
	return albumTableName
}

func (i *Album) BeforeCreate(tx *gorm.DB) (err error) {
	i.CreatedOn = time.Now().Unix()
	return
}

func (i *Album) BeforeUpdate(tx *gorm.DB) (err error) {
	i.UpdatedOn = time.Now().Unix()
	return
}
