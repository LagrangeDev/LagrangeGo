package message

// 部分借鉴 https://github.com/Mrs4s/MiraiGo/blob/master/message/message.go

import (
	"reflect"
	"strconv"
	"strings"

	oidb2 "github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

type IMessage interface {
	GetElements() []IMessageElement
	Chat() int64
	ToString() string
	Texts() []string
}

const (
	Text     ElementType = iota // 文本
	Image                       // 图片
	Face                        // 表情
	At                          // 艾特
	Reply                       // 回复
	Service                     // 服务
	Forward                     // 转发
	File                        // 文件
	Voice                       // 语音
	Video                       // 视频
	LightApp                    // 轻应用
	RedBag                      // 红包
)

type (
	PrivateMessage struct {
		Id         int32
		InternalId int32
		Self       int64
		Target     int64
		Time       int32
		Sender     *Sender
		Elements   []IMessageElement
	}

	TempMessage struct {
		Id        int32
		GroupCode uint32
		GroupName string
		Self      uint32
		Sender    *Sender
		Elements  []IMessageElement
	}

	GroupMessage struct {
		Id             int32
		InternalId     int32
		GroupCode      uint32
		GroupName      string
		Sender         *Sender
		Time           uint64
		Elements       []IMessageElement
		OriginalObject *message.PushMsgBody
	}

	SendingMessage struct {
		Elements []IMessageElement
	}

	Sender struct {
		Uin           uint32
		Uid           string
		Nickname      string
		CardName      string
		AnonymousInfo *AnonymousInfo
		IsFriend      bool
	}

	AnonymousInfo struct {
		AnonymousId   string
		AnonymousNick string
	}

	IMessageElement interface {
		Type() ElementType
	}

	ElementType int
)

func (s *Sender) IsAnonymous() bool {
	return s.Uin == 80000000
}

func ParsePrivateMessage(msg *message.PushMsg) *PrivateMessage {
	prvMsg := &PrivateMessage{
		Id:         int32(msg.Message.ContentHead.Sequence.Unwrap()),
		InternalId: int32(msg.Message.ContentHead.MsgId.Unwrap()),
		Self:       int64(msg.Message.ResponseHead.ToUin),
		Target:     int64(msg.Message.ResponseHead.FromUin),
		Sender: &Sender{
			Uin:      msg.Message.ResponseHead.FromUin,
			Uid:      msg.Message.ResponseHead.FromUid.Unwrap(),
			IsFriend: true,
		},
		Time:     int32(msg.Message.ContentHead.TimeStamp.Unwrap()),
		Elements: parseMessageElements(msg.Message.Body.RichText.Elems),
	}
	if msg.Message != nil && msg.Message.Body != nil {
		prvMsg.Elements = append(prvMsg.Elements, ParseMessageBody(msg.Message.Body, false)...)
	}

	return prvMsg
}

func ParseGroupMessage(msg *message.PushMsg) *GroupMessage {
	grpMsg := &GroupMessage{
		Id:         int32(msg.Message.ContentHead.Sequence.Unwrap()),
		InternalId: int32(msg.Message.ContentHead.MsgId.Unwrap()),
		GroupCode:  msg.Message.ResponseHead.Grp.GroupUin,
		GroupName:  msg.Message.ResponseHead.Grp.GroupName,
		Sender: &Sender{
			Uin:      msg.Message.ResponseHead.FromUin,
			Uid:      msg.Message.ResponseHead.FromUid.Unwrap(),
			Nickname: msg.Message.ResponseHead.Grp.MemberName,
			CardName: msg.Message.ResponseHead.Grp.MemberName,
			IsFriend: false,
		},
		Time:           msg.Message.ContentHead.TimeStamp.Unwrap(),
		Elements:       parseMessageElements(msg.Message.Body.RichText.Elems),
		OriginalObject: msg.Message,
	}
	if msg.Message != nil && msg.Message.Body != nil {
		grpMsg.Elements = append(grpMsg.Elements, ParseMessageBody(msg.Message.Body, true)...)
	}
	return grpMsg
}

func ParseTempMessage(msg *message.PushMsg) *TempMessage {
	return &TempMessage{
		Elements: parseMessageElements(msg.Message.Body.RichText.Elems),
	}
}

func parseMessageElements(msg []*message.Elem) []IMessageElement {
	var res []IMessageElement

	for _, elem := range msg {
		if elem.SrcMsg != nil && len(elem.SrcMsg.OrigSeqs) != 0 {
			r := &ReplyElement{
				ReplySeq: int32(elem.SrcMsg.OrigSeqs[0]),
				Time:     elem.SrcMsg.Time.Unwrap(),
				Sender:   elem.SrcMsg.SenderUin,
				GroupID:  elem.SrcMsg.ToUin.Unwrap(),
				Elements: parseMessageElements(elem.SrcMsg.Elems),
			}
			res = append(res, r)
		}

		if elem.Text != nil {
			switch {
			case len(elem.Text.Attr6Buf) > 0:
				att6 := binary.NewReader(elem.Text.Attr6Buf)
				att6.SkipBytes(7)
				target := att6.ReadU32()
				at := NewAt(target, elem.Text.Str.Unwrap())
				at.SubType = AtTypeGroupMember
				res = append(res, at)
			default:
				res = append(res, NewText(func() string {
					if strings.Contains(elem.Text.Str.Unwrap(), "\r") && !strings.Contains(elem.Text.Str.Unwrap(), "\r\n") {
						return strings.ReplaceAll(elem.Text.Str.Unwrap(), "\r", "\r\n")
					}
					return elem.Text.Str.Unwrap()
				}()))
			}
		}

		if elem.Face != nil {
			if len(elem.Face.Old) > 0 {
				faceId := elem.Face.Index
				if faceId.IsSome() {
					res = append(res, &FaceElement{FaceID: uint16(faceId.Unwrap())})
				}
			} else if elem.CommonElem.ServiceType == 37 && elem.CommonElem.PbElem != nil {
				qFace := message.QFaceExtra{}
				err := proto.Unmarshal(elem.CommonElem.PbElem, &qFace)
				if err == nil {
					faceId := qFace.FaceId
					if faceId.IsSome() {
						res = append(res, &FaceElement{FaceID: uint16(faceId.Unwrap()), isLargeFace: true})
					}
				}
			} else if elem.CommonElem.ServiceType == 33 && elem.CommonElem.PbElem != nil {
				qFace := message.QSmallFaceExtra{}
				err := proto.Unmarshal(elem.CommonElem.PbElem, &qFace)
				if err == nil {
					res = append(res, &FaceElement{FaceID: uint16(qFace.FaceId), isLargeFace: false})
				}
			}
		}

		if elem.VideoFile != nil {
			return []IMessageElement{
				&ShortVideoElement{
					Name:      elem.VideoFile.FileName,
					Uuid:      utils.S2B(elem.VideoFile.FileUuid),
					Size:      elem.VideoFile.FileSize,
					ThumbSize: elem.VideoFile.ThumbFileSize,
					Md5:       elem.VideoFile.FileMd5,
					ThumbMd5:  elem.VideoFile.ThumbFileMd5,
				},
			}
		}

		if elem.CustomFace != nil {
			if len(elem.CustomFace.Md5) == 0 {
				continue
			}

			var url string
			if strings.Contains(elem.CustomFace.OrigUrl, "rkey") {
				url = "https://multimedia.nt.qq.com.cn" + elem.CustomFace.OrigUrl
			} else {
				url = "http://gchat.qpic.cn" + elem.CustomFace.OrigUrl
			}

			res = append(res, &ImageElement{
				FileId:  int64(elem.CustomFace.FileId),
				ImageId: elem.CustomFace.FilePath,
				Size:    elem.CustomFace.Size,
				Width:   uint32(elem.CustomFace.Width),
				Height:  uint32(elem.CustomFace.Height),
				Url:     url,
				Md5:     elem.CustomFace.Md5,
			})
		}

		if elem.NotOnlineImage != nil {
			if len(elem.NotOnlineImage.PicMd5) == 0 {
				continue
			}

			var url string
			if strings.Contains(elem.NotOnlineImage.OrigUrl, "rkey") {
				url = "https://multimedia.nt.qq.com.cn" + elem.NotOnlineImage.OrigUrl
			} else {
				url = "http://gchat.qpic.cn" + elem.NotOnlineImage.OrigUrl
			}

			res = append(res, &ImageElement{
				ImageId: elem.NotOnlineImage.FilePath,
				Size:    elem.NotOnlineImage.FileLen,
				Url:     url,
				Md5:     elem.NotOnlineImage.PicMd5,
			})
		}
	}

	return res
}

func ParseMessageBody(body *message.MessageBody, isGroup bool) []IMessageElement {
	var res []IMessageElement
	if body != nil {
		if body.RichText != nil && body.RichText.Ptt != nil {
			ptt := body.RichText.Ptt
			switch {
			case isGroup && ptt.FileId != 0:
				res = append(res, &VoiceElement{
					Name: ptt.FileName,
					Node: &oidb2.IndexNode{
						FileUuid: ptt.GroupFileKey,
					},
				})
			case !isGroup:
				res = append(res, &VoiceElement{
					Name: ptt.FileName,
					Node: &oidb2.IndexNode{
						FileUuid: ptt.FileUuid,
					},
				})
			}
		}
	}
	return res
}

func (msg *GroupMessage) ToString() string {
	var strBuilder strings.Builder
	for _, elem := range msg.Elements {
		switch e := elem.(type) {
		case *TextElement:
			strBuilder.WriteString(e.Content)
		case *ImageElement:
			strBuilder.WriteString("[Image: ")
			strBuilder.WriteString(e.ImageId)
			strBuilder.WriteString("]")
		case *AtElement:
			strBuilder.WriteString(e.Display)
		case *ReplyElement:
			strBuilder.WriteString("[Reply: ")
			strBuilder.WriteString(strconv.FormatInt(int64(e.ReplySeq), 10))
			strBuilder.WriteString("]")
		case *FaceElement:
			strBuilder.WriteString("[Face: ")
			strBuilder.WriteString(strconv.FormatInt(int64(e.FaceID), 10))
			if e.isLargeFace {
				strBuilder.WriteString(", isLargeFace: true]")
			}
			strBuilder.WriteString("]")
		case *VoiceElement:
			strBuilder.WriteString("[Record: ")
			strBuilder.WriteString(e.Name)
			strBuilder.WriteString("]")
		}
	}
	return strBuilder.String()
}

func (msg *PrivateMessage) ToString() string {
	return ToReadableString(msg.Elements)
}

func ToReadableString(m []IMessageElement) string {
	sb := new(strings.Builder)
	for _, elem := range m {
		switch e := elem.(type) {
		case *TextElement:
			sb.WriteString(e.Content)
		case *ImageElement:
			sb.WriteString("[图片]")
		case *AtElement:
			sb.WriteString(e.Display)
		case *ReplyElement:
			sb.WriteString("[回复]")
		case *FaceElement:
			sb.WriteString("[表情:")
			sb.WriteString(strconv.FormatInt(int64(e.FaceID), 10))
			sb.WriteString("]")
		case *VoiceElement:
			sb.WriteString("[语音:")
			sb.WriteString(e.Name)
			sb.WriteString("]")
		}
	}
	return sb.String()
}

func NewSendingMessage() *SendingMessage {
	return &SendingMessage{}
}

func (msg *SendingMessage) GetElems() []IMessageElement {
	return msg.Elements
}

// Append 要传入msg的引用
func (msg *SendingMessage) Append(e IMessageElement) *SendingMessage {
	v := reflect.ValueOf(e)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		msg.Elements = append(msg.Elements, e)
	}
	return msg
}

func (msg *SendingMessage) FirstOrNil(f func(element IMessageElement) bool) IMessageElement {
	for _, elem := range msg.Elements {
		if f(elem) {
			return elem
		}
	}
	return nil
}

func PackElementsToBody(msgElems []IMessageElement) (msgBody *message.MessageBody) {
	if len(msgElems) == 0 {
		return
	}
	elems := make([]*message.Elem, 0, len(msgElems))
	for _, elem := range msgElems {
		bd, ok := elem.(ElementBuilder)
		if !ok {
			continue
		}
		elems = append(elems, bd.BuildElement()...)
	}
	msgBody = &message.MessageBody{
		RichText: &message.RichText{Elems: elems},
	}
	return
}
