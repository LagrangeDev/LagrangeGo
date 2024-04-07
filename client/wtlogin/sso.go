package wtlogin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/Redmomn/LagrangeGo/utils"
	"github.com/Redmomn/LagrangeGo/utils/crypto/ecdh"

	binary2 "github.com/Redmomn/LagrangeGo/utils/binary"

	qqtea "github.com/Redmomn/LagrangeGo/utils/crypto"
)

type SSOPacket struct {
	Seq       int
	RetCode   int
	Extra     string
	SessionID []byte
	Cmd       string
	Data      []byte
}

type SSOHeader struct {
	Flag uint8
	Uin  string
	Dec  []byte
}

func NewSSOPacket(seq, retCode int, extra string, sessionID []byte, cmd string, data []byte) *SSOPacket {
	return &SSOPacket{
		Seq:       seq,
		RetCode:   retCode,
		Extra:     extra,
		SessionID: sessionID,
		Cmd:       cmd,
		Data:      data,
	}
}

func ParseLv(buffer *bytes.Buffer) []byte {
	var length uint32
	_ = binary.Read(buffer, binary.BigEndian, &length)

	data := make([]byte, length-4)
	_ = binary.Read(buffer, binary.BigEndian, data)
	return data
}

func ParseSSOHeader(raw, d2Key []byte) (*SSOHeader, error) {
	buf := bytes.NewBuffer(raw)
	//keyExangeLogger.Debugln("buf ", buf.Len())
	// parse sso header
	buf.Next(4)
	var flag uint8
	_ = binary.Read(buf, binary.BigEndian, &flag)
	buf.Next(1)
	uin := string(ParseLv(buf))

	var dec []byte
	if flag == 0 { // no encrypted
		dec = make([]byte, buf.Len())
		_ = binary.Read(buf, binary.BigEndian, dec)
	} else if flag == 1 { // enc with d2key
		temp := make([]byte, buf.Len())
		_ = binary.Read(buf, binary.BigEndian, temp)
		dec = qqtea.QQTeaDecrypt(temp, d2Key)
	} else if flag == 2 { // enc with \x00*16
		temp := make([]byte, buf.Len())
		_ = binary.Read(buf, binary.BigEndian, temp)
		dec = qqtea.QQTeaDecrypt(temp, make([]byte, 16))
	} else {
		return nil, errors.New(fmt.Sprintf("invalid encrypt flag: %d", flag))
	}
	return &SSOHeader{
		Flag: flag,
		Uin:  uin,
		Dec:  dec,
	}, nil
}

func ParseSSOFrame(buffer []byte, IsOicqBody bool) (*SSOPacket, error) {
	reader := binary2.NewReader(buffer)
	_ = reader.ReadU32() // head length
	seq := reader.ReadI32()
	retCode := reader.ReadI32()
	extra := reader.ReadStringWithLength("u32", true) // extra
	cmd := reader.ReadStringWithLength("u32", true)
	sessionID := reader.ReadBytesWithLength("u32", true)

	if retCode != 0 {
		return NewSSOPacket(int(seq), int(retCode), extra, sessionID, "", make([]byte, 0)), nil
	}

	compressType := reader.ReadU32()

	reader.ReadBytesWithLength("u32", false)
	// withprefix 应该是true
	data := reader.ReadBytesWithLength("u32", true)

	// TODO py版本这里判断data是否为空，应该是不需要
	if compressType == 0 {

	} else if compressType == 1 {
		data, _ = utils.DecompressData(data)
	} else if compressType == 8 {
		data = data[4:]
	} else {
		return nil, errors.New(fmt.Sprintf("Unsupported compress type %d", compressType))
	}

	var err error
	if IsOicqBody && strings.Index(cmd, "wtlogin") == 0 {
		data, err = ParseOicqBody(data)
		if err != nil {
			return nil, err
		}
	}

	return NewSSOPacket(int(seq), int(retCode), extra, sessionID, cmd, data), nil
}

func ParseOicqBody(buf []byte) ([]byte, error) {
	buffer := bytes.NewBuffer(buf[:15])
	var flag uint8
	var encType uint16
	_ = binary.Read(buffer, binary.BigEndian, &flag)
	buffer.Next(12)
	_ = binary.Read(buffer, binary.BigEndian, &encType)
	// 还有1字节没读，不要了

	if flag != 2 {
		return nil, errors.New(fmt.Sprintf("Invalid OICQ response flag. Expected 2, got %d", flag))
	}

	body := buf[16 : len(buf)-1]
	if encType == 0 {
		return qqtea.QQTeaDecrypt(body, ecdh.ECDH["secp192k1"].GetShareKey()), nil
	} else {
		return nil, errors.New(fmt.Sprintf("Unknown encrypt type: %d", encType))
	}
}
