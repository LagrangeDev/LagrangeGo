package message

import "strconv"

type (
	TextElement struct {
		Content string
	}

	AtElement struct {
		Target  int64
		Display string
		SubType AtType
	}

	ReplyElement struct {
		ReplySeq int32
		Sender   int64
		GroupID  int64 // 私聊回复群聊时
		Time     int32
		Elements []IMessageElement
	}

	AtType int
)

const (
	AtTypeGroupMember = 0 // At群成员
)

func NewText(s string) *TextElement {
	return &TextElement{Content: s}
}

func NewAt(target int64, display ...string) *AtElement {
	dis := "@" + strconv.FormatInt(target, 10)
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
