package message

// from https://github.com/Mrs4s/MiraiGo/blob/master/message/elements.go

import (
	"os"
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

type (
	TextElement struct {
		Content string
	}

	AtElement struct {
		Target  uint32
		UID     string
		Display string
		SubType AtType
	}

	FaceElement struct {
		FaceID      uint16
		isLargeFace bool
	}

	ReplyElement struct {
		ReplySeq int32
		Sender   uint64
		GroupID  uint64 // 私聊回复群聊时
		Time     int32
		Elements []IMessageElement
	}

	VoiceElement struct {
		Name    string
		Md5     []byte
		Size    uint32
		Url     string
		IsGroup bool
		Sha1    []byte
		Node    *oidb.IndexNode

		// --- sending ---
		Data []byte
	}
	ShortVideoElement struct {
		Name      string
		Uuid      []byte
		Size      int32
		ThumbSize int32
		Md5       []byte
		ThumbMd5  []byte
		Url       string
	}

	AtType int
)

const (
	AtTypeGroupMember = 0 // At群成员
)

func NewText(s string) *TextElement {
	return &TextElement{Content: s}
}

func NewAt(target uint32, display ...string) *AtElement {
	dis := "@" + strconv.FormatInt(int64(target), 10)
	if target == 0 {
		dis = "@全体成员"
	}
	if len(display) != 0 {
		dis = display[0]
	}
	return &AtElement{
		Target:  target,
		Display: dis,
	}
}

func NewGroupImage(data []byte) *GroupImageElement {
	return &GroupImageElement{
		Stream: data,
	}
}

func NewGroupImageByFile(path string) (*GroupImageElement, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &GroupImageElement{
		Stream: data,
	}, nil
}

func (e *TextElement) Type() ElementType {
	return Text
}

func (e *AtElement) Type() ElementType {
	return At
}

func (e *FaceElement) Type() ElementType {
	return Face
}

func (e *GroupImageElement) Type() ElementType {
	return Image
}

func (e *FriendImageElement) Type() ElementType {
	return Image
}

func (e *ReplyElement) Type() ElementType {
	return Reply
}

func (e *VoiceElement) Type() ElementType {
	return Voice
}

func (e *ShortVideoElement) Type() ElementType {
	return Video
}
