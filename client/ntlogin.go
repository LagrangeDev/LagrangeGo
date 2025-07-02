package client

import (
	"errors"
	"fmt"
	"strconv"

	ftea "github.com/fumiama/gofastTEA"

	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/login"
	"github.com/LagrangeDev/LagrangeGo/client/packets/wtlogin/loginstate"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	"github.com/LagrangeDev/LagrangeGo/utils/io"
)

func buildPasswordLoginRequest(uin uint32, app *auth.AppInfo, device *auth.DeviceInfo, sig *auth.SigInfo, passwordMD5 [16]byte) ([]byte, error) {
	key := crypto.MD5Digest(binary.NewBuilder().
		WriteBytes(passwordMD5[:]).
		WriteU32(0).
		WriteU32(uin).
		ToBytes(),
	)

	plainBytes := binary.NewBuilder().
		WriteU16(4).
		WriteU32(crypto.RandU32()).
		WriteU32(0).
		WriteU32(uint32(app.AppID)).
		WriteU32(8001).
		WriteU32(0).
		WriteU32(uin).
		WriteU32(uint32(io.TimeStamp())).
		WriteU32(0).
		WriteU8(1).
		WriteBytes(passwordMD5[:]).
		WriteBytes(crypto.RandomBytes(16)).
		WriteU32(0).
		WriteU8(1).
		WriteBytes(io.MustParseHexStr(device.GUID)).
		WriteU32(1).
		WriteU32(1).
		WritePacketString(strconv.Itoa(int(uin)), "u16", false).
		ToBytes()

	encryptedPlain := ftea.NewTeaCipher(key).Encrypt(plainBytes)
	return buildNtloginRequest(uin, app, device, sig, encryptedPlain)
}

func buildNtloginRequest(uin uint32, app *auth.AppInfo, device *auth.DeviceInfo, sig *auth.SigInfo, credential []byte) ([]byte, error) {
	if sig.ExchangeKey == nil || sig.KeySig == nil {
		return nil, errors.New("empty key")
	}

	packet := login.SsoNTLoginBase{
		Header: &login.SsoNTLoginHeader{
			Uin: &login.SsoNTLoginUin{Uin: proto.Some(strconv.Itoa(int(uin)))},
			System: &login.SsoNTLoginSystem{
				OS:         proto.Some(app.OS),
				DeviceName: proto.Some(device.DeviceName),
				Type:       int32(app.NTLoginType),
				Guid:       io.MustParseHexStr(device.GUID),
			},
			Version: &login.SsoNTLoginVersion{
				KernelVersion: proto.Some(device.KernelVersion),
				AppId:         int32(app.AppID),
				PackageName:   proto.Some(app.PackageName),
			},
			Cookie: &login.SsoNTLoginCookie{Cookie: proto.Some(sig.Cookies)},
		},
		Body: func() []byte {
			body := login.SsoNTLoginEasyLogin{
				TempPassword: credential,
			}
			if all(sig.CaptchaInfo[:3]) {
				body.Captcha = &login.SsoNTLoginCaptchaSubmit{
					Ticket:  sig.CaptchaInfo[0],
					RandStr: sig.CaptchaInfo[1],
					Aid:     sig.CaptchaInfo[2],
				}
			}
			b, _ := proto.Marshal(&body)
			return b
		}(),
	}

	pkt, err := proto.Marshal(&packet)
	if err != nil {
		return nil, err
	}

	data, err := crypto.AESGCMEncrypt(pkt, sig.ExchangeKey)
	if err != nil {
		return nil, err
	}

	return proto.Marshal(&login.SsoNTLoginEncryptedData{
		Sign:    sig.KeySig,
		GcmCalc: data,
		Type:    1,
	})
}

func parseNtloginResponse(response []byte, sig *auth.SigInfo) (LoginResponse, error) {
	var frame login.SsoNTLoginEncryptedData
	err := proto.Unmarshal(response, &frame)
	if err != nil {
		return LoginResponse{
			Success: false,
		}, fmt.Errorf("proto decode failed: %w", err)
	}

	var base login.SsoNTLoginBase
	data, err := crypto.AESGCMDecrypt(frame.GcmCalc, sig.ExchangeKey)
	if err != nil {
		return LoginResponse{
			Success: false,
		}, err
	}
	err = proto.Unmarshal(data, &base)
	if err != nil {
		return LoginResponse{
			Success: false,
		}, fmt.Errorf("proto decode failed: %w", err)
	}
	var body login.SsoNTLoginResponse
	err = proto.Unmarshal(base.Body, &body)
	if err != nil {
		return LoginResponse{
			Success: false,
		}, fmt.Errorf("proto decode failed: %w", err)
	}

	if body.Credentials != nil {
		sig.Tgt = body.Credentials.Tgt
		sig.D2 = body.Credentials.D2
		sig.D2Key = body.Credentials.D2Key
		sig.TempPwd = body.Credentials.TempPassword
		return LoginResponse{Success: true}, nil
	}

	ret := loginstate.State(base.Header.Error.ErrorCode)
	if base.Header.Error != nil && ret.NeedVerify() {
		sig.UnusualSig = func() []byte {
			if body.Unusual != nil {
				return body.Unusual.Sig
			}
			return nil
		}()
		sig.Cookies = base.Header.Cookie.Cookie.Unwrap()
		sig.NewDeviceVerifyURL = base.Header.Error.NewDeviceVerifyUrl.Unwrap()
		err = nil
	}
	if base.Header.Error.Tag != "" {
		stat := base.Header.Error
		title := stat.Tag
		content := stat.Message
		err = fmt.Errorf("login fail on ntlogin(%s): [%s]>%s", ret.Name(), title, content)
	}
	var loginErr LoginError
	var verifyURL string
	//nolint:exhaustive
	switch ret {
	case loginstate.CaptchaVerify:
		loginErr = SliderNeededError
		verifyURL = body.Captcha.Url
	case loginstate.NewDeviceVerify:
		loginErr = UnsafeDeviceError
		verifyURL = base.Header.Error.NewDeviceVerifyUrl.Unwrap()
	default:
		loginErr = OtherLoginError
	}
	return LoginResponse{
		Success:      false,
		Error:        loginErr,
		ErrorMessage: ret.Name(),
		VerifyURL:    verifyURL,
	}, err
}

func parseNewDeviceLoginResponse(response []byte, sig *auth.SigInfo) (loginstate.State, error) {
	if len(sig.ExchangeKey) == 0 {
		return -999, errors.New("empty exchange key")
	}

	var encrypted login.SsoNTLoginEncryptedData
	err := proto.Unmarshal(response, &encrypted)
	if err != nil {
		return -999, err
	}

	if encrypted.GcmCalc != nil {
		decrypted, err := crypto.AESGCMDecrypt(encrypted.GcmCalc, sig.ExchangeKey)
		if err != nil {
			return -999, err
		}
		var base login.SsoNTLoginBase
		err = proto.Unmarshal(decrypted, &base)
		if err != nil {
			return -999, err
		}
		var body login.SsoNTLoginResponse
		err = proto.Unmarshal(base.Body, &body)
		if err != nil {
			return -999, err
		}
		ret := loginstate.State(base.Header.Error.ErrorCode)
		if body.Credentials == nil {
			return ret, errors.New(ret.Name())
		}
		sig.Tgt = body.Credentials.Tgt
		sig.D2 = body.Credentials.D2
		sig.D2Key = body.Credentials.D2Key
		sig.TempPwd = body.Credentials.TempPassword
		return ret, nil
	}
	return -999, errors.New("empty data")
}

func all(b []string) bool {
	for _, b1 := range b {
		if b1 == "" {
			return false
		}
	}
	return true
}
