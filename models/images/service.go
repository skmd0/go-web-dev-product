package images

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ImageService interface {
	Create(galleryID uint, r io.ReadCloser, filename string) error
	ByGalleryID(galleryID uint) ([]Image, error)
	Delete(image *Image) error
}

func NewImageService() ImageService {
	return &imageService{}
}

type imageService struct {
}

func (is *imageService) Delete(image *Image) error {
	return os.Remove("../" + image.RelativePath())
}

func (is *imageService) ByGalleryID(galleryID uint) ([]Image, error) {
	path := is.imagePath(galleryID)
	files, err := filepath.Glob(path + "*")
	if err != nil {
		return nil, err
	}
	ret := make([]Image, len(files))
	for i, s := range files {
		s = strings.Replace(s, path, "", 1)
		ret[i] = Image{
			GalleryID: galleryID,
			Filename:  s,
		}
	}
	return ret, nil
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
