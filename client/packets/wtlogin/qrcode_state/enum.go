package qrcode_state

type State int

const (
	Confirmed         State = 0
	Expired           State = 17
	WaitingForScan    State = 48
	WaitingForConfirm State = 53
	Canceled          State = 54
)

var statenames = map[State]string{
	Confirmed:         "Confirmed",
	Expired:           "Expired",
	WaitingForScan:    "WaitingForScan",
	WaitingForConfirm: "WaitingForConfirm",
	Canceled:          "Canceled",
}

func (r State) Name() string {
	name, ok := statenames[r]
	if ok {
		return name
	}
	return "Unknown"
}

func (r State) Waitable() bool {
	return r == WaitingForScan || r == WaitingForConfirm
}

func (r State) Success() bool {
	return r == Confirmed
}
