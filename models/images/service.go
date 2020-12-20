package images

import (
	"fmt"
	"gorm.io/gorm"
	"io"
	"os"
	"path/filepath"
)

type ImageService interface {
	ImageDB
	Create(galleryID uint, r io.ReadCloser, filename string) error
	ByGalleryID(galleryID uint) ([]string, error)
}

func NewImageService(db *gorm.DB) ImageService {
	idb := &imageGorm{db}
	iv := &imageValidator{idb}
	return &imageService{iv}
}

type imageService struct {
	ImageDB
}

func (is *imageService) ByGalleryID(galleryID uint) ([]string, error) {
	path := is.imagePath(galleryID)
	strings, err := filepath.Glob(path + "*")
	if err != nil {
		return nil, err
	}
	for i, s := range strings {
		strings[i] = "/" + s
	}
	return strings, nil
}

func (is *imageService) imagePath(galleryID uint) string {
	return fmt.Sprintf("../images/galleries/%v/", galleryID)
}

func (is *imageService) Create(galleryID uint, r io.ReadCloser, filename string) error {
	path, err := is.mkImagePath(galleryID)
	if err != nil {
		return err
	}

	dst, err := os.Create(path + filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, r)
	if err != nil {
		return err
	}

	_ = dst.Close()
	_ = r.Close()
	return nil
}

func (is *imageService) mkImagePath(galleryID uint) (string, error) {
	galleryPath := is.imagePath(galleryID)
	err := os.MkdirAll(galleryPath, 0755)
	if err != nil {
		return "", err
	}
	return galleryPath, nil
}
