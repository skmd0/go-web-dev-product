package gallery

import "gorm.io/gorm"

func NewGalleryService(db *gorm.DB) GalleryService {
	gg := &galleryGorm{db}
	gv := newGalleryValidator(gg)
	return &galleryService{gv}
}

type GalleryService interface {
	GalleryDB
}

type galleryService struct {
	GalleryDB
}
