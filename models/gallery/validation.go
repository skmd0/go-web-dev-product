package gallery

import "go-web-dev/errs"

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
