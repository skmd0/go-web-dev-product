package user

import (
	"go-web-dev/errs"
	"go-web-dev/hash"
	"go-web-dev/rand"
	"gorm.io/gorm"
)

type PwReset struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	Token     string `gorm:"-"`
	TokenHash string `gorm:"not null;uniqueIndex"`
}

type PwResetDB interface {
	ByToken(token string) (*PwReset, error)
	Create(pwr *PwReset) error
	Delete(id uint) error
}

var _ PwResetDB = &PwResetGorm{}

type PwResetGorm struct {
	DB *gorm.DB
}

func (pv *pwResetValidator) ByToken(token string) (*PwReset, error) {
	pwr := &PwReset{Token: token}
	err := runPwResetValFns(pwr, pv.hmacToken)
	if err != nil {
		return nil, err
	}
	return pv.PwResetDB.ByToken(pwr.TokenHash)
}

func (pv *pwResetValidator) Create(pwr *PwReset) error {
	err := runPwResetValFns(pwr, pv.requireUserID, pv.setTokenIfUnset, pv.hmacToken)
	if err != nil {
		return err
	}
	return pv.PwResetDB.Create(pwr)
}

func (pv *pwResetValidator) Delete(id uint) error {
	if id == 0 {
		return errs.ErrInvalidID
	}
	return pv.PwResetDB.Delete(id)
}

func (pg *PwResetGorm) ByToken(tokenHash string) (*PwReset, error) {
	var pwr *PwReset
	pwr, err := pg.first(pg.DB.Where("token_hash = ?", tokenHash))
	if err != nil {
		return nil, err
	}
	return pwr, nil
}

func (pg *PwResetGorm) Create(pwr *PwReset) error {
	return pg.DB.Create(pwr).Error
}

func (pg *PwResetGorm) Delete(id uint) error {
	var pwr PwReset
	pwr.ID = id
	return pg.DB.Delete(&pwr).Error
}

func (pg *PwResetGorm) first(db *gorm.DB) (*PwReset, error) {
	var pwr PwReset
	err := db.First(&pwr).Error
	switch err {
	case nil:
		return &pwr, nil
	case gorm.ErrRecordNotFound:
		return nil, errs.ErrNotFound
	default:
		return nil, err
	}
}

func NewPwResetValidator(db PwResetDB, hmac hash.HMAC) *pwResetValidator {
	return &pwResetValidator{
		PwResetDB: db,
		hmac:      hmac,
	}
}

func (pv *pwResetValidator) requireUserID(pwr *PwReset) error {
	if pwr.UserID == 0 {
		return errs.ErrUserIDRequired
	}
	return nil
}

func (pv *pwResetValidator) setTokenIfUnset(pwr *PwReset) error {
	if pwr.Token != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	pwr.Token = token
	return nil
}

func (pv *pwResetValidator) hmacToken(pwr *PwReset) error {
	if pwr.Token == "" {
		return nil
	}
	pwr.TokenHash = pv.hmac.Hash(pwr.Token)
	return nil
}

type pwResetValidator struct {
	PwResetDB
	hmac hash.HMAC
}

type pwResetValFn func(*PwReset) error

func runPwResetValFns(pwr *PwReset, fns ...pwResetValFn) error {
	for _, fn := range fns {
		if err := fn(pwr); err != nil {
			return err
		}
	}
	return nil
}
