package message

// from https://github.com/Mrs4s/MiraiGo/blob/master/message/elements.go

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	_ "embed"

	"github.com/tidwall/gjson"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/audio"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
)

//go:embed default_thumb.jpg
var DefaultThumb []byte

type (
	TextElement struct {
		Content string
	}

	AtElement struct {
		TargetUin uint32
		TargetUID string
		Display   string
		SubType   AtType
	}

	FaceElement struct {
		FaceID      uint16
		ResultID    uint16 // 猜拳和骰子的值
		isLargeFace bool
	}

	ReplyElement struct {
		ReplySeq  uint32
		SenderUin uint32
		SenderUID string
		GroupUin  uint32 // 私聊回复群聊时
		Time      uint32
		Elements  []IMessageElement
	}

	VoiceElement struct {
		Name string
		UUID string
		Size uint32
		URL  string
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
		ImageID  string
		FileUUID string // only in new protocol photo
		Size     uint32
		Width    uint32
		Height   uint32
		URL      string
		SubType  int32

		// EffectID show pic effect id.
		EffectID int32 // deprecated
		Flash    bool

		// send & receive
		Summary string
		Md5     []byte // only in old protocol photo
		IsGroup bool

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
		FileURL  string
		FileID   string // group
		FileUUID string // private
		FileHash string

		// send
		FileStream io.ReadSeeker
		FileSha1   []byte
	}

	ShortVideoElement struct {
		Name     string
		UUID     string
		Size     uint32
		URL      string
		Duration uint32

		// send
		Thumb   *VideoThumb
		Summary string
		Md5     []byte
		Sha1    []byte
		Stream  io.ReadSeeker
		MsgInfo *oidb.MsgInfo
		Compat  *message.VideoFile
	}

	VideoThumb struct {
		Stream io.ReadSeeker
		Size   uint32
		Md5    []byte
		Sha1   []byte
		Width  uint32
		Height uint32
	}

	LightAppElement struct {
		AppName string
		Content string
	}

	ForwardMessage struct {
		IsGroup bool
		SelfID  uint32
		ResID   string
		Nodes   []*ForwardNode
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
		ReplySeq:  m.ID,
		SenderUin: m.Sender.Uin,
		Time:      m.Time,
		Elements:  m.Elements,
	}
}

func NewPrivateReply(m *PrivateMessage) *ReplyElement {
	return &ReplyElement{
		ReplySeq:  m.ID,
		SenderUin: m.Sender.Uin,
		Time:      m.Time,
		Elements:  m.Elements,
	}
}

func NewRecord(data []byte, summary ...string) *VoiceElement {
	return NewStreamRecord(bytes.NewReader(data), summary...)
}

func NewStreamRecord(r io.ReadSeeker, summary ...string) *VoiceElement {
	md5, sha1, length := crypto.ComputeMd5AndSha1AndLength(r)
	return &VoiceElement{
		Size: uint32(length),
		Summary: func() string {
			if len(summary) != 0 {
				return summary[0]
			}
			return ""
		}(),
		Stream: r,
		Md5:    md5,
		Sha1:   sha1,
		Duration: func() uint32 {
			if info, err := audio.Decode(r); err == nil {
				return uint32(info.Time)
			}
			return uint32(length)
		}(),
	}
}

func NewFileRecord(path string, summary ...string) (*VoiceElement, error) {
	voice, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewStreamRecord(voice, summary...), nil
}

func NewImage(data []byte, summary ...string) *ImageElement {
	return NewStreamImage(bytes.NewReader(data), summary...)
}

func NewStreamImage(r io.ReadSeeker, summary ...string) *ImageElement {
	md5, sha1, length := crypto.ComputeMd5AndSha1AndLength(r)
	return &ImageElement{
		Size: uint32(length),
		Summary: func() string {
			if len(summary) != 0 {
				return summary[0]
			}
			return ""
		}(),
		Stream: r,
		Md5:    md5,
		Sha1:   sha1,
	}
}

func NewFileImage(path string, summary ...string) (*ImageElement, error) {
	img, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewStreamImage(img, summary...), nil
}

func NewVideo(data, thumb []byte, summary ...string) *ShortVideoElement {
	return NewStreamVideo(bytes.NewReader(data), bytes.NewReader(thumb), summary...)
}

func NewStreamVideo(r io.ReadSeeker, thumb io.ReadSeeker, summary ...string) *ShortVideoElement {
	md5, sha1, length := crypto.ComputeMd5AndSha1AndLength(r)
	return &ShortVideoElement{
		Size:  uint32(length),
		Thumb: NewVideoThumb(thumb),
		Summary: func() string {
			if len(summary) != 0 {
				return summary[0]
			}
			return ""
		}(),
		Md5:    md5,
		Sha1:   sha1,
		Stream: r,
		Compat: &message.VideoFile{},
	}
}

func NewFileVideo(path string, thumb []byte, summary ...string) (*ShortVideoElement, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewStreamVideo(file, bytes.NewReader(thumb), summary...), nil
}

func NewVideoThumb(r io.ReadSeeker) *VideoThumb {
	width := uint32(1920)
	height := uint32(1080)
	md5, sha1, size := crypto.ComputeMd5AndSha1AndLength(r)
	_, imgSize, err := utils.ImageResolve(r)
	if err == nil {
		width = uint32(imgSize.Width)
		height = uint32(imgSize.Height)
	}
	return &VideoThumb{
		Stream: r,
		Size:   uint32(size),
		Md5:    md5,
		Sha1:   sha1,
		Width:  width,
		Height: height,
	}
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

func NewLocalFile(path string, name ...string) (*FileElement, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return NewStreamFile(file, utils.LazyTernary(len(name) == 0, func() string {
		return filepath.Base(file.Name())
	}, func() string {
		return name[0]
	})), nil
}

func NewLightApp(content string) *LightAppElement {
	return &LightAppElement{
		AppName: gjson.Get(content, "app").Str,
		Content: content,
	}
}

func NewForward(resid string, nodes []*ForwardNode) *ForwardMessage {
	return &ForwardMessage{
		ResID: resid,
		Nodes: nodes,
	}
}

func NewForwardWithResID(resid string) *ForwardMessage {
	return &ForwardMessage{
		ResID: resid,
	}
}

func NewForwardWithNodes(nodes []*ForwardNode) *ForwardMessage {
	return &ForwardMessage{
		Nodes: nodes,
	}
}

func NewFace(id uint16) *FaceElement {
	return &FaceElement{FaceID: id}
}

func NewDice(value uint16) *FaceElement {
	if value > 6 {
		value = uint16(crypto.RandU32()%3) + 1
	}
	return &FaceElement{
		FaceID:      358,
		ResultID:    value,
		isLargeFace: true,
	}
}

type FingerGuessingType uint16

const (
	FingerGuessingRock     FingerGuessingType = 3 // 石头
	FingerGuessingScissors FingerGuessingType = 2 // 剪刀
	FingerGuessingPaper    FingerGuessingType = 1 // 布
)

func (m FingerGuessingType) String() string {
	switch m {
	case FingerGuessingRock:
		return "石头"
	case FingerGuessingScissors:
		return "剪刀"
	case FingerGuessingPaper:
		return "布"
	}
	return fmt.Sprint(int(m))
}

func NewFingerGuessing(value FingerGuessingType) *FaceElement {
	return &FaceElement{
		FaceID:      359,
		ResultID:    uint16(value),
		isLargeFace: true,
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
