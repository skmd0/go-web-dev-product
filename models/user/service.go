package user

import (
	"go-web-dev/errs"
	"go-web-dev/hash"
	"go-web-dev/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type UserService interface {
	Authenticate(email, password string) (*User, error)
	InitiateReset(email string) (string, error)
	CompleteReset(token, newPw string) (*User, error)
	UserDB
}

func (us *userService) InitiateReset(email string) (string, error) {
	usr, err := us.ByEmail(email)
	if err != nil {
		return "", err
	}
	pwr := &models.PwReset{UserID: usr.ID}
	if err := us.pwResetDB.Create(pwr); err != nil {
		return "", err
	}
	return pwr.Token, nil
}

func (us *userService) CompleteReset(token, newPw string) (*User, error) {
	pwr, err := us.pwResetDB.ByToken(token)
	if err != nil {
		if err == errs.ErrNotFound {
			return nil, errs.ErrTokenInvalid
		}
		return nil, err
	}
	if time.Now().Sub(pwr.CreatedAt) > (time.Hour * 12) {
		return nil, errs.ErrTokenInvalid
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

func NewUserService(db *gorm.DB, hmacKey, pepper string) UserService {
	ug := &userGorm{db}
	uv := newUserValidator(ug, hmacKey, pepper)
	hmac := hash.NewHMAC(hmacKey)
	pwrGorm := &models.PwResetGorm{DB: db}
	pwr := models.NewPwResetValidator(pwrGorm, hmac)
	return &userService{
		UserDB:    uv,
		pepper:    pepper,
		pwResetDB: pwr,
	}
}

var _ UserService = &userService{}

type userService struct {
	UserDB
	pepper    string
	pwResetDB models.PwResetDB
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
