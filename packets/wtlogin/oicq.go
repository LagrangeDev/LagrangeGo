package wtlogin

import (
	"fmt"

	ftea "github.com/fumiama/gofastTEA"

	"github.com/LagrangeDev/LagrangeGo/info"
	"github.com/LagrangeDev/LagrangeGo/packets/pb"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto/ecdh"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
)

var loginLogger = utils.GetLogger("login")

func BuildCode2dPacket(uin uint32, cmdID int, appInfo *info.AppInfo, body []byte) []byte {
	return BuildLoginPacket(
		uin,
		"wtlogin.trans_emp",
		appInfo,
		binary.NewBuilder(nil).
			WriteU8(0).
			WriteU16(uint16(len(body))+53).
			WriteU32(uint32(appInfo.AppID)).
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
			WriteU32(uint32(appInfo.AppID)).
			WriteBytes(body).
			ToBytes(),
	)
}

func BuildLoginPacket(uin uint32, cmd string, appinfo *info.AppInfo, body []byte) []byte {
	encBody := ftea.NewTeaCipher(ecdh.S192().SharedKey()).Encrypt(body)

	var _cmd uint16
	if cmd == "wtlogin.login" {
		_cmd = 2064
	} else {
		_cmd = 2066
	}

	pk := ecdh.S192().PublicKey()

	frameBody := binary.NewBuilder(nil).
		WriteU16(8001).
		WriteU16(_cmd).
		WriteU16(0).
		WriteU32(uin).
		WriteU8(3).
		WriteU8(135).
		WriteU32(0).
		WriteU8(19).
		WriteU16(0).
		WriteU16(uint16(appinfo.AppClientVersion)).
		WriteU32(0).
		WriteU8(1).
		WriteU8(1).
		WriteBytes(make([]byte, 16)).
		WriteU16(0x102).
		WriteU16(uint16(len(pk))).
		WriteBytes(pk).
		WriteBytes(encBody).
		WriteU8(3).
		ToBytes()

	frame := binary.NewBuilder(nil).
		WriteU8(2).
		WriteU16(uint16(len(frameBody)) + 3). // + 2 + 1
		WriteBytes(frameBody).
		ToBytes()

	return frame
}

func DecodeLoginResponse(buf []byte, sig *info.SigInfo) error {
	reader := binary.NewReader(buf)
	reader.SkipBytes(2)
	typ := reader.ReadU8()
	tlv := reader.ReadTlv()

	var title, content string

	if typ == 0 {
		reader = binary.NewReader(ftea.NewTeaCipher(sig.Tgtgt).Decrypt(tlv[0x119]))
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
			loginLogger.Debugf("tlv11a data: %x", tlv11a)
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
			err = fmt.Errorf("parsing login response error: %s", err)
			loginLogger.Errorln(err)
			return err
		}
		sig.Uid = resp.Layer1.Layer2.Uid

		loginLogger.Debugln("SigInfo got")

		return nil
	} else if errData, ok := tlv[0x146]; ok {
		errBuf := binary.NewReader(errData)
		errBuf.SkipBytes(4)
		title = errBuf.ReadString(int(errBuf.ReadU16()))
		content = errBuf.ReadString(int(errBuf.ReadU16()))
	} else if errData, ok := tlv[0x149]; ok {
		errBuf := binary.NewReader(errData)
		errBuf.SkipBytes(2)
		title = errBuf.ReadString(int(errBuf.ReadU16()))
		content = errBuf.ReadString(int(errBuf.ReadU16()))
	} else {
		title = "未知错误"
		content = "无法解析错误原因，请将完整日志提交给开发者"
	}

	err := fmt.Errorf("login fail on oicq (0x%02x): [%s]>[%s]", typ, title, content)
	loginLogger.Errorln(err)
	return err
}
