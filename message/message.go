package message

import (
	"github.com/LagrangeDev/LagrangeGo/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"strings"
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
		GroupCode int64
		GroupName string
		Self      int64
		Sender    *Sender
		Elements  []IMessageElement
	}

	GroupMessage struct {
		Id             int32
		InternalId     int32
		GroupCode      int64
		GroupName      string
		Sender         *Sender
		Time           int32
		Elements       []IMessageElement
		OriginalObject *message.Message
		// OriginalElements []*msg.Elem
	}

	SendingMessage struct {
		Elements []IMessageElement
	}

	Sender struct {
		Uin           int64
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

func ParseMessageElements(msg *message.PushMsg) []IMessageElement {
	var res []IMessageElement

	for _, elem := range msg.Message.Body.RichText.Elems {
		if elem.Text != nil {
			switch {
			case len(elem.Text.Attr6Buf) > 0:
				att6 := binary.NewReader(elem.Text.Attr6Buf)
				att6.ReadBytes(7)
				target := int64(uint32(att6.ReadI32()))
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

		if elem.CustomFace != nil {

		}

		if elem.NotOnlineImage != nil {

		}
	}

	return res
}

func BuildMessageElements() {

}
