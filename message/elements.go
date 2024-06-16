package message

// from https://github.com/Mrs4s/MiraiGo/blob/master/message/elements.go

import (
	"os"
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"

	"github.com/LagrangeDev/LagrangeGo/utils/crypto"

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
		Name string
		Size uint32
		Url  string
		Md5  []byte
		Sha1 []byte
		Node *oidb.IndexNode

		// --- sending ---
		MsgInfo  *oidb.MsgInfo
		Compat   []byte
		Duration uint32
		Data     []byte
	}

	ImageElement struct {
		ImageId   string
		FileId    int64
		ImageType int32
		Size      uint32
		Width     uint32
		Height    uint32
		Url       string

		// EffectID show pic effect id.
		EffectID int32
		Flash    bool

		// send
		Summary     string
		Ext         string
		Md5         []byte
		Sha1        []byte
		MsgInfo     *oidb.MsgInfo
		Stream      []byte
		CompatFace  *message.CustomFace     // GroupImage
		CompatImage *message.NotOnlineImage // FriendImage
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

func NewRecord(data []byte, duration uint32) *VoiceElement {
	return &VoiceElement{
		Size:     uint32(len(data)),
		Md5:      crypto.MD5Digest(data),
		Sha1:     crypto.SHA1Digest(data),
		Data:     data,
		Duration: duration,
	}
}

func NewImage(data []byte, Summary ...string) *ImageElement {
	var summary string
	if len(Summary) != 0 {
		summary = Summary[0]
	}
	return &ImageElement{
		Size:    uint32(len(data)),
		Summary: summary,
		Stream:  data,
		Md5:     crypto.MD5Digest(data),
		Sha1:    crypto.SHA1Digest(data),
	}
}

func NewImageByFile(path string, Summary ...string) (*ImageElement, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewImage(data, Summary...), nil
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

func (e *ReplyElement) Type() ElementType {
	return Reply
}

func (e *VoiceElement) Type() ElementType {
	return Voice
}

func (e *ImageElement) Type() ElementType {
	return Image
}

func (e *ShortVideoElement) Type() ElementType {
	return Video
}
