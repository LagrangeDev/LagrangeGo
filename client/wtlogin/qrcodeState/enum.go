package qrcodeState

type State int

const (
	Confirmed         State = 0
	Expired                 = 17
	WaitingForScan          = 48
	WaitingForConfirm       = 53
	Canceled                = 54
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
