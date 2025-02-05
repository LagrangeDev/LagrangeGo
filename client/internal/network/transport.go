package network

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/internal/network/transport.go

import (
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	tea "github.com/fumiama/gofastTEA"
)

// Transport is a network transport.
type Transport struct {
	Sig     auth.SigInfo
	Version *auth.AppInfo
	Device  *auth.DeviceInfo
}

// PackPacket packs a packet.
func (t *Transport) PackPacket(req *Request) []byte {
	head := proto.DynamicMessage{
		15: utils.NewTrace(),
		16: t.Sig.UID,
	}

	if req.Sign != nil {
		sign := req.Sign.Value
		head[24] = proto.DynamicMessage{
			1: utils.MustParseHexStr(sign.Sign),
			2: utils.MustParseHexStr(sign.Token),
			3: utils.MustParseHexStr(sign.Extra),
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
		WritePacketBytes(utils.MustParseHexStr(t.Device.GUID), "u32", true).
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
