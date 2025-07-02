package tlv

import (
	"strconv"

	ftea "github.com/fumiama/gofastTEA"

	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	"github.com/LagrangeDev/LagrangeGo/utils/io"
)

// T18 默认参数 pingVersion, unknown = 0, ssoVersion = 5
func T18(appID, appClientVersion, uin, pingVersion, ssoVersion, unknown int) []byte {
	return binary.NewBuilder().
		WriteU16(uint16(pingVersion)).
		WriteU32(uint32(ssoVersion)).
		WriteU32(uint32(appID)).
		WriteU32(uint32(appClientVersion)).
		WriteU32(uint32(uin)).
		WriteU16(uint16(unknown)).
		WriteU16(0).Pack(0x18)
}

// T100 dbBufVer 默认为 0
func T100(ssoVersion, appID, subAppID, appClientVersion, sigmap, dbBufVer int) []byte {
	return binary.NewBuilder().
		WriteU16(uint16(dbBufVer)).
		WriteU32(uint32(ssoVersion)).
		WriteU32(uint32(appID)).
		WriteU32(uint32(subAppID)).
		WriteU32(uint32(appClientVersion)).
		WriteU32(uint32(sigmap)).
		Pack(0x100)
}

// T106 抄的时候注意参数顺序
func T106(appID, appClientVersion, uin int, guid string, passwordMd5, tgtgtKey, ip []byte, savePassword bool) []byte {
	// password_md5 + bytes(4) + write_u32(uin).pack()
	key := crypto.MD5Digest(binary.NewBuilder().
		WriteBytes(passwordMd5).
		WriteU32(0).
		WriteU32(uint32(uin)).
		ToBytes())

	body := binary.NewBuilder().
		WriteStruct(uint16(4), //  tgtgt version
			crypto.RandU32(),
			uint32(0), // sso_version, depreciated
			uint32(appID),
			uint32(appClientVersion),
			uint64(uin)).
		WriteU32(uint32(io.TimeStamp())).
		WriteBytes(ip).
		WriteBool(savePassword).
		WriteBytes(passwordMd5).
		WriteBytes(tgtgtKey).
		WriteU32(0).
		WriteBool(true).
		WriteBytes(io.MustParseHexStr(guid)).
		WriteU32(0).
		WriteU32(1).
		WritePacketString(strconv.Itoa(uin), "u16", false).
		ToBytes()

	return binary.NewBuilder().
		WritePacketBytes(ftea.NewTeaCipher(key).Encrypt(body), "u32", true).
		Pack(0x106)
}

// T107 默认参数为 1, 0x0d, 0, 1
func T107(picType, capType, picSize, retType int) []byte {
	return binary.NewBuilder().
		WriteU16(uint16(picType)).
		WriteU8(uint8(capType)).
		WriteU16(uint16(picSize)).
		WriteU8(uint8(retType)).
		Pack(0x107)
}

func T116(subSigmap int) []byte {
	return binary.NewBuilder().
		WriteU8(0).
		WriteU32(12058620). // unknown?
		WriteU32(uint32(subSigmap)).
		WriteU8(0).
		Pack(0x116)
}

func T124() []byte {
	return binary.NewBuilder().
		WriteBytes(make([]byte, 12)).
		Pack(0x124)
}

func T128(appInfoOS string, deviceGUID []byte) []byte {
	return binary.NewBuilder().
		WriteU16(0).
		WriteU8(0).
		WriteU8(1).
		WriteU8(0).
		WriteU32(0).
		WritePacketString(appInfoOS, "u16", false).
		WritePacketBytes(deviceGUID, "u16", false).
		WritePacketString("", "u16", false).
		Pack(0x128)
}

// T141 默认参数 apn = []byte{0}
func T141(simInfo, apn []byte) []byte {
	return binary.NewBuilder().
		WritePacketBytes(simInfo, "u32", false).
		WritePacketBytes(apn, "u32", false).
		Pack(0x141)
}

// T142 默认参数 version = 0 注意apkID长度要过32
func T142(apkID string, version int) []byte {
	return binary.NewBuilder().
		WriteU16(uint16(version)).
		// WritePacketString(apkID[:32], "u16", false).
		// apkID长度没有32，不动了
		WritePacketString(apkID, "u16", false).
		Pack(0x142)
}

func T144(tgtgtKey []byte, appInfo *auth.AppInfo, device *auth.DeviceInfo) []byte {
	return binary.NewBuilder(tgtgtKey...).
		WriteTLV(
			T16e(device.DeviceName),
			T147(appInfo.AppID, appInfo.PTVersion, appInfo.PackageName),
			T128(appInfo.OS, io.MustParseHexStr(device.GUID)),
			T124(),
		).Pack(0x144)
}

func T145(guid []byte) []byte {
	return binary.NewBuilder().
		WriteBytes(guid).
		Pack(0x145)
}

func T147(appID int, ptVersion string, packageName string) []byte {
	return binary.NewBuilder().
		WriteU32(uint32(appID)).
		WritePacketString(ptVersion, "u16", false).
		WritePacketString(packageName, "u16", false).
		Pack(0x147)
}

func T166(imageType int) []byte {
	return binary.NewBuilder().
		WriteI8(int8(imageType)).
		Pack(0x166)
}

func T16a(noPicSig []byte) []byte {
	return binary.NewBuilder().
		WriteBytes(noPicSig).
		Pack(0x16a)
}

func T16e(deviceName string) []byte {
	return binary.NewBuilder().
		WriteBytes(io.S2B(deviceName)).
		Pack(0x16e)
}

// T177 默认参数 buildTime=0
func T177(sdkVersion string, buildTime int) []byte {
	return binary.NewBuilder().
		WriteStruct(uint8(1), uint32(buildTime)).
		WritePacketString(sdkVersion, "u16", false).
		Pack(0x177)
}

// T191 默认参数 canWebVerify=0
func T191(canWebVerify int) []byte {
	return binary.NewBuilder().
		WriteU8(uint8(canWebVerify)).
		Pack(0x191)
}

// T318 默认参数 tgtQr = []byte{0}
func T318(tgtQr []byte) []byte {
	return binary.NewBuilder().
		WriteBytes(tgtQr).
		Pack(0x318)
}

// T521 默认参数 0x13, "basicim"
func T521(productType int, productDesc string) []byte {
	return binary.NewBuilder().
		WriteU32(uint32(productType)).
		WritePacketString(productDesc, "u16", false).
		Pack(0x521)
}
