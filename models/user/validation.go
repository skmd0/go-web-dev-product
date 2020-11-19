package user

import (
	"errors"
	"go-web-dev/hash"
	"go-web-dev/rand"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"regexp"
	"strings"
)

var (
	ErrNotFound        = errors.New("models: resource not found")
	ErrInvalidID       = errors.New("models: ID provided is invalid")
	ErrInvalidPassword = errors.New("models: incorrect password provided")
	ErrEmailRequired   = errors.New("email address is required")
	ErrEmailInvalid    = errors.New("email address is invalid")
	ErrEmailTaken      = errors.New("email address is already registered")
	//ErrInvalidEmail = errors.New("models: invalid email address provided")
)

const userPwPepper = "secret-user-pepper-string"
const hmacSecretKey = "my-hmac-secret-key"

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

func newUserValidator(udb UserDB) *userValidator {
	return &userValidator{
		UserDB:     udb,
		hmac:       hash.NewHMAC(hmacSecretKey),
		emailRegex: regexp.MustCompile("^[\\w-.]+@([\\w-]+\\.)+[\\w-]{2,16}$"),
	}
}

var _ UserDB = &userValidator{}

type userValidator struct {
	UserDB
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
}

func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := &User{Email: email}
	if err := runUserValFunc(user, uv.requireEmail, uv.normalizeEmail, uv.validEmail); err != nil {
		return nil, err
	}
	return uv.UserDB.ByEmail(email)
}

func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := &User{Remember: token}
	if err := runUserValFunc(user, uv.hmacHashToken); err != nil {
		return nil, err
	}
	return uv.UserDB.ByRemember(user.RememberHash)
}

func (uv *userValidator) Create(user *User) error {
	err := runUserValFunc(user, uv.bcryptPassword, uv.hmacGenerateIfMissing, uv.hmacHashToken, uv.requireEmail,
		uv.normalizeEmail, uv.validEmail, uv.emailIsAvailable)
	if err != nil {
		return err
	}
	return uv.UserDB.Create(user)
}

func (uv *userValidator) Update(user *User) error {
	err := runUserValFunc(user, uv.bcryptPassword, uv.hmacHashToken, uv.requireEmail, uv.normalizeEmail, uv.validEmail,
		uv.emailIsAvailable)
	if err != nil {
		return err
	}
	return uv.Update(user)
}

func (uv *userValidator) Delete(id uint) error {
	user := &User{Model: gorm.Model{ID: id}}
	if err := runUserValFunc(user, uv.checkUserID); err != nil {
		return err
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

func (uv *userValidator) hmacHashToken(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}
func (uv *userValidator) hmacGenerateIfMissing(user *User) error {
	if user.Remember != "" {
		return nil
	}
	rememberToken, err := rand.GenerateRememberToken(rand.RememberTokenBytes)
	if err != nil {
		return err
	}
	user.Remember = rememberToken
	return nil
}

func (uv *userValidator) checkUserID(user *User) error {
	if user.ID == 0 {
		return ErrInvalidID
	}
	return nil
}

func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.TrimSpace(strings.ToLower(user.Email))
	return nil
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func (uv *userValidator) validEmail(user *User) error {
	if !uv.emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
	}
	return nil
}

func (uv *userValidator) emailIsAvailable(user *User) error {
	_, err := uv.ByEmail(user.Email)
	if err == ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	return ErrEmailTaken
}
