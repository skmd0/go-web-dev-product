package models

import "strings"

const (
	ErrNotFound          modelError = "models: resource not found"
	ErrInvalidID         modelError = "models: ID provided is invalid"
	ErrPasswordIncorrect modelError = "models: incorrect password provided"
	ErrEmailRequired     modelError = "email address is required"
	ErrEmailInvalid      modelError = "email address is invalid"
	ErrEmailTaken        modelError = "email address is already registered"
	ErrPasswordTooShort  modelError = "password must be at least 8 characters long"
	ErrPasswordRequired  modelError = "password is required"
	ErrRememberTooShort  modelError = "remember token is too short"
	ErrRememberRequired  modelError = "remember token hash is required"
)

type modelError string

func (me modelError) Error() string {
	return "models: " + strings.ToLower(me.Error())
}

func (me modelError) Public() string {
	return me.Error()
}
