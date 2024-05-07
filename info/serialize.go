package info

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"os"

	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
)

var (
	ErrDataHashMismatch = errors.New("data hash mismatch")
)

func init() {
	// 这里不注册好像也可以
	gob.Register(SigInfo{})
}

func Encode(sig *SigInfo) ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := gob.NewEncoder(buffer).Encode(sig)
	if err != nil {
		return nil, err
	}
	dataHash := crypto.MD5Digest(buffer.Bytes())

	return binary.NewBuilder(nil).
		WriteBytes(dataHash, true).
		WriteBytes(buffer.Bytes(), true).
		ToBytes(), nil
}

func Decode(buf []byte, verify bool) (siginfo SigInfo, err error) {
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

func LoadDevice(path string) (*DeviceInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		deviceinfo := NewDeviceInfo(int(crypto.RandU32()))
		return deviceinfo, SaveDevice(deviceinfo, path)
	}
	var dinfo DeviceInfo
	err = json.Unmarshal(data, &dinfo)
	if err != nil {
		deviceinfo := NewDeviceInfo(int(crypto.RandU32()))
		return deviceinfo, SaveDevice(deviceinfo, path)
	}
	return &dinfo, nil
}

func SaveDevice(deviceInfo *DeviceInfo, path string) error {
	data, err := json.Marshal(deviceInfo)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, data, 0666)
	if err != nil {
		return err
	}
	return nil
}

func LoadSig(path string) (SigInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return NewSigInfo(8848), nil
	}
	return Decode(data, true)
}

func SaveSig(sig *SigInfo, path string) error {
	data, err := Encode(sig)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0666)
}
