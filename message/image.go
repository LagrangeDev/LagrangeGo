package message

import (
	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
)

type GroupImageElement struct {
	ImageId   string
	FileId    int64
	ImageType int32
	Size      uint32
	Width     int32
	Height    int32
	Md5       []byte
	Url       string

	// EffectID show pic effect id.
	EffectID int32
	Flash    bool

	// Send
	MsgInfo    *oidb.MsgInfo
	Stream     []byte
	CompatFace []byte
}

type FriendImageElement struct {
	ImageId string
	Md5     []byte
	Size    uint32
	Width   int32
	Height  int32
	Url     string

	Flash bool
}
