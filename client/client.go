package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/packets/tlv"
	"github.com/LagrangeDev/LagrangeGo/client/packets/wtlogin"
	"github.com/LagrangeDev/LagrangeGo/client/packets/wtlogin/qrcodestate"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

func (c *QQClient) TokenLogin() (*LoginResponse, error) {
	if c.Online.Load() {
		return nil, ErrAlreadyOnline
	}
	data, err := buildNtloginRequest(c.Uin, c.Version(), c.Device(), &c.transport.Sig, c.transport.Sig.TempPwd)
	if err != nil {
		return nil, err
	}
	packet, err := c.sendUniPacketAndWait(
		"trpc.login.ecdh.EcdhService.SsoNTLoginEasyLogin",
		data,
	)
	if err != nil {
		return nil, err
	}
	res, err := parseNtloginResponse(packet, &c.transport.Sig)
	if err != nil {
		return nil, err
	}
	if res.Success {
		err = c.init()
	}
	return &res, err
}

func (c *QQClient) Relogin() (*LoginResponse, error) {
	if !c.Online.Load() {
		return nil, ErrNotOnline
	}

	c.Disconnect()
	if err := c.connect(); err != nil {
		return nil, err
	}

	if err := c.keyExchange(); err != nil {
		return nil, err
	}

	return c.TokenLogin()
}

func (c *QQClient) FetchQRCodeDefault() ([]byte, string, error) {
	return c.FetchQRCode(3, 4, 2)
}

func (c *QQClient) FetchQRCode(size, margin, ecLevel uint32) ([]byte, string, error) {
	if c.Online.Load() {
		return nil, "", ErrAlreadyOnline
	}
	err := c.connect()
	if err != nil {
		return nil, "", err
	}

	packet := c.buildCode2dPacket(c.Uin, 0x31, binary.NewBuilder().
		WriteU16(0).
		WriteU64(0).
		WriteU8(0).
		WriteTLV(
			tlv.T16(c.Version().AppID, c.Version().SubAppID,
				utils.MustParseHexStr(c.Device().GUID), c.Version().PTVersion, c.Version().PackageName),
			tlv.T1b(0, 0, size, margin, 72, ecLevel, 2),
			tlv.T1d(c.Version().MiscBitmap),
			tlv.T33(utils.MustParseHexStr(c.Device().GUID)),
			tlv.T35(c.Version().PTOSVersion),
			tlv.T66(c.Version().PTOSVersion),
			tlv.Td1(c.Version().OS, c.Device().DeviceName),
		).WriteU8(3).ToBytes())

	response, err := c.sendUniPacketAndWait("wtlogin.trans_emp", packet)
	if err != nil {
		return nil, "", err
	}

	decrypted := binary.NewReader(response)
	decrypted.SkipBytes(54)
	retCode := decrypted.ReadU8()
	qrsig := decrypted.ReadBytesWithLength("u16", false)
	tlvs := decrypted.ReadTlv()

	if retCode == 0 && tlvs[0x17] != nil {
		c.transport.Sig.Qrsig = qrsig
		urlreader := binary.NewReader(tlvs[209])
		// 这样是不对的，调试后发现应该丢一个字节，然后读下一个字节才是数据的大小
		// string(binary.NewReader(tlvs[209]).ReadBytesWithLength("u16", true))
		urlreader.ReadU8()
		return tlvs[0x17], utils.B2S(urlreader.ReadBytesWithLength("u8", false)), nil
	}

	return nil, "", fmt.Errorf("err qr retcode %d", retCode)
}

func (c *QQClient) GetQRCodeResult() (qrcodestate.State, error) {
	c.debugln("get qrcode result")
	if c.transport.Sig.Qrsig == nil {
		return -1, errors.New("no qrsig found, execute fetch_qrcode first")
	}

	response, err := c.sendUniPacketAndWait("wtlogin.trans_emp",
		c.buildCode2dPacket(0, 0x12, binary.NewBuilder().
			WritePacketBytes(c.transport.Sig.Qrsig, "u16", false).
			WriteU64(0).
			WriteU32(0).
			WriteU8(0).
			WriteU8(0x83).ToBytes()))
	if err != nil {
		return -1, err
	}

	reader := binary.NewReader(response)
	//length := reader.ReadU32()
	reader.SkipBytes(8) // 4 + 4
	reader.ReadU16()    // cmd, 0x12
	reader.SkipBytes(40)
	_ = reader.ReadU32() // app id
	retCode := qrcodestate.State(reader.ReadU8())

	if retCode == 0 {
		reader.SkipBytes(4)
		c.Uin = reader.ReadU32()
		reader.SkipBytes(4)
		t := reader.ReadTlv()
		c.t106 = t[0x18]
		c.t16a = t[0x19]
		c.transport.Sig.Tgtgt = t[0x1e]
		c.debugln("len(c.t106) =", len(c.t106), "len(c.t16a) =", len(c.t16a))
		c.debugln("len(c.transport.Sig.Tgtgt) =", len(c.transport.Sig.Tgtgt))
	}

	return retCode, nil
}

func (c *QQClient) keyExchange() error {
	data, err := wtlogin.BuildKexExchangeRequest(c.Uin, c.Device().GUID)
	if err != nil {
		return err
	}
	packet, err := c.sendUniPacketAndWait(
		"trpc.login.ecdh.EcdhService.SsoKeyExchange",
		data,
	)
	if err != nil {
		return err
	}
	return wtlogin.ParseKeyExchangeResponse(packet, c.Sig())
}

func (c *QQClient) PasswordLogin() (*LoginResponse, error) {
	if c.Online.Load() {
		return nil, ErrAlreadyOnline
	}

	err := c.connect()
	if err != nil {
		return nil, err
	}

	err = c.keyExchange()
	if err != nil {
		return nil, err
	}

	data, err := buildPasswordLoginRequest(c.Uin, c.Version(), c.Device(), &c.transport.Sig, c.PasswordMD5)
	if err != nil {
		return nil, err
	}
	packet, err := c.sendUniPacketAndWait(
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLogin",
		data,
	)
	if err != nil {
		return nil, err
	}
	res, err := parseNtloginResponse(packet, &c.transport.Sig)
	if err != nil {
		return nil, err
	}
	if res.Success {
		err = c.init()
	}
	return &res, err
}

func (c *QQClient) SubmitCaptcha(ticket, randStr, aid string) (*LoginResponse, error) {
	c.Sig().CaptchaInfo = [3]string{ticket, randStr, aid}
	data, err := buildPasswordLoginRequest(c.Uin, c.Version(), c.Device(), &c.transport.Sig, c.PasswordMD5)
	if err != nil {
		return nil, err
	}
	packet, err := c.sendUniPacketAndWait(
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLogin",
		data,
	)
	if err != nil {
		return nil, err
	}
	res, err := parseNtloginResponse(packet, &c.transport.Sig)
	if err != nil {
		return nil, err
	}
	if res.Success {
		err = c.init()
	}
	return &res, err
}

func (c *QQClient) GetNewDeviceVerifyURL() (string, error) {
	if c.Sig().NewDeviceVerifyURL == "" {
		return "", errors.New("no verify_url found")
	}

	queryParams := func() url.Values {
		parsedURL, err := url.Parse(c.Sig().NewDeviceVerifyURL)
		if err != nil {
			return url.Values{}
		}
		return parsedURL.Query()
	}()

	request, _ := json.Marshal(&NTNewDeviceQrCodeRequest{
		StrDevAuthToken: queryParams.Get("sig"),
		Uint32Flag:      1,
		Uint32UrlType:   0,
		StrUinToken:     queryParams.Get("uin-token"),
		StrDevType:      c.Version().OS,
		StrDevName:      c.Device().DeviceName,
	})

	resp, err := func() (*NTNewDeviceQrCodeResponse, error) {
		resp, err := http.Post(fmt.Sprintf("https://oidb.tim.qq.com/v3/oidbinterface/oidb_0xc9e_8?uid=%d&getqrcode=1&sdkappid=39998&actype=2", c.Uin),
			"application/json", bytes.NewReader(request))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		var response NTNewDeviceQrCodeResponse
		err = json.Unmarshal(data, &response)
		if err != nil {
			return nil, err
		}
		return &response, nil
	}()
	if err != nil {
		return "", err
	}
	return resp.StrURL, nil
}

func (c *QQClient) NewDeviceVerify(verifyURL string) error {
	original := func() string {
		params := strings.FieldsFunc(strings.Split(verifyURL, "?")[1], func(r rune) bool {
			return r == '&'
		})
		for _, param := range params {
			if strings.HasPrefix(param, "str_url=") {
				return strings.TrimPrefix(param, "str_url=")
			}
		}
		return ""
	}()

	query := func() []byte {
		data, _ := json.Marshal(&NTNewDeviceQrCodeQuery{
			Uint32Flag: 0,
			Token:      base64.StdEncoding.EncodeToString(utils.S2B(original)),
		})
		return data
	}()
	for errCount, timeSec := 0, 0; timeSec < 120 || errCount < 3; timeSec++ {
		resp, err := func() (*NTNewDeviceQrCodeResponse, error) {
			resp, err := http.Post(fmt.Sprintf("https://oidb.tim.qq.com/v3/oidbinterface/oidb_0xc9e_8?uid=%d&getqrcode=1&sdkappid=39998&actype=2", c.Uin),
				"application/json", bytes.NewReader(query))
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			data, _ := io.ReadAll(resp.Body)
			var response NTNewDeviceQrCodeResponse
			err = json.Unmarshal(data, &response)
			if err != nil {
				return nil, err
			}
			return &response, nil
		}()
		if err != nil {
			errCount++
			continue
		}
		if resp.StrNtSuccToken != "" {
			c.transport.Sig.TempPwd = utils.S2B(resp.StrNtSuccToken)
			data, err := buildNtloginRequest(c.Uin, c.Version(), c.Device(), &c.transport.Sig, c.transport.Sig.TempPwd)
			if err != nil {
				return err
			}

			packet, err := c.sendUniPacketAndWait(
				"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLoginNewDevice",
				data,
			)
			if err != nil {
				return err
			}
			ret, err := parseNewDeviceLoginResponse(packet, &c.transport.Sig)
			if err != nil {
				return err
			}
			if ret.Successful() {
				return c.init()
			}
			return fmt.Errorf("verify error: %s", ret.Name())
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("verify timeout error")
}

func (c *QQClient) QRCodeLogin() (*LoginResponse, error) {
	app := c.Version()
	device := c.Device()
	response, err := c.sendUniPacketAndWait(
		"wtlogin.login",
		c.buildLoginPacket(c.Uin, "wtlogin.login", binary.NewBuilder().
			WriteU16(0x09).
			WriteTLV(
				binary.NewBuilder().WriteBytes(c.t106).Pack(0x106),
				tlv.T144(c.transport.Sig.Tgtgt, app, device),
				tlv.T116(app.SubSigmap),
				tlv.T142(app.PackageName, 0),
				tlv.T145(utils.MustParseHexStr(device.GUID)),
				tlv.T18(0, app.AppClientVersion, int(c.Uin), 0, 5, 0),
				tlv.T141([]byte("Unknown"), nil),
				tlv.T177(app.WTLoginSDK, 0),
				tlv.T191(0),
				tlv.T100(5, app.AppID, app.SubAppID, 8001, app.MainSigmap, 0),
				tlv.T107(1, 0x0d, 0, 1),
				tlv.T318(nil),
				binary.NewBuilder().WriteBytes(c.t16a).Pack(0x16a),
				tlv.T166(5),
				tlv.T521(0x13, "basicim"),
			).ToBytes()))

	if err != nil {
		return nil, err
	}

	res, err := c.decodeLoginResponse(response, &c.transport.Sig)
	if err != nil {
		return &res, err
	}
	if res.Success {
		err = c.init()
	}
	return &res, err
}

func (c *QQClient) FastLogin() error {
	if c.transport.Sig.Uin == 0 || len(c.transport.Sig.D2) == 0 {
		return errors.New("no login cache")
	}
	c.Uin = c.transport.Sig.Uin
	if c.Online.Load() {
		return ErrAlreadyOnline
	}
	err := c.connect()
	if err != nil {
		return err
	}
	err = c.register()
	if err != nil {
		return fmt.Errorf("failed to register session: %w", err)

	}
	return nil
}

func (c *QQClient) init() error {
	return c.register()
}

func (c *QQClient) register() error {
	response, err := c.sendUniPacketAndWait(
		"trpc.qq_new_tech.status_svc.StatusService.Register",
		wtlogin.BuildRegisterRequest(c.Version(), c.Device()))

	if err != nil {
		c.errorln(err)
		return err
	}

	err = wtlogin.ParseRegisterResponse(response)
	if err != nil {
		c.errorln("register failed:", err)
		return err
	}
	c.transport.Sig.Uin = c.Uin
	c.setOnline()
	go c.doHeartbeat()
	return nil
}

type (
	NTNewDeviceQrCodeRequest struct {
		StrDevAuthToken string `json:"str_dev_auth_token"`
		Uint32Flag      int    `json:"uint32_flag"`
		Uint32UrlType   int    `json:"uint32_url_type"`
		StrUinToken     string `json:"str_uin_token"`
		StrDevType      string `json:"str_dev_type"`
		StrDevName      string `json:"str_dev_name"`
	}

	NTNewDeviceQrCodeResponse struct {
		Uint32GuaranteeStatus int    `json:"uint32_guarantee_status"`
		StrURL                string `json:"str_url"`
		ActionStatus          string `json:"ActionStatus"`
		StrNtSuccToken        string `json:"str_nt_succ_token"`
		ErrorCode             int    `json:"ErrorCode"`
		ErrorInfo             string `json:"ErrorInfo"`
	}

	NTNewDeviceQrCodeQuery struct {
		Uint32Flag uint64 `json:"uint32_flag"`
		Token      string `json:"bytes_token"`
	}
)
