package internal

import (
	"strings"
)

const (
	ErrNotFound          modelError = "models: resource not found"
	ErrPasswordIncorrect modelError = "models: incorrect password provided"
	ErrEmailRequired     modelError = "email address is required"
	ErrEmailInvalid      modelError = "email address is invalid"
	ErrEmailTaken        modelError = "email address is already registered"
	ErrPasswordTooShort  modelError = "password must be at least 8 characters long"
	ErrPasswordRequired  modelError = "password is required"
	ErrTitleRequired     modelError = "gallery need to have a title"
	ErrTokenInvalid      modelError = "models: token provided is not valid"

	ErrInvalidID        privateError = "models: ID provided is invalid"
	ErrRememberTooShort privateError = "remember token is too short"
	ErrRememberRequired privateError = "remember token hash is required"
	ErrUserIDRequired   privateError = "gallery does not have a user specified"
)

type modelError string

func (mErr modelError) Error() string {
	return "models: " + strings.ToLower(string(mErr))
}

func (mErr modelError) Public() string {
	return mErr.Error()
}

type privateError string

func (pe privateError) Error() string {
	return "models: " + strings.ToLower(string(pe))
}
