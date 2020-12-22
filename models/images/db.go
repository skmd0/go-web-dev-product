package images

import (
	"fmt"
	"gorm.io/gorm"
	"net/url"
)

type Image struct {
	GalleryID uint
	Filename  string
}

func (i *Image) URLPath() string {
	// by creating url.URL struct the URL will be encoded correctly
	urlPath := url.URL{
		Path: "/" + i.RelativePath(),
	}
	return urlPath.String()
}

func (i *Image) RelativePath() string {
	return fmt.Sprintf("images/galleries/%v/%v", i.GalleryID, i.Filename)
}

type ImageDB interface {
}

type imageGorm struct {
	db *gorm.DB
}
