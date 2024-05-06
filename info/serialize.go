package info

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"os"

	"github.com/LagrangeDev/LagrangeGo/utils"

	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

func init() {
	// 这里不注册好像也可以
	gob.Register(SigInfo{})
}

func Encode(sig *SigInfo) []byte {
	buffer := new(bytes.Buffer)
	err := gob.NewEncoder(buffer).Encode(sig)
	if err != nil {
		panic(err)
	}
	dataHash := utils.MD5Digest(buffer.Bytes())

	return binary.NewBuilder(nil).
		WriteBytes(dataHash, true).
		WriteBytes(buffer.Bytes(), true).
		Pack(binary.PackTypeNone)
}

func Decode(buf []byte, verify bool) *SigInfo {
	reader := binary.NewReader(buf)
	dataHash := reader.ReadBytesWithLength("u16", false)
	data := reader.ReadBytesWithLength("u16", false)

	if verify && string(dataHash) != string(utils.MD5Digest(data)) {
		panic("Data hash does not match")
	}
	buffer := bytes.NewBuffer(data)
	var siginfo SigInfo
	err := gob.NewDecoder(buffer).Decode(&siginfo)
	if err != nil {
		panic(err)
	}
	return &siginfo
}

func LoadDevice(path string) *DeviceInfo {
	data, err := os.ReadFile(path)
	if err != nil {
		deviceinfo := NewDeviceInfo(int(utils.RandU32()))
		_ = SaveDevice(deviceinfo, path)
		return deviceinfo
	}
	var dinfo DeviceInfo
	err = json.Unmarshal(data, &dinfo)
	if err != nil {
		deviceinfo := NewDeviceInfo(int(utils.RandU32()))
		_ = SaveDevice(deviceinfo, path)
		return deviceinfo
	}
	return &dinfo
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

func LoadSig(path string) *SigInfo {
	data, err := os.ReadFile(path)
	if err != nil {
		return NewSigInfo(8848)
	}
	return Decode(data, true)
}

func SaveSig(sig *SigInfo, path string) error {
	data := Encode(sig)
	return os.WriteFile(path, data, 0666)
}
