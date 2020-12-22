package user

import (
	"go-web-dev/errs"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	Authenticate(email, password string) (*User, error)
	UserDB
}

func NewUserService(db *gorm.DB, hmacKey, pepper string) UserService {
	ug := &userGorm{db}
	uv := newUserValidator(ug, hmacKey, pepper)
	return &userService{
		UserDB: uv,
		pepper: pepper,
	}
}

var _ UserService = &userService{}

type userService struct {
	UserDB
	pepper string
}

func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password+us.pepper))
	switch err {
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, errs.ErrPasswordIncorrect
	case nil:
		return foundUser, nil
	default:
		return nil, err
	}
}
