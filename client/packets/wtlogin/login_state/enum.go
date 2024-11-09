package login_state

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

var statenames = map[State]string{
	TokenExpired:     "TokenExpired",
	UnusualVerify:    "UnusualVerify",
	LoginFailure:     "LoginFailure",
	UserTokenExpired: "UserTokenExpired",
	ServerFailure:    "ServerFailure",
	WrongCaptcha:     "WrongCaptcha",
	WrongArgument:    "WrongArgument",
	NewDeviceVerify:  "NewDeviceVerify",
	CaptchaVerify:    "CaptchaVerify",
	UnknownError:     "UnknownError",
	Success:          "Success",
}

func (r State) Name() string {
	name, ok := statenames[r]
	if ok {
		return name
	}
	return "Unknown"
}

func (r State) Missing() bool {
	return r == UnknownError
}

func (r State) Successful() bool {
	return r == Success
}
