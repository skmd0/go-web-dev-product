package gallery

import (
	"go-web-dev/errs"
	"gorm.io/gorm"
)

func newGalleryValidator(gdb GalleryDB) *galleryValidator {
	return &galleryValidator{
		GalleryDB: gdb,
	}
}

type galleryValidator struct {
	GalleryDB
}

func (gv *galleryValidator) Create(gallery *Gallery) error {
	err := runGalleryValFunc(gallery, gv.titleRequired, gv.userIDRequired)
	if err != nil {
		return err
	}
	return gv.GalleryDB.Create(gallery)
}

func (gv *galleryValidator) Update(gallery *Gallery) error {
	err := runGalleryValFunc(gallery, gv.titleRequired, gv.userIDRequired)
	if err != nil {
		return err
	}
	return gv.GalleryDB.Update(gallery)
}

func (gv *galleryValidator) Delete(id uint) error {
	glr := &Gallery{Model: gorm.Model{ID: id}}
	if err := runGalleryValFunc(glr, gv.checkGalleryID); err != nil {
		return err
	}
	return gv.GalleryDB.Delete(id)
}

func (gv *galleryValidator) userIDRequired(gallery *Gallery) error {
	if gallery.UserID <= 0 {
		return errs.ErrUserIDRequired
	}
	return nil
}

func (gv *galleryValidator) titleRequired(gallery *Gallery) error {
	if gallery.Title == "" {
		return errs.ErrTitleRequired
	}
	return nil
}

func (gv *galleryValidator) checkGalleryID(glr *Gallery) error {
	if glr.ID == 0 {
		return errs.ErrInvalidID
	}
	return nil
}

type galleryValFunc func(*Gallery) error

func runGalleryValFunc(gallery *Gallery, fns ...galleryValFunc) error {
	for _, fn := range fns {
		err := fn(gallery)
		if err != nil {
			return err
		}
	}
	return nil
}
