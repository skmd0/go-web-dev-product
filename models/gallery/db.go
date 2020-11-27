package gallery

import (
	"go-web-dev/errs"
	"gorm.io/gorm"
)

type Gallery struct {
	gorm.Model
	UserID uint   `gorm:"not null;index"`
	Title  string `gorm:"not null"`
}

type GalleryDB interface {
	ByID(id uint) (*Gallery, error)
	Create(gallery *Gallery) error
	Update(gallery *Gallery) error
	Delete(id uint) error
}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}

func (gg *galleryGorm) Update(gallery *Gallery) error {
	return gg.db.Save(gallery).Error
}

func (gg *galleryGorm) Delete(id uint) error {
	glr := Gallery{Model: gorm.Model{ID: id}}
	return gg.db.Delete(&glr).Error
}

func (gg *galleryGorm) ByID(id uint) (*Gallery, error) {
	return gg.first(gg.db.Where("id = ?", id))
}

func (gg *galleryGorm) first(db *gorm.DB) (*Gallery, error) {
	var gallery Gallery
	err := db.First(&gallery).Error
	switch err {
	case nil:
		return &gallery, nil
	case gorm.ErrRecordNotFound:
		return nil, errs.ErrNotFound
	default:
		return nil, err
	}
}
