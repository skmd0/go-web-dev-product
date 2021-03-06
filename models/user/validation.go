package user

import (
	"go-web-dev/internal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"unicode/utf8"
)

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

func newUserValidator(udb TableUser, hmacKey, pepper string) *userValidator {
	return &userValidator{
		TableUser:  udb,
		hmac:       internal.NewHMAC(hmacKey),
		pepper:     pepper,
		emailRegex: regexp.MustCompile("^[\\w-.]+@([\\w-]+\\.)+[\\w-]{2,16}$"),
	}
}

var _ TableUser = &userValidator{}

type userValidator struct {
	TableUser
	hmac       internal.HMAC
	pepper     string
	emailRegex *regexp.Regexp
}

func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := &User{Email: email}
	if err := runUserValFunc(user, uv.requireEmail, uv.normalizeEmail, uv.validEmail); err != nil {
		return nil, err
	}
	return uv.TableUser.ByEmail(email)
}

func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := &User{Remember: token}
	if err := runUserValFunc(user, uv.hmacHashToken); err != nil {
		return nil, err
	}
	return uv.TableUser.ByRemember(user.RememberHash)
}

func (uv *userValidator) Create(user *User) error {
	err := runUserValFunc(user, uv.passRequired, uv.passMinLength, uv.bcryptPassword, uv.passHashRequired,
		uv.hmacGenerateIfMissing, uv.hmacMinBytes, uv.hmacHashToken, uv.hmacHashRequired, uv.requireEmail,
		uv.normalizeEmail, uv.validEmail, uv.emailIsAvailable)
	if err != nil {
		return err
	}
	return uv.TableUser.Create(user)
}

func (uv *userValidator) Update(user *User) error {
	err := runUserValFunc(user, uv.passMinLength, uv.bcryptPassword, uv.hmacMinBytes, uv.hmacHashToken,
		uv.hmacHashRequired, uv.requireEmail, uv.normalizeEmail, uv.validEmail, uv.emailIsAvailable)
	if err != nil {
		return err
	}
	return uv.TableUser.Update(user)
}

func (uv *userValidator) Delete(id uint) error {
	user := &User{Model: gorm.Model{ID: id}}
	if err := runUserValFunc(user, uv.checkUserID); err != nil {
		return err
	}
	return uv.TableUser.Delete(id)
}

func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}
	pwBytes := []byte(user.Password + uv.pepper)
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
	rememberToken, err := internal.GenerateRememberToken(internal.RememberTokenBytes)
	if err != nil {
		return err
	}
	user.Remember = rememberToken
	return nil
}

func (uv *userValidator) hmacMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}
	n, err := internal.NBytes(user.Remember)
	if err != nil {
		return err
	}
	if n < internal.RememberTokenBytes {
		return internal.ErrRememberTooShort
	}
	return nil
}

func (uv *userValidator) hmacHashRequired(user *User) error {
	if user.RememberHash == "" {
		return internal.ErrRememberRequired
	}
	return nil
}

func (uv *userValidator) checkUserID(user *User) error {
	if user.ID == 0 {
		return internal.ErrInvalidID
	}
	return nil
}

func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.TrimSpace(strings.ToLower(user.Email))
	return nil
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return internal.ErrEmailRequired
	}
	return nil
}

func (uv *userValidator) validEmail(user *User) error {
	if !uv.emailRegex.MatchString(user.Email) {
		return internal.ErrEmailInvalid
	}
	return nil
}

func (uv *userValidator) emailIsAvailable(user *User) error {
	existing, err := uv.ByEmail(user.Email)
	if err == internal.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	// we found the user with this email address
	if user.ID != existing.ID {
		return internal.ErrEmailTaken
	}
	return nil
}

func (uv *userValidator) passRequired(user *User) error {
	if user.Password == "" {
		return internal.ErrPasswordRequired
	}
	return nil
}

func (uv *userValidator) passMinLength(user *User) error {
	// when updating user data we don't necessary always want to update the password also if not provided
	if user.Password == "" {
		return nil
	}
	if utf8.RuneCountInString(user.Password) < 8 {
		return internal.ErrPasswordTooShort
	}
	return nil
}

func (uv *userValidator) passHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return internal.ErrPasswordRequired
	}
	return nil
}
