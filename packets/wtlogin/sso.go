package wtlogin

import (
	"bytes"
	"encoding/binary"
	"fmt"

	ftea "github.com/fumiama/gofastTEA"

	"github.com/LagrangeDev/LagrangeGo/utils/crypto/ecdh"
)

func ParseOicqBody(buf []byte) ([]byte, error) {
	buffer := bytes.NewBuffer(buf[:15])
	var flag uint8
	var encType uint16
	_ = binary.Read(buffer, binary.BigEndian, &flag)
	buffer.Next(12)
	_ = binary.Read(buffer, binary.BigEndian, &encType)
	// 还有1字节没读，不要了

	if flag != 2 {
		return nil, fmt.Errorf("invalid OICQ response flag. Expected 2, got %d", flag)
	}

	body := buf[16 : len(buf)-1]
	if encType == 0 {
		return ftea.NewTeaCipher(ecdh.S192().SharedKey()).Decrypt(body), nil
	} else {
		return nil, fmt.Errorf("unknown encrypt type: %d", encType)
	}
}
