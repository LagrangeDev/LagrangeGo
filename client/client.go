package client

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/auth"

	"github.com/LagrangeDev/LagrangeGo/client/packets/tlv"
	"github.com/LagrangeDev/LagrangeGo/client/packets/wtlogin"
	"github.com/LagrangeDev/LagrangeGo/client/packets/wtlogin/loginstate"
	"github.com/LagrangeDev/LagrangeGo/client/packets/wtlogin/qrcodestate"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
)

func (c *QQClient) Login(password, qrcodePath string) error {
	// prefer session login
	if len(c.transport.Sig.D2) != 0 && c.transport.Sig.Uin != 0 {
		c.infoln("Session found, try to login with session")
		c.Uin = c.transport.Sig.Uin
		if c.Online.Load() {
			return ErrAlreadyOnline
		}
		err := c.connect()
		if err != nil {
			return err
		}
		err = c.Register()
		if err != nil {
			err = fmt.Errorf("failed to register session: %v", err)
			c.errorln(err)
			return err
		}
		return nil
	}

	if len(c.transport.Sig.TempPwd) != 0 {
		err := c.keyExchange()
		if err != nil {
			return err
		}

		ret, err := c.TokenLogin()
		if err != nil {
			return fmt.Errorf("EasyLogin fail: %s", err)
		}

		if ret.Successful() {
			return c.Register()
		}
	}

	if password != "" {
		c.infoln("login with password")
		err := c.keyExchange()
		if err != nil {
			return err
		}

		for {
			ret, err := c.PasswordLogin(password)
			if err != nil {
				return err
			}
			switch {
			case ret.Successful():
				return c.Register()
			case ret == loginstate.CaptchaVerify:
				c.warningln("captcha verification required")
				c.infoln("ticket?->")
				c.transport.Sig.CaptchaInfo[0] = utils.ReadLine()
				c.infoln("rand_str?->")
				c.transport.Sig.CaptchaInfo[1] = utils.ReadLine()
			default:
				c.error("Unhandled exception raised: %s", ret.Name())
			}
		}
		// panic("unreachable")
	}
	c.infoln("login with qrcode")
	png, _, err := c.FetchQRCodeDefault()
	if err != nil {
		return err
	}
	err = os.WriteFile(qrcodePath, png, 0666)
	if err != nil {
		return err
	}
	c.info("qrcode saved to %s", qrcodePath)
	err = c.QRCodeLogin(3)
	if err != nil {
		return err
	}
	return c.Register()
}

func (c *QQClient) TokenLogin() (loginstate.State, error) {
	if c.Online.Load() {
		return -996, ErrAlreadyOnline
	}
	err := c.connect()
	if err != nil {
		return -997, err
	}
	data, err := buildNtloginRequest(c.Uin, c.version(), c.Device(), &c.transport.Sig, c.transport.Sig.TempPwd)
	if err != nil {
		return -998, err
	}
	packet, err := c.sendUniPacketAndWait(
		"trpc.login.ecdh.EcdhService.SsoNTLoginEasyLogin",
		data,
	)
	if err != nil {
		return -999, err
	}
	return parseNtloginResponse(packet, &c.transport.Sig)
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

	body := binary.NewBuilder(nil).
		WriteU16(0).
		WriteU64(0).
		WriteU8(0).
		WriteTLV(
			tlv.T16(c.version().AppID, c.version().SubAppID,
				utils.MustParseHexStr(c.Device().GUID), c.version().PTVersion, c.version().PackageName),
			tlv.T1b(0, 0, size, margin, 72, ecLevel, 2),
			tlv.T1d(c.version().MiscBitmap),
			tlv.T33(utils.MustParseHexStr(c.Device().GUID)),
			tlv.T35(c.version().PTOSVersion),
			tlv.T66(c.version().PTOSVersion),
			tlv.Td1(c.version().OS, c.Device().DeviceName),
		).WriteU8(3).ToBytes()

	packet := c.buildCode2dPacket(c.Uin, 0x31, body)
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

	body := binary.NewBuilder(nil).
		WritePacketBytes(c.transport.Sig.Qrsig, "u16", false).
		WriteU64(0).
		WriteU32(0).
		WriteU8(0).
		WriteU8(0x83).ToBytes()

	response, err := c.sendUniPacketAndWait("wtlogin.trans_emp",
		c.buildCode2dPacket(0, 0x12, body))
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
		c.errorln(err)
		return err
	}
	c.debug("keyexchange proto data: %x", packet)
	c.transport.Sig.ExchangeKey, c.transport.Sig.KeySig, err = wtlogin.ParseKeyExchangeResponse(packet)
	return err
}

func (c *QQClient) PasswordLogin(password string) (loginstate.State, error) {
	if c.Online.Load() {
		return -996, ErrAlreadyOnline
	}
	err := c.connect()
	if err != nil {
		return -997, err
	}

	md5Password := crypto.MD5Digest(utils.S2B(password))

	cr := tlv.T106(
		c.version().AppID,
		c.version().AppClientVersion,
		int(c.Uin),
		c.Device().GUID,
		md5Password,
		c.transport.Sig.Tgtgt,
		make([]byte, 4),
		true)[4:]

	data, err := buildNtloginRequest(c.Uin, c.version(), c.Device(), &c.transport.Sig, cr)
	if err != nil {
		return -998, err
	}
	packet, err := c.sendUniPacketAndWait(
		"trpc.login.ecdh.EcdhService.SsoNTLoginPasswordLogin",
		data,
	)
	if err != nil {
		return -999, err
	}
	return parseNtloginResponse(packet, &c.transport.Sig)
}

func (c *QQClient) QRCodeLogin(refreshInterval int) error {
	if c.transport.Sig.Qrsig == nil {
		return errors.New("no QrSig found, fetch qrcode first")
	}

	for {
		retCode, err := c.GetQRCodeResult()
		if err != nil {
			c.errorln(err)
			return err
		}
		if retCode.Waitable() {
			time.Sleep(time.Duration(refreshInterval) * time.Second)
			continue
		}
		if !retCode.Success() {
			return errors.New(retCode.Name())
		}
		break
	}

	app := c.version()
	device := c.Device()
	response, err := c.sendUniPacketAndWait(
		"wtlogin.login",
		c.buildLoginPacket(c.Uin, "wtlogin.login", binary.NewBuilder(nil).
			WriteU16(0x09).
			WriteTLV(
				binary.NewBuilder(nil).WriteBytes(c.t106).Pack(0x106),
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
				binary.NewBuilder(nil).WriteBytes(c.t16a).Pack(0x16a),
				tlv.T166(5),
				tlv.T521(0x13, "basicim"),
			).ToBytes()))

	if err != nil {
		c.errorln(err)
		return err
	}

	return c.decodeLoginResponse(response, &c.transport.Sig)
}

func (c *QQClient) FastLogin(sig *auth.SigInfo) error {
	if sig != nil {
		c.UseSig(*sig)
	}
	c.Uin = c.transport.Sig.Uin
	if c.Online.Load() {
		return ErrAlreadyOnline
	}
	err := c.connect()
	if err != nil {
		return err
	}
	err = c.Register()
	if err != nil {
		err = fmt.Errorf("failed to register session: %v", err)
		c.errorln(err)
		return err
	}
	return nil
}

func (c *QQClient) Register() error {
	response, err := c.sendUniPacketAndWait(
		"trpc.qq_new_tech.status_svc.StatusService.Register",
		wtlogin.BuildRegisterRequest(c.version(), c.Device()))

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
	c.infoln("register succeeded")
	return nil
}
