package models

import (
	"errors"
	"go-web-dev/hash"
	"go-web-dev/rand"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ErrNotFound        = errors.New("models: resource not found")
	ErrInvalidID       = errors.New("models: ID provided is invalid")
	ErrInvalidPassword = errors.New("models: incorrect password provided")
	//ErrInvalidEmail = errors.New("models: invalid email address provided")
)

const userPwPepper = "secret-user-pepper-string"
const hmacSecretKey = "my-hmac-secret-key"

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

type UserService interface {
	Authenticate(email, password string) (*User, error)
	UserDB
}

func NewUserService(dsn string) (UserService, error) {
	ug, err := newUserGorm(dsn)
	if err != nil {
		return nil, err
	}
	hmac := hash.NewHMAC(hmacSecretKey)
	return &userService{
		&userValidator{
			ug,
			hmac,
		},
	}, nil
}

var _ UserService = &userService{}

type userService struct {
	UserDB
}

func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password+userPwPepper))
	switch err {
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrInvalidPassword
	case nil:
		return foundUser, nil
	default:
		return nil, err
	}
}

type userValFunc func(*User) error

func runUserValFunc(user *User, fns ...userValFunc) error {
	for _, fn := range fns {
		err := fn(user)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ UserDB = &userValidator{}

type userValidator struct {
	UserDB
	hmac hash.HMAC
}

func (uv *userValidator) ByRemember(token string) (*User, error) {
	rememberHash := uv.hmac.Hash(token)
	return uv.UserDB.ByRemember(rememberHash)
}

func (uv *userValidator) Create(user *User) error {
	if err := runUserValFunc(user, uv.bcryptPassword); err != nil {
		return err
	}
	if user.Remember == "" {
		rememberToken, err := rand.GenerateRememberToken(rand.RememberTokenBytes)
		if err != nil {
			return err
		}
		user.Remember = rememberToken
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return uv.UserDB.Create(user)
}

func (uv *userValidator) Update(user *User) error {
	if err := runUserValFunc(user, uv.bcryptPassword); err != nil {
		return err
	}
	// only change RememberHash when there is a new Remember token
	if user.Remember != "" {
		user.RememberHash = uv.hmac.Hash(user.Remember)
	}
	return uv.Update(user)
}

func (uv *userValidator) Delete(id uint) error {
	if id <= 0 {
		return ErrInvalidID
	}
	return uv.Delete(id)
}

func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}
	pwBytes := []byte(user.Password + userPwPepper)
	hashedPassword, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)
	user.Password = ""
	return nil
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
