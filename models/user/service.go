package user

import (
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Authenticate(email, password string) (*User, error)
	UserDB
}

func NewUserService(dsn string) (UserService, error) {
	ug, err := newUserGorm(dsn)
	if err != nil {
		return nil, err
	}
	return &userService{newUserValidator(ug)}, nil
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
