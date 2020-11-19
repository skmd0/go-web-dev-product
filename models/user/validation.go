package user

import (
	"errors"
	"go-web-dev/hash"
	"go-web-dev/rand"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound        = errors.New("models: resource not found")
	ErrInvalidID       = errors.New("models: ID provided is invalid")
	ErrInvalidPassword = errors.New("models: incorrect password provided")
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
