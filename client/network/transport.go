package network

import (
	"strconv"

	tea "github.com/fumiama/gofastTEA"

	"github.com/LagrangeDev/LagrangeGo/info"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
)

// Transport is a network transport.
type Transport struct {
	Sig     info.SigInfo
	Version *info.AppInfo
	Device  *info.DeviceInfo
}

// PackPacket packs a packet.
func (t *Transport) PackPacket(req *Request) []byte {
	head := proto.DynamicMessage{
		15: utils.NewTrace(),
		16: t.Sig.Uid,
	}

	if req.Sign != nil {
		head[24] = proto.DynamicMessage{
			1: utils.MustParseHexStr(req.Sign["sign"]),
			2: utils.MustParseHexStr(req.Sign["token"]),
			3: utils.MustParseHexStr(req.Sign["extra"]),
		}
	}

	ssoHeader := binary.NewBuilder(nil).
		WriteU32(req.SequenceID).
		WriteU32(uint32(t.Version.SubAppID)).
		WriteU32(2052).                                        // locate id
		WriteBytes(append([]byte{0x02}, make([]byte, 11)...)). //020000000000000000000000
		WritePacketBytes(t.Sig.Tgt, "u32", true).
		WritePacketString(req.CommandName, "u32", true).
		WritePacketBytes(nil, "u32", true).
		WritePacketBytes(utils.MustParseHexStr(t.Device.Guid), "u32", true).
		WritePacketBytes(nil, "u32", true).
		WritePacketString(t.Version.CurrentVersion, "u16", true).
		WritePacketBytes(head.Encode(), "u32", true).
		ToBytes()

	ssoPacket := binary.NewBuilder(nil).
		WritePacketBytes(ssoHeader, "u32", true).
		WritePacketBytes(req.Body, "u32", true).
		ToBytes()

	encrypted := tea.NewTeaCipher(t.Sig.D2Key).Encrypt(ssoPacket)

	var _s uint8
	if len(t.Sig.D2) == 0 {
		_s = 2
	} else {
		_s = 1
	}

	service := binary.NewBuilder(nil).
		WriteU32(12).
		WriteU8(_s).
		WritePacketBytes(t.Sig.D2, "u32", true).
		WriteU8(0).
		WritePacketString(strconv.Itoa(int(req.Uin)), "u32", true).
		WriteBytes(encrypted).
		ToBytes()

	return binary.NewBuilder(nil).WritePacketBytes(service, "u32", true).ToBytes()
}
