package message

// from https://github.com/Mrs4s/MiraiGo/blob/master/message/elements.go

import (
	"os"
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/utils"

	"github.com/LagrangeDev/LagrangeGo/pkg/pb/message"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
)

var messageLogger = utils.GetLogger("message")

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

func (e *TextElement) BuildElement() []*message.Elem {
	return []*message.Elem{{Text: &message.Text{Str: proto.Some(e.Content)}}}
}

func (e *AtElement) BuildElement() []*message.Elem {
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
	return []*message.Elem{{Text: &message.Text{
		Str:       proto.Some(e.Display),
		PbReserve: reserveData,
	}}}
}

func (e *FaceElement) BuildElement() []*message.Elem {
	faceId := int32(e.FaceID)
	if e.isLargeFace {
		qFace := message.QFaceExtra{
			Field1:  proto.Some("1"),
			Field2:  proto.Some("8"),
			FaceId:  proto.Some(faceId),
			Field4:  proto.Some(int32(1)),
			Field5:  proto.Some(int32(1)),
			Field6:  proto.Some(""),
			Preview: proto.Some(""),
			Field9:  proto.Some(int32(1)),
		}
		qFaceData, _ := proto.Marshal(&qFace)
		return []*message.Elem{{
			CommonElem: &message.CommonElem{
				ServiceType:  37,
				PbElem:       qFaceData,
				BusinessType: 1,
			},
		}}
	} else {
		return []*message.Elem{{
			Face: &message.Face{Index: proto.Some(faceId)},
		}}
	}
}

func (e *GroupImageElement) BuildElement() []*message.Elem {
	common, err := proto.Marshal(e.MsgInfo)
	if err != nil {
		messageLogger.Errorln("ImageBuild Common Proto Marshall failed:", err)
		return nil
	}
	msg := []*message.Elem{{
		CommonElem: &message.CommonElem{
			ServiceType:  48,
			PbElem:       common,
			BusinessType: 10,
		},
	}}
	return msg
}

func (e *FriendImageElement) BuildElement() []*message.Elem {
	common, err := proto.Marshal(e.MsgInfo)
	if err != nil {
		messageLogger.Errorln("ImageBuild Common Proto Marshall failed:", err)
		return nil
	}
	var msg []*message.Elem
	if e.CompatImage != nil {
		msg = []*message.Elem{{
			NotOnlineImage: e.CompatImage,
		}}
	}
	msg = append(msg, &message.Elem{
		CommonElem: &message.CommonElem{
			ServiceType:  48,
			PbElem:       common,
			BusinessType: 10,
		},
	})
	return msg
}

func (e *ReplyElement) BuildElement() []*message.Elem {
	return nil
}

func (e *VoiceElement) BuildElement() []*message.Elem {
	return nil
}

func (e *ShortVideoElement) BuildElement() []*message.Elem {
	return nil
}
