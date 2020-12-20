package images

import (
	"gorm.io/gorm"
)

type Image struct {
	gorm.Model
}

type ImageDB interface {
}

type imageGorm struct {
	db *gorm.DB
}
