package loginState

type State int

const (
	TokenExpired     State = 140022015
	UnusualVerify    State = 140022011
	LoginFailure     State = 140022013
	UserTokenExpired State = 140022016
	ServerFailure    State = 140022002 //unknown reason
	WrongCaptcha     State = 140022007
	WrongArgument    State = 140022001
	NewDeviceVerify  State = 140022010
	CaptchaVerify    State = 140022008
	UnknownError     State = -1
	Success          State = 0
)

func (r State) Name() string {
	switch r {
	case TokenExpired:
		return "TokenExpired"
	case UnusualVerify:
		return "UnusualVerify"
	case LoginFailure:
		return "LoginFailure"
	case UserTokenExpired:
		return "UserTokenExpired"
	case ServerFailure:
		return "ServerFailure"
	case WrongCaptcha:
		return "WrongCaptcha"
	case WrongArgument:
		return "WrongArgument"
	case NewDeviceVerify:
		return "NewDeviceVerify"
	case CaptchaVerify:
		return "CaptchaVerify"
	case UnknownError:
		return "UnknownError"
	case Success:
		return "Success"
	}
	return "Unknown"
}

func (r State) Missing() bool {
	return r == UnknownError
}

func (r State) Successful() bool {
	return r == Success
}
