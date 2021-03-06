package user

import (
	"go-web-dev/internal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type ServiceUser interface {
	Authenticate(email, password string) (*User, error)
	InitiateReset(email string) (string, error)
	CompleteReset(token, newPw string) (*User, error)
	TableUser
}

func (us *userService) InitiateReset(email string) (string, error) {
	usr, err := us.ByEmail(email)
	if err != nil {
		return "", err
	}
	pwr := &PwReset{UserID: usr.ID}
	if err := us.pwResetDB.Create(pwr); err != nil {
		return "", err
	}
	return pwr.Token, nil
}

func (us *userService) CompleteReset(token, newPw string) (*User, error) {
	pwr, err := us.pwResetDB.ByToken(token)
	if err != nil {
		if err == internal.ErrNotFound {
			return nil, internal.ErrTokenInvalid
		}
		return nil, err
	}
	if time.Now().Sub(pwr.CreatedAt) > (time.Hour * 12) {
		return nil, internal.ErrTokenInvalid
	}
	usr, err := us.ByID(pwr.UserID)
	if err != nil {
		return nil, err
	}
	usr.Password = newPw
	err = us.Update(usr)
	if err != nil {
		return nil, err
	}
	_ = us.pwResetDB.Delete(pwr.ID)
	return usr, nil
}

func NewUserService(db *gorm.DB, hmacKey, pepper string) ServiceUser {
	ug := &userGorm{db}
	uv := newUserValidator(ug, hmacKey, pepper)
	hmac := internal.NewHMAC(hmacKey)
	pwrGorm := &PwResetGorm{DB: db}
	pwr := NewPwResetValidator(pwrGorm, hmac)
	return &userService{
		TableUser: uv,
		pepper:    pepper,
		pwResetDB: pwr,
	}
}

var _ ServiceUser = &userService{}

type userService struct {
	TableUser
	pepper    string
	pwResetDB PwResetDB
}

func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password+us.pepper))
	switch err {
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, internal.ErrPasswordIncorrect
	case nil:
		return foundUser, nil
	default:
		return nil, err
	}
}
