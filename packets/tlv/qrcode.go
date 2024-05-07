package tlv

import (
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
)

func T11(unusualSign []byte) []byte {
	return binary.NewBuilder(nil).
		WritePacketBytes(unusualSign, "", true).
		Pack(0x11)
}

func T16(appid, subAppid int, guid []byte, ptVersion, packageName string) []byte {
	return binary.NewBuilder(nil).
		WriteU32(0).
		WriteU32(uint32(appid)).
		WriteU32(uint32(subAppid)).
		WritePacketBytes(guid, "", true).
		WritePacketString(packageName, "u16", false).
		WritePacketString(ptVersion, "u16", false).
		WritePacketString(packageName, "u16", false).
		Pack(0x16)
}

func T1b() []byte {
	return binary.NewBuilder(nil).
		WriteStruct(uint32(0), uint32(0), uint32(3), uint32(4), uint32(72), uint32(2), uint32(2), uint16(0)).
		Pack(0x1B)
}

func T1d(miscBitmap int) []byte {
	return binary.NewBuilder(nil).
		WriteU8(1).
		WriteU32(uint32(miscBitmap)).
		WriteU32(0).
		WriteU8(0).
		Pack(0x1d)
}

func T33(guid []byte) []byte {
	return binary.NewBuilder(nil).WritePacketBytes(guid, "", true).Pack(0x33)
}

func T35(PtOSVersion int) []byte {
	return binary.NewBuilder(nil).WriteU32(uint32(PtOSVersion)).Pack(0x35)
}

func T66(PtOSVersion int) []byte {
	return binary.NewBuilder(nil).WriteU32(uint32(PtOSVersion)).Pack(0x66)
}

func Td1(AppOS, DeviceName string) []byte {
	return binary.NewBuilder(nil).
		WritePacketBytes(proto.DynamicMessage{
			1: proto.DynamicMessage{
				1: AppOS,
				2: DeviceName,
			},
			4: proto.DynamicMessage{
				6: 1,
			},
		}.Encode(), "", true).
		Pack(0xd1)
}
