package info

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	"github.com/LagrangeDev/LagrangeGo/utils"

	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

func JsonLoad(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	if err != nil {
		panic(err)
	}
}

func JsonDump(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func Encode(sig *SigInfo) []byte {
	buffer := new(bytes.Buffer)
	err := gob.NewEncoder(buffer).Encode(sig)
	if err != nil {
		panic(err)
	}
	dataHash := utils.Md5Digest(buffer.Bytes())

	return binary.NewBuilder(nil).
		WriteBytes(dataHash, true).
		WriteBytes(buffer.Bytes(), true).
		Pack(-1)
}

func Decode(buf []byte, verify bool) *SigInfo {
	reader := binary.NewReader(buf)
	dataHash := reader.ReadBytesWithLength("u16", false)
	data := reader.ReadBytesWithLength("u16", false)

	if verify && string(dataHash) != string(utils.Md5Digest(data)) {
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

func init() {
	// 这里不注册好像也可以
	gob.Register(SigInfo{})
}
