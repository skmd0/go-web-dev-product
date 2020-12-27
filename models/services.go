package models

import (
	"go-web-dev/models/gallery"
	"go-web-dev/models/images"
	"go-web-dev/models/user"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServicesConfig func(*Services) error

func WithImages() ServicesConfig {
	return func(s *Services) error {
		s.Images = images.NewImageService()
		return nil
	}
}

func WithGallery() ServicesConfig {
	return func(s *Services) error {
		s.Gallery = gallery.NewGalleryService(s.db)
		return nil
	}
}

func WithGorm(connectionInfo string) ServicesConfig {
	return func(s *Services) error {
		db, err := gorm.Open(postgres.Open(connectionInfo), &gorm.Config{})
		if err != nil {
			return err
		}
		s.db = db
		return nil
	}
}

func WithUser(hmacKey, pepper string) ServicesConfig {
	return func(s *Services) error {
		s.User = user.NewUserService(s.db, hmacKey, pepper)
		return nil
	}
}

func NewServices(configs ...ServicesConfig) (*Services, error) {
	var s Services
	for _, cfg := range configs {
		if err := cfg(&s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

type Services struct {
	Gallery gallery.ServiceGallery
	Images  images.ImageService
	User    user.ServiceUser
	db      *gorm.DB
}

func (s *Services) DestructiveReset() error {
	err := s.db.Migrator().DropTable(&user.User{}, &gallery.Gallery{}, &user.PwReset{})
	if err != nil {
		return err
	}
	return s.AutoMigrate()
}

func (s *Services) AutoMigrate() error {
	if err := s.db.Migrator().AutoMigrate(&user.User{}, &gallery.Gallery{}, &user.PwReset{}); err != nil {
		return err
	}
	return nil
}
