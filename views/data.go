package views

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
	Yield interface{}
}

func (d *Data) SetAlert(err error) {
	if pubErr, ok := err.(PublicError); ok {
		d.Alert = &Alert{
			Level:   AlertLvlError,
			Message: pubErr.Public(),
		}
	} else {
		d.Alert = &Alert{
			Level:   AlertLvlError,
			Message: AlertMsgGeneric,
		}
	}
}

func (d *Data) SetAlertErr(msg string) {
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}

type PublicError interface {
	error
	Public() string
}
