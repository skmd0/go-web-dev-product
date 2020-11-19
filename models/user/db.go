package user

import (
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;uniqueIndex"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;uniqueIndex"`
}

type UserDB interface {
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)

	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error

	AutoMigrate() error
	DestructiveReset() error
}

var _ UserDB = &userGorm{}

type userGorm struct {
	db *gorm.DB
}

func newUserGorm(dsn string) (*userGorm, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &userGorm{
		db: db,
	}, nil
}

func (ug *userGorm) ByID(id uint) (*User, error) {
	return ug.first(ug.db.Where("id = ?", id))
}

func (ug *userGorm) ByEmail(email string) (*User, error) {
	return ug.first(ug.db.Where("email = ?", email))
}

func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	return ug.first(ug.db.Where("remember_hash = ?", rememberHash))
}

func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

func (ug *userGorm) Delete(id uint) error {
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(&user).Error
}

func (ug *userGorm) DestructiveReset() error {
	err := ug.db.Migrator().DropTable(&User{})
	if err != nil {
		return err
	}
	return ug.AutoMigrate()
}

func (ug *userGorm) AutoMigrate() error {
	err := ug.db.Migrator().AutoMigrate(&User{})
	if err != nil {
		return err
	}
	return nil
}

func (ug *userGorm) first(db *gorm.DB) (*User, error) {
	var user User
	err := db.First(&user).Error
	switch err {
	case nil:
		return &user, nil
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
