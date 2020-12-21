package views

import (
	"go-web-dev/models/user"
	"log"
)

const (
	AlertLvlSuccess = "primary"
	AlertLvlInfo    = "info"
	AlertLvlWarning = "warning"
	AlertLvlError   = "danger"

	AlertMsgGeneric = "Something went wrong. Please try again later. Contact us if the problem persists."
)

type Alert struct {
	Level   string
	Message string
}

type Data struct {
	Alert *Alert
	User  *user.User
	Yield interface{}
}

func (d *Data) SetAlert(err error) {
	if pubErr, ok := err.(PublicError); ok {
		d.Alert = &Alert{
			Level:   AlertLvlError,
			Message: pubErr.Public(),
		}
	} else {
		log.Println(err)
		d.Alert = &Alert{
			Level:   AlertLvlError,
			Message: AlertMsgGeneric,
		}
	}
}

func (d *Data) SetAlertErr(msg string) {
	log.Println(msg)
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}

type PublicError interface {
	error
	Public() string
}
