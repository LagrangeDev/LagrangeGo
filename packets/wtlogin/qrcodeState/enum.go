package qrcodeState

type State int

const (
	Confirmed         State = 0
	Expired           State = 17
	WaitingForScan    State = 48
	WaitingForConfirm State = 53
	Canceled          State = 54
)

func (r State) Name() string {
	switch r {
	case Confirmed:
		return "Confirmed"
	case Expired:
		return "Expired"
	case WaitingForScan:
		return "WaitingForScan"
	case WaitingForConfirm:
		return "WaitingForConfirm"
	case Canceled:
		return "Canceled"
	}
	return "Unknown"
}

func (r State) Waitable() bool {
	return r == WaitingForScan || r == WaitingForConfirm
}

func (r State) Success() bool {
	return r == Confirmed
}
