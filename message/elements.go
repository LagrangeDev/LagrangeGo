package message

import (
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
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

	ReplyElement struct {
		ReplySeq int32
		Sender   uint64
		GroupID  uint64 // 私聊回复群聊时
		Time     int32
		Elements []IMessageElement
	}

	VoiceElement struct {
		Name string
		Md5  []byte
		Size uint32
		Url  string

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

func (e *TextElement) Type() ElementType {
	return Text
}

func (e *AtElement) Type() ElementType {
	return At
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

func (e *TextElement) BuildElement() *message.Elem {
	return &message.Elem{Text: &message.Text{Str: proto.Some(e.Content)}}
}

func (e *AtElement) BuildElement() *message.Elem {
	var atAll int32 = 2
	if e.Target == 0 {
		atAll = 1
	}
	reserve := message.MentionExtra{
		Type:   proto.Some(atAll),
		Uin:    proto.Some(uint32(0)),
		Field5: proto.Some(int32(0)),
		Uid:    proto.Some(e.UID),
	}
	reserveData, _ := proto.Marshal(&reserve)
	return &message.Elem{Text: &message.Text{
		Str:       proto.Some(e.Display),
		PbReserve: reserveData,
	}}
}

func (e *GroupImageElement) BuildElement() *message.Elem {
	return nil
}

func (e *FriendImageElement) BuildElement() *message.Elem {
	return nil
}

func (e *ReplyElement) BuildElement() *message.Elem {
	return nil
}

func (e *VoiceElement) BuildElement() *message.Elem {
	return nil
}

func (e *ShortVideoElement) BuildElement() *message.Elem {
	return nil
}
