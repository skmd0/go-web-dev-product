package gallery

import "gorm.io/gorm"

func NewGalleryService(db *gorm.DB) ServiceGallery {
	gg := &galleryGorm{db}
	gv := newGalleryValidator(gg)
	return &galleryService{gv}
}

type ServiceGallery interface {
	TableGallery
}

type galleryService struct {
	TableGallery
}
