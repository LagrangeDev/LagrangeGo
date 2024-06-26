package message

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
)

type ElementBuilder interface {
	BuildElement() []*message.Elem
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

func (e *ImageElement) BuildElement() []*message.Elem {
	common, err := proto.Marshal(e.MsgInfo)
	if err != nil {
		return nil
	}
	msg := []*message.Elem{{}, {
		CommonElem: &message.CommonElem{
			ServiceType:  48,
			PbElem:       common,
			BusinessType: 10,
		},
	}}
	if e.CompatFace != nil {
		msg[0].CustomFace = e.CompatFace
	}
	if e.CompatImage != nil {
		msg[0].NotOnlineImage = e.CompatImage
	}
	return msg
}

func (e *ReplyElement) BuildElement() []*message.Elem {
	return nil
}

func (e *VoiceElement) BuildElement() []*message.Elem {
	common, err := proto.Marshal(e.MsgInfo)
	if err != nil {
		return nil
	}
	return []*message.Elem{{
		CommonElem: &message.CommonElem{
			ServiceType:  48,
			PbElem:       common,
			BusinessType: 22,
		},
	}}
}

func (e *ShortVideoElement) BuildElement() []*message.Elem {
	return nil
}
