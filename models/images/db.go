package images

import (
	"fmt"
	"gorm.io/gorm"
)

type Image struct {
	GalleryID uint
	Filename  string
}

func (i *Image) URLPath() string {
	return "/" + i.RelativePath()
}

func (i *Image) RelativePath() string {
	return fmt.Sprintf("images/galleries/%v/%v", i.GalleryID, i.Filename)
}

type ImageDB interface {
}

type imageGorm struct {
	db *gorm.DB
}
