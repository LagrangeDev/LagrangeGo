package client

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/packets/wtlogin/loginState"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/login"

	"github.com/LagrangeDev/LagrangeGo/utils/crypto"

	"github.com/LagrangeDev/LagrangeGo/info"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
)

var loginLogger = utils.GetLogger("login")

func buildNtloginCaptchaSubmit(ticket, randStr, aid string) proto.DynamicMessage {
	return proto.DynamicMessage{
		1: ticket,
		2: randStr,
		3: aid,
	}
}

func buildNtloginRequest(uin uint32, app *info.AppInfo, device *info.DeviceInfo, sig *info.SigInfo, credential []byte) ([]byte, error) {
	body := proto.DynamicMessage{
		1: proto.DynamicMessage{
			1: proto.DynamicMessage{
				1: strconv.Itoa(int(uin)),
			},
			2: proto.DynamicMessage{
				1: app.OS,
				2: device.DeviceName,
				3: app.NTLoginType,
				4: utils.MustParseHexStr(device.Guid),
			},
			3: proto.DynamicMessage{
				1: device.KernelVersion,
				2: app.AppID,
				3: app.PackageName,
			},
		},
		2: proto.DynamicMessage{
			1: credential,
		},
	}

	if sig.Cookies != "" {
		body[1].(proto.DynamicMessage)[5] = proto.DynamicMessage{1: sig.Cookies}
	}
	if all(sig.CaptchaInfo[:3]) {
		loginLogger.Debugln("login with captcha info")
		body[2].(proto.DynamicMessage)[2] = buildNtloginCaptchaSubmit(sig.CaptchaInfo[0], sig.CaptchaInfo[1], sig.CaptchaInfo[2])
	}

	data, err := crypto.AESGCMEncrypt(body.Encode(), sig.ExchangeKey)
	if err != nil {
		return nil, err
	}

	return proto.DynamicMessage{
		1: sig.KeySig,
		3: data,
		4: 1,
	}.Encode(), nil
}

func ParseNtloginResponse(response []byte, sig *info.SigInfo) (loginState.State, error) {
	var frame login.SsoNTLoginEncryptedData
	err := proto.Unmarshal(response, &frame)
	if err != nil {
		return -1, fmt.Errorf("proto decode failed: %s", err)
	}

	var base login.SsoNTLoginBase
	data, err := crypto.AESGCMDecrypt(frame.GcmCalc, sig.ExchangeKey)
	if err != nil {
		return -1, err
	}
	err = proto.Unmarshal(data, &base)
	if err != nil {
		return -1, fmt.Errorf("proto decode failed: %s", err)
	}
	var body login.SsoNTLoginResponse
	err = proto.Unmarshal(base.Body, &body)
	if err != nil {
		return -1, fmt.Errorf("proto decode failed: %s", err)
	}

	if body.Credentials != nil {
		sig.Tgt = body.Credentials.Tgt
		sig.D2 = body.Credentials.D2
		sig.D2Key = body.Credentials.D2Key
		sig.TempPwd = body.Credentials.TempPassword
		loginLogger.Debugln("SigInfo got")
		return loginState.Success, nil
	} else {
		ret := loginState.State(base.Header.Error.ErrorCode)
		if ret == loginState.CaptchaVerify {
			sig.Cookies = base.Header.Cookie.Cookie.Unwrap()
			verifyUrl := body.Captcha.Url
			aid := getAid(verifyUrl)
			sig.CaptchaInfo[2] = aid
			loginLogger.Warningln("need captcha verify: ", verifyUrl)
		} else if base.Header.Error.Tag != "" {
			stat := base.Header.Error
			title := stat.Tag
			content := stat.Message
			loginLogger.Errorf("Login fail on ntlogin(%s): [%s]>%s", ret.Name(), title, content)
			return -999, fmt.Errorf("login fail on ntlogin(%s): [%s]>%s", ret.Name(), title, content)
		} else {
			loginLogger.Errorf("Login fail: %s", ret.Name())
			return -999, fmt.Errorf("login fail: %s", ret.Name())
		}
	}
	return loginState.State(base.Header.Error.ErrorCode), nil
}

func getAid(verifyUrl string) string {
	u, _ := url.Parse(verifyUrl)
	q := u.Query()
	return q["sid"][0]
}

func all(b []string) bool {
	for _, b1 := range b {
		if b1 == "" {
			return false
		}
	}
	return true
}
