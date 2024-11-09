package client

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/login"
	"github.com/LagrangeDev/LagrangeGo/client/packets/wtlogin/loginstate"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
)

func buildNtloginCaptchaSubmit(ticket, randStr, aid string) proto.DynamicMessage {
	return proto.DynamicMessage{
		1: ticket,
		2: randStr,
		3: aid,
	}
}

func buildNtloginRequest(uin uint32, app *auth.AppInfo, device *auth.DeviceInfo, sig *auth.SigInfo, credential []byte) ([]byte, error) {
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

func parseNtloginResponse(response []byte, sig *auth.SigInfo) (loginstate.State, error) {
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
		return loginstate.Success, nil
	}
	ret := loginstate.State(base.Header.Error.ErrorCode)
	if ret == loginstate.CaptchaVerify {
		sig.Cookies = base.Header.Cookie.Cookie.Unwrap()
		verifyUrl := body.Captcha.Url
		aid := getAid(verifyUrl)
		sig.CaptchaInfo[2] = aid
		return ret, fmt.Errorf("need captcha verify: %v", verifyUrl)
	}
	if base.Header.Error.Tag != "" {
		stat := base.Header.Error
		title := stat.Tag
		content := stat.Message
		return ret, fmt.Errorf("login fail on ntlogin(%s): [%s]>%s", ret.Name(), title, content)
	}
	return ret, fmt.Errorf("login fail: %s", ret.Name())
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
