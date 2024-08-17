package message

// from https://github.com/Mrs4s/MiraiGo/blob/master/message/elements.go

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/LagrangeDev/LagrangeGo/utils/audio"

	"github.com/tidwall/gjson"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"

	"github.com/LagrangeDev/LagrangeGo/utils/crypto"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

type (
	TextElement struct {
		Content string
	}

	AtElement struct {
		TargetUin uint32
		TargetUid string
		Display   string
		SubType   AtType
	}

	FaceElement struct {
		FaceID      uint16
		isLargeFace bool
	}

	ReplyElement struct {
		ReplySeq  uint32
		SenderUin uint32
		SenderUid string
		GroupUin  uint32 // 私聊回复群聊时
		Time      uint32
		Elements  []IMessageElement
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
		Stream   io.ReadSeeker
		Summary  string
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
		Stream      io.ReadSeeker
		CompatFace  *message.CustomFace     // GroupImage
		CompatImage *message.NotOnlineImage // FriendImage
	}

	FileElement struct {
		FileSize uint64
		FileName string
		FileMd5  []byte
		FileUrl  string
		FileId   string // group
		FileUUID string // private
		FileHash string

		// send
		FileStream io.ReadSeeker
		FileSha1   []byte
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

	LightAppElement struct {
		AppName string
		Content string
	}

	ForwardMessage struct {
		ResID string
		Nodes []*ForwardNode
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
		TargetUin: target,
		Display:   dis,
	}
}

func NewGroupReply(m *GroupMessage) *ReplyElement {
	return &ReplyElement{
		ReplySeq:  uint32(m.Id),
		SenderUin: m.Sender.Uin,
		Time:      uint32(m.Time),
		Elements:  m.Elements,
	}
}

func NewPrivateReply(m *PrivateMessage) *ReplyElement {
	return &ReplyElement{
		ReplySeq:  uint32(m.Id),
		SenderUin: m.Sender.Uin,
		Time:      uint32(m.Time),
		Elements:  m.Elements,
	}
}

func NewRecord(data []byte, Summary ...string) *VoiceElement {
	return NewStreamRecord(bytes.NewReader(data), Summary...)
}

func NewStreamRecord(r io.ReadSeeker, Summary ...string) *VoiceElement {
	var summary string
	if len(Summary) != 0 {
		summary = Summary[0]
	}
	md5, sha1, length := crypto.ComputeMd5AndSha1AndLength(r)
	info, err := audio.Decode(r)
	if err != nil {
		return &VoiceElement{
			Size:     uint32(length),
			Summary:  summary,
			Stream:   r,
			Md5:      md5,
			Sha1:     sha1,
			Duration: uint32(length),
		}
	}
	return &VoiceElement{
		Size:     uint32(length),
		Summary:  summary,
		Stream:   r,
		Md5:      md5,
		Sha1:     sha1,
		Duration: uint32(info.Time),
	}
}

func NewFileRecord(path string, Summary ...string) (*VoiceElement, error) {
	voice, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewStreamRecord(voice, Summary...), nil
}

func NewImage(data []byte, Summary ...string) *ImageElement {
	return NewStreamImage(bytes.NewReader(data), Summary...)
}

func NewStreamImage(r io.ReadSeeker, Summary ...string) *ImageElement {
	var summary string
	if len(Summary) != 0 {
		summary = Summary[0]
	}
	md5, sha1, length := crypto.ComputeMd5AndSha1AndLength(r)
	return &ImageElement{
		Size:    uint32(length),
		Summary: summary,
		Stream:  r,
		Md5:     md5,
		Sha1:    sha1,
	}
}

func NewFileImage(path string, Summary ...string) (*ImageElement, error) {
	img, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewStreamImage(img, Summary...), nil
}

func NewFile(data []byte, fileName string) *FileElement {
	return NewStreamFile(bytes.NewReader(data), fileName)
}

func NewStreamFile(r io.ReadSeeker, fileName string) *FileElement {
	md5, sha1, length := crypto.ComputeMd5AndSha1AndLength(r)
	return &FileElement{
		FileName:   fileName,
		FileSize:   length,
		FileStream: r,
		FileMd5:    md5,
		FileSha1:   sha1,
	}
}

func NewLocalFile(path string) (*FileElement, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewStreamFile(file, filepath.Base(file.Name())), nil
}

func NewLightApp(content string) *LightAppElement {
	return &LightAppElement{
		AppName: gjson.Get(content, "app").Str,
		Content: content,
	}
}

func NewFoward(resid string) *ForwardMessage {
	// TODO 根据resid获取node内容
	return &ForwardMessage{
		ResID: resid,
	}
}

func NewNodeFoward(nodes []*ForwardNode) *ForwardMessage {
	return &ForwardMessage{
		Nodes: nodes,
	}
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

func (e *FileElement) Type() ElementType { return File }

func (e *ShortVideoElement) Type() ElementType {
	return Video
}

func (e *LightAppElement) Type() ElementType {
	return LightApp
}

func (e *ForwardMessage) Type() ElementType {
	return Forward
}
