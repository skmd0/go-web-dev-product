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
