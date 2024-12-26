package client

import (
	"fmt"

	"github.com/RomiChan/protobuf/proto"
	tea "github.com/fumiama/gofastTEA"

	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"github.com/LagrangeDev/LagrangeGo/client/internal/oicq"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
)

// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/entities.go#L18

type (
	LoginError int

	LoginResponse struct {
		Success bool
		Code    byte
		Error   LoginError

		// Captcha info
		CaptchaImage []byte
		CaptchaSign  []byte

		// Unsafe device
		VerifyURL string

		// SMS needed
		SMSPhone string

		// other error
		ErrorMessage string
	}
)

//go:generate stringer -type=LoginError
const (
	NeedCaptcha            LoginError = 1
	OtherLoginError        LoginError = 3
	UnsafeDeviceError      LoginError = 4
	SMSNeededError         LoginError = 5
	TooManySMSRequestError LoginError = 6
	SMSOrVerifyNeededError LoginError = 7
	SliderNeededError      LoginError = 8
	UnknownLoginError      LoginError = -1
)

func (c *QQClient) buildCode2dPacket(uin uint32, cmdID int, body []byte) []byte {
	return c.buildLoginPacket(
		uin,
		"wtlogin.trans_emp",
		binary.NewBuilder(nil).
			WriteU8(0).
			WriteU16(uint16(len(body))+53).
			WriteU32(uint32(c.Version().AppID)).
			WriteU32(0x72).
			WriteBytes(make([]byte, 3)).
			WriteU32(uint32(utils.TimeStamp())).
			WriteU8(2).
			WriteU16(uint16(len(body)+49)).
			WriteU16(uint16(cmdID)).
			WriteBytes(make([]byte, 21)).
			WriteU8(3).
			WriteU32(50).
			WriteBytes(make([]byte, 14)).
			WriteU32(uint32(c.Version().AppID)).
			WriteBytes(body).
			ToBytes(),
	)
}

func (c *QQClient) buildLoginPacket(uin uint32, cmd string, body []byte) []byte {
	var _cmd uint16
	if cmd == "wtlogin.login" {
		_cmd = 2064
	} else {
		_cmd = 2066
	}

	return c.oicq.Marshal(&oicq.Message{
		Uin:              uin,
		Command:          _cmd,
		EncryptionMethod: oicq.EM_ECDH,
		Body:             body,
	})
}

func (c *QQClient) decodeLoginResponse(buf []byte, sig *auth.SigInfo) (LoginResponse, error) {
	reader := binary.NewReader(buf)
	reader.SkipBytes(2)
	typ := reader.ReadU8()
	tlv := reader.ReadTlv()

	switch typ {
	case 0:
		reader = binary.NewReader(tea.NewTeaCipher(sig.Tgtgt).Decrypt(tlv[0x119]))
		tlv = reader.ReadTlv()
		if tgt, ok := tlv[0x10A]; ok {
			sig.Tgt = tgt
		}
		if d2, ok := tlv[0x143]; ok {
			sig.D2 = d2
		}
		if d2Key, ok := tlv[0x305]; ok {
			sig.D2Key = d2Key
		}
		if tlv11a, ok := tlv[0x11A]; ok {
			c.debug("tlv11a data: %x", tlv11a)
			tlvReader := binary.NewReader(tlv11a)
			tlvReader.ReadU16()
			sig.Age = tlvReader.ReadU8()
			sig.Gender = tlvReader.ReadU8()
			sig.Nickname = tlvReader.ReadStringWithLength("u8", false)
		}
		sig.Tgtgt = crypto.MD5Digest(sig.D2Key)
		sig.TempPwd = tlv[0x106]

		var resp pb.Tlv543
		err := proto.Unmarshal(tlv[0x543], &resp)
		if err != nil {
			err = fmt.Errorf("parsing login response error: %w", err)
			c.errorln(err)
			return LoginResponse{
				Success:      false,
				Code:         typ,
				Error:        UnknownLoginError,
				ErrorMessage: err.Error(),
			}, err
		}
		sig.UID = resp.Layer1.Layer2.Uid

		c.debugln("SigInfo got")

		return LoginResponse{
			Success: true,
		}, nil
	default:
	}

	var loginErr LoginError
	var title, content string

	if t146, ok := tlv[0x146]; ok {
		errBuf := binary.NewReader(t146)
		errBuf.SkipBytes(4)
		title = errBuf.ReadString(int(errBuf.ReadU16()))
		content = errBuf.ReadString(int(errBuf.ReadU16()))
		loginErr = OtherLoginError

	} else if t149, ok := tlv[0x149]; ok {
		errBuf := binary.NewReader(t149)
		errBuf.SkipBytes(2)
		title = errBuf.ReadString(int(errBuf.ReadU16()))
		content = errBuf.ReadString(int(errBuf.ReadU16()))
		loginErr = OtherLoginError
	} else {
		title = "未知错误"
		content = "无法解析错误原因，请将完整日志提交给开发者"
		loginErr = UnknownLoginError
	}

	err := fmt.Errorf("login fail on oicq (0x%02x): [%s]>[%s]", typ, title, content)
	c.errorln(err)
	return LoginResponse{
		Success:      false,
		Error:        loginErr,
		ErrorMessage: err.Error(),
	}, nil
}
