package auth

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
)

var (
	ErrDataHashMismatch = errors.New("data hash mismatch")
)

type SigInfo struct {
	Uin      uint32
	Sequence uint32
	UID      string

	Tgtgt []byte
	Tgt   []byte
	D2    []byte
	D2Key []byte

	Qrsig       []byte
	ExchangeKey []byte
	KeySig      []byte
	Cookies     string
	UnusualSig  []byte
	TempPwd     []byte
	CaptchaInfo [3]string

	NewDeviceVerifyURL string

	Nickname string
	Age      uint8
	Gender   uint8
}

func init() {
	// 这里不注册好像也可以
	gob.Register(SigInfo{})
}

func (sig *SigInfo) ClearSession() {
	sig.D2 = nil
	sig.Tgt = nil
	sig.D2Key = make([]byte, 16)
}

func (sig *SigInfo) Marshal() ([]byte, error) {
	buffer := binary.NewBuilder()
	err := gob.NewEncoder(buffer).Encode(sig)
	if err != nil {
		return nil, err
	}
	data := buffer.ToBytes()
	dataHash := crypto.MD5Digest(data)
	return binary.NewBuilder().
		WriteLenBytes(dataHash).
		WriteLenBytes(data).
		ToBytes(), nil
}

func UnmarshalSigInfo(buf []byte, verify bool) (siginfo SigInfo, err error) {
	reader := binary.NewReader(buf)
	dataHash := reader.ReadBytesWithLength("u16", false)
	data := reader.ReadBytesWithLength("u16", false)

	if verify && !bytes.Equal(dataHash, crypto.MD5Digest(data)) {
		err = ErrDataHashMismatch
		return
	}

	err = gob.NewDecoder(bytes.NewReader(data)).Decode(&siginfo)
	return
}
