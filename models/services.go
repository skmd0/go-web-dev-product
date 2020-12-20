package models

import (
	"go-web-dev/models/gallery"
	"go-web-dev/models/images"
	"go-web-dev/models/user"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewServices(connectionInfo string) (*Services, error) {
	db, err := gorm.Open(postgres.Open(connectionInfo), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Services{
		Gallery: gallery.NewGalleryService(db),
		Images:  images.NewImageService(db),
		User:    user.NewUserService(db),
		db:      db,
	}, nil
}

type Services struct {
	Gallery gallery.GalleryService
	Images  images.ImageService
	User    user.UserService
	db      *gorm.DB
}

func (s *Services) DestructiveReset() error {
	err := s.db.Migrator().DropTable(&user.User{}, &gallery.Gallery{})
	if err != nil {
		return err
	}
	return s.AutoMigrate()
}

func (s *Services) AutoMigrate() error {
	if err := s.db.Migrator().AutoMigrate(&user.User{}, &gallery.Gallery{}); err != nil {
		return err
	}
	return nil
}
