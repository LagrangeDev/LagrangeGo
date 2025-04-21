package tlv

import (
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

func T11(unusualSign []byte) []byte {
	return binary.NewBuilder().
		WriteBytes(unusualSign).
		Pack(0x11)
}

func T16(appid, appidQrcode int, guid []byte, ptVersion, packageName string) []byte {
	return binary.NewBuilder().
		WriteU32(0).
		WriteU32(uint32(appid)).
		WriteU32(uint32(appidQrcode)).
		WriteBytes(guid).
		WritePacketString(packageName, "u16", false).
		WritePacketString(ptVersion, "u16", false).
		WritePacketString(packageName, "u16", false).
		Pack(0x16)
}

func T1b(micro, version, size, margin, dpi, ecLevel, hint uint32) []byte {
	return binary.NewBuilder().
		WriteStruct(micro, version, size, margin, dpi, ecLevel, hint, uint16(0)).
		Pack(0x1B)
}

func T1d(miscBitmap int) []byte {
	return binary.NewBuilder().
		WriteU8(1).
		WriteU32(uint32(miscBitmap)).
		WriteU32(0).
		WriteU8(0).
		Pack(0x1d)
}

func T33(guid []byte) []byte {
	return binary.NewBuilder().WriteBytes(guid).Pack(0x33)
}

func T35(ptOSVersion int) []byte {
	return binary.NewBuilder().WriteU32(uint32(ptOSVersion)).Pack(0x35)
}

func T66(ptOSVersion int) []byte {
	return binary.NewBuilder().WriteU32(uint32(ptOSVersion)).Pack(0x66)
}

func Td1(appOS, deviceName string) []byte {
	return binary.NewBuilder().
		WriteBytes(proto.DynamicMessage{
			1: proto.DynamicMessage{
				1: appOS,
				2: deviceName,
			},
			4: proto.DynamicMessage{
				6: 1,
			},
		}.Encode()).
		Pack(0xd1)
}
