package wtlogin

import (
	"crypto/rand"
	"fmt"
	"strconv"

	"github.com/Redmomn/LagrangeGo/utils/crypto/ecdh"

	"github.com/Redmomn/LagrangeGo/packets/pb"

	"github.com/Redmomn/LagrangeGo/utils/binary"

	"github.com/Redmomn/LagrangeGo/utils/proto"

	"github.com/Redmomn/LagrangeGo/info"
	"github.com/Redmomn/LagrangeGo/utils"
	qqtea "github.com/Redmomn/LagrangeGo/utils/crypto"
)

var loginLogger = utils.GetLogger("login")

func BuildCode2dPacket(uin uint32, cmdID int, appInfo *info.AppInfo, body []byte) []byte {
	return BuildLoginPacket(
		uin,
		"wtlogin.trans_emp",
		appInfo,
		utils.NewPacketBuilder(nil).
			WriteU8(0).
			WriteU16(uint16(len(body))+53).
			WriteU32(uint32(appInfo.AppID)).
			WriteU32(0x72).
			WriteBytes(make([]byte, 3), "", true).
			WriteU32(uint32(utils.TimeStamp())).
			WriteU8(2).
			WriteU16(uint16(len(body)+49)).
			WriteU16(uint16(cmdID)).
			WriteBytes(make([]byte, 21), "", true).
			WriteU8(3).
			WriteU32(50).
			WriteBytes(make([]byte, 14), "", true).
			WriteU32(uint32(appInfo.AppID)).
			WriteBytes(body, "", true).
			Pack(-1),
	)
}

func BuildLoginPacket(uin uint32, cmd string, appinfo *info.AppInfo, body []byte) []byte {
	encBody := qqtea.QQTeaEncrypt(body, ecdh.ECDH["secp192k1"].GetShareKey())

	var _cmd uint16
	if cmd == "wtlogin.login" {
		_cmd = 2064
	} else {
		_cmd = 2066
	}

	frameBody := utils.NewPacketBuilder(nil).
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
		WriteBytes(make([]byte, 16), "", true).
		WriteU16(0x102).
		WriteU16(uint16(len(ecdh.ECDH["secp192k1"].GetPublicKey()))).
		WriteBytes(ecdh.ECDH["secp192k1"].GetPublicKey(), "", true).
		WriteBytes(encBody, "", true).
		WriteU8(3).
		Pack(-1)

	frame := utils.NewPacketBuilder(nil).
		WriteU8(2).
		WriteU16(uint16(len(frameBody))+3). // + 2 + 1
		WriteBytes(frameBody, "", true).
		Pack(-1)

	return frame
}

func BuildUniPacket(uin, seq int, cmd string, sign map[string]string,
	appInfo *info.AppInfo, deviceInfo *info.DeviceInfo, sigInfo *info.SigInfo, body []byte) []byte {

	trace := generateTrace()

	head := proto.DynamicMessage{
		15: trace,
		16: sigInfo.Uid,
	}

	if sign != nil {
		head[24] = proto.DynamicMessage{
			1: utils.GetBytesFromHex(sign["sign"]),
			2: utils.GetBytesFromHex(sign["token"]),
			3: utils.GetBytesFromHex(sign["extra"]),
		}
	}

	ssoHeader := utils.NewPacketBuilder(nil).
		WriteU32(uint32(seq)).
		WriteU32(uint32(appInfo.SubAppID)).
		WriteU32(2052).                                                  // locate id
		WriteBytes(append([]byte{0x02}, make([]byte, 11)...), "", true). //020000000000000000000000
		WriteBytes(sigInfo.Tgt, "u32", true).
		WriteString(cmd, "u32", true).
		WriteBytes(make([]byte, 0), "u32", true).
		WriteBytes(utils.GetBytesFromHex(deviceInfo.Guid), "u32", true).
		WriteBytes(make([]byte, 0), "u32", true).
		WriteString(appInfo.CurrentVersion, "u16", true).
		WriteBytes(head.Encode(), "u32", true).
		Pack(-1)

	ssoPacket := utils.NewPacketBuilder(nil).
		WriteBytes(ssoHeader, "u32", true).
		WriteBytes(body, "u32", true).
		Pack(-1)

	encrypted := qqtea.QQTeaEncrypt(ssoPacket, sigInfo.D2Key)

	var _s uint8
	if len(sigInfo.D2) == 0 {
		_s = 2
	} else {
		_s = 1
	}

	service := utils.NewPacketBuilder(nil).
		WriteU32(12).
		WriteU8(_s).
		WriteBytes(sigInfo.D2, "u32", true).
		WriteU8(0).
		WriteString(strconv.Itoa(uin), "u32", true).
		WriteBytes(encrypted, "", true).
		Pack(-1)

	return utils.NewPacketBuilder(nil).WriteBytes(service, "u32", true).Pack(-1)
}

func DecodeLoginResponse(buf []byte, sig *info.SigInfo) bool {
	reader := binary.NewReader(buf)
	reader.ReadBytes(2)
	typ := reader.ReadU8()
	tlv := reader.ReadTlv()

	var title, content string

	if typ == 0 {
		reader = binary.NewReader(qqtea.QQTeaDecrypt(tlv[0x119], sig.Tgtgt))
		tlv = reader.ReadTlv()
		if tgt, ok := tlv[0x10a]; ok {
			sig.Tgt = tgt
		}
		if d2, ok := tlv[0x143]; ok {
			sig.D2 = d2
		}
		if d2Key, ok := tlv[0x305]; ok {
			sig.D2Key = d2Key
		}
		sig.Tgtgt = utils.Md5Digest(sig.D2Key)
		sig.TempPwd = tlv[0x106]

		var resp pb.Tlv543
		err := proto.Unmarshal(tlv[0x543], &resp)
		if err != nil {
			loginLogger.Errorf("parsing login response error: %s", err)
		}
		sig.Uid = resp.Layer1.Layer2.Uid

		loginLogger.Debugln("SigInfo got")

		return true
	} else if errData, ok := tlv[0x146]; ok {
		errBuf := binary.NewReader(errData)
		errBuf.ReadBytes(4)
		title = errBuf.ReadString(int(errBuf.ReadU16()))
		content = errBuf.ReadString(int(errBuf.ReadU16()))
	} else if errData, ok := tlv[0x149]; ok {
		errBuf := binary.NewReader(errData)
		errBuf.ReadBytes(2)
		title = errBuf.ReadString(int(errBuf.ReadU16()))
		content = errBuf.ReadString(int(errBuf.ReadU16()))
	} else {
		title = "未知错误"
		content = "无法解析错误原因，请将完整日志提交给开发者"
	}

	loginLogger.Errorf("Login fail on oicq (%2x): [%s]>[%s]", typ, title, content)
	return false
}

func generateTrace() string {
	randomBytes1 := make([]byte, 16)
	randomBytes2 := make([]byte, 8)

	if _, err := rand.Read(randomBytes1); err != nil {
		return ""
	}
	if _, err := rand.Read(randomBytes2); err != nil {
		return ""
	}

	trace := fmt.Sprintf("00-%x-%x-01", randomBytes1, randomBytes2)
	return trace
}
