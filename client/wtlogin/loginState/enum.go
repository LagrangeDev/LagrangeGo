package loginState

type State int

const (
	TokenExpired     State = 140022015
	UnusualVerify          = 140022011
	LoginFailure           = 140022013
	UserTokenExpired       = 140022016
	ServerFailure          = 140022002 //unknown reason
	WrongCaptcha           = 140022007
	WrongArgument          = 140022001
	NewDeviceVerify        = 140022010
	CaptchaVerify          = 140022008
	UnknownError           = -1
	Success                = 0
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
