package message

// 部分借鉴 https://github.com/Mrs4s/MiraiGo/blob/master/message/message.go

import (
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	oidb2 "github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
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
	Text       ElementType = iota // 文本
	Image                         // 图片
	Face                          // 表情
	At                            // 艾特
	Reply                         // 回复
	Service                       // 服务
	Forward                       // 转发
	File                          // 文件
	Voice                         // 语音
	Video                         // 视频
	LightApp                      // 轻应用
	RedBag                        // 红包
	MarketFace                    // 魔法表情
)

type (
	PrivateMessage struct {
		ID         uint32
		InternalID uint32
		ClientSeq  uint32
		Self       uint32
		Target     uint32
		Time       uint32
		Sender     *Sender
		Elements   []IMessageElement
	}

	TempMessage struct {
		ID        uint32
		GroupUin  uint32
		GroupName string
		Self      uint32
		Sender    *Sender
		Elements  []IMessageElement
	}

	GroupMessage struct {
		ID             uint32
		InternalID     uint32
		GroupUin       uint32
		GroupName      string
		Sender         *Sender
		Time           uint32
		Elements       []IMessageElement
		OriginalObject *message.PushMsgBody
	}

	SendingMessage struct {
		Elements []IMessageElement
	}

	Sender struct {
		Uin           uint32
		UID           string
		Nickname      string
		CardName      string
		AnonymousInfo *AnonymousInfo
		IsFriend      bool
	}

	AnonymousInfo struct {
		AnonymousID   string
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

func ParsePrivateMessage(msg *message.PushMsgBody) *PrivateMessage {
	prvMsg := &PrivateMessage{
		ID:         msg.ContentHead.Sequence.Unwrap(),
		InternalID: msg.ContentHead.MsgId.Unwrap(),
		Self:       msg.ResponseHead.ToUin,
		Target:     msg.ResponseHead.FromUin,
		Sender: &Sender{
			Uin:      msg.ResponseHead.FromUin,
			UID:      msg.ResponseHead.FromUid.Unwrap(),
			IsFriend: true,
		},
		Time:     msg.ContentHead.TimeStamp.Unwrap(),
		Elements: ParseMessageElements(msg.Body.RichText.Elems),
	}
	if msg.Body != nil {
		prvMsg.Elements = append(prvMsg.Elements, ParseMessageBody(msg.Body, false)...)
	}

	return prvMsg
}

func ParseGroupMessage(msg *message.PushMsgBody) *GroupMessage {
	grpMsg := &GroupMessage{
		ID:         msg.ContentHead.Sequence.Unwrap(),
		InternalID: msg.ContentHead.MsgId.Unwrap(),
		GroupUin:   msg.ResponseHead.Grp.GroupUin,
		GroupName:  msg.ResponseHead.Grp.GroupName,
		Sender: &Sender{
			Uin:      msg.ResponseHead.FromUin,
			UID:      msg.ResponseHead.FromUid.Unwrap(),
			Nickname: msg.ResponseHead.Grp.MemberName,
			CardName: msg.ResponseHead.Grp.MemberName,
			IsFriend: false,
		},
		Time:           msg.ContentHead.TimeStamp.Unwrap(),
		Elements:       ParseMessageElements(msg.Body.RichText.Elems),
		OriginalObject: msg,
	}
	if msg.Body != nil {
		grpMsg.Elements = append(grpMsg.Elements, ParseMessageBody(msg.Body, true)...)
	}
	return grpMsg
}

func ParseTempMessage(msg *message.PushMsgBody) *TempMessage {
	return &TempMessage{
		Elements: ParseMessageElements(msg.Body.RichText.Elems),
	}
}

func ParseMessageElements(msg []*message.Elem) []IMessageElement {
	var res []IMessageElement
	skipNext := false

	for _, elem := range msg {
		if skipNext {
			skipNext = false
			continue
		}
		if elem.SrcMsg != nil && len(elem.SrcMsg.OrigSeqs) != 0 {
			r := &ReplyElement{
				ReplySeq:  elem.SrcMsg.OrigSeqs[0],
				Time:      uint32(elem.SrcMsg.Time.Unwrap()),
				SenderUin: uint32(elem.SrcMsg.SenderUin),
				GroupUin:  uint32(elem.SrcMsg.ToUin.Unwrap()),
				Elements:  ParseMessageElements(elem.SrcMsg.Elems),
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
			switch {
			case len(elem.Face.Old) > 0:
				faceID := elem.Face.Index
				if faceID.IsSome() {
					res = append(res, &FaceElement{FaceID: uint32(faceID.Unwrap())})
				}
			case elem.CommonElem != nil && elem.CommonElem.ServiceType == 37 && elem.CommonElem.PbElem != nil:
				qFace := message.QFaceExtra{}
				if err := proto.Unmarshal(elem.CommonElem.PbElem, &qFace); err == nil {
					if qFace.Qsid.IsSome() {
						res = append(res, &FaceElement{FaceID: uint32(qFace.Qsid.Unwrap()), isLargeFace: true})
					}
				}
			case elem.CommonElem != nil && elem.CommonElem.ServiceType == 33 && elem.CommonElem.PbElem != nil:
				qFace := message.QSmallFaceExtra{}
				err := proto.Unmarshal(elem.CommonElem.PbElem, &qFace)
				if err == nil {
					res = append(res, &FaceElement{FaceID: qFace.FaceId, isLargeFace: false})
				}
			}
		}

		if elem.VideoFile != nil {
			video := elem.VideoFile
			res = append(res, &ShortVideoElement{
				Name: video.FileName,
				UUID: video.FileUuid,
				Size: uint32(video.FileSize),
				Md5:  video.FileMd5,
				Node: &oidb2.IndexNode{
					Info: &oidb2.FileInfo{
						FileName: video.FileName,
						FileSize: uint32(video.FileSize),
						FileHash: hex.EncodeToString(video.FileMd5),
					},
					FileUuid: video.FileUuid,
				},
				Thumb: &VideoThumb{
					Size: uint32(elem.VideoFile.ThumbFileSize),
					Md5:  elem.VideoFile.ThumbFileMd5,
				},
			})
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

			res = append(res, func() *ImageElement {
				img := &ImageElement{
					ImageID: elem.CustomFace.FilePath,
					Size:    elem.CustomFace.Size,
					Width:   uint32(elem.CustomFace.Width),
					Height:  uint32(elem.CustomFace.Height),
					URL:     url,
					Md5:     elem.CustomFace.Md5,
				}
				if elem.CustomFace.PbRes != nil {
					img.SubType = elem.CustomFace.PbRes.SubType
					img.Summary = elem.CustomFace.PbRes.Summary
				}
				return img
			}())
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

			img := &ImageElement{
				ImageID: elem.NotOnlineImage.FilePath,
				Size:    elem.NotOnlineImage.FileLen,
				Width:   elem.NotOnlineImage.PicWidth,
				Height:  elem.NotOnlineImage.PicHeight,
				URL:     url,
				Md5:     elem.NotOnlineImage.PicMd5,
			}

			if elem.NotOnlineImage.PbRes != nil {
				img.SubType = elem.NotOnlineImage.PbRes.SubType
				img.Summary = elem.NotOnlineImage.PbRes.Summary
			}

			res = append(res, img)
		}

		// new protocol image
		if elem.CommonElem != nil {
			switch elem.CommonElem.ServiceType {
			case 48:
				extra := &oidb2.MsgInfo{}
				err := proto.Unmarshal(elem.CommonElem.PbElem, extra)
				if err != nil {
					continue
				}
				index := extra.MsgInfoBody[0].Index
				switch elem.CommonElem.BusinessType {
				case 10, 20: // img
					res = append(res, &ImageElement{
						ImageID:  index.Info.FileName,
						FileUUID: index.FileUuid,
						SubType:  int32(extra.ExtBizInfo.Pic.BizType),
						Summary:  utils.Ternary(extra.ExtBizInfo.Pic.TextSummary == "", "[图片]", extra.ExtBizInfo.Pic.TextSummary),
						Md5:      utils.MustParseHexStr(index.Info.FileHash),
						Sha1:     utils.MustParseHexStr(index.Info.FileSha1),
						Width:    index.Info.Width,
						Height:   index.Info.Height,
						Size:     index.Info.FileSize,
						MsgInfo:  extra,
					})
				case 12, 22: // record 22 for Group
					res = append(res, &VoiceElement{
						Name:     index.Info.FileName,
						UUID:     index.FileUuid,
						Md5:      utils.MustParseHexStr(index.Info.FileHash),
						Sha1:     utils.MustParseHexStr(index.Info.FileSha1),
						Duration: index.Info.Time,
						Node:     index,
					})
				case 11, 21: // video
					var thumb = new(VideoThumb)
					if !(len(extra.MsgInfoBody) < 2) {
						info := extra.MsgInfoBody[1].Index
						thumb.Size = info.Info.FileSize
						thumb.Width = info.Info.Width
						thumb.Height = info.Info.Height
						thumb.Md5 = utils.MustParseHexStr(info.Info.FileHash)
						thumb.Sha1 = utils.MustParseHexStr(info.Info.FileSha1)
					}
					res = append(res, &ShortVideoElement{
						Name:  index.Info.FileName,
						UUID:  index.FileUuid,
						Md5:   utils.MustParseHexStr(index.Info.FileHash),
						Sha1:  utils.MustParseHexStr(index.Info.FileSha1),
						Size:  index.Info.FileSize,
						Thumb: thumb,
						Node:  index,
					})
				}
			case 3: // 闪照
				img := message.NotOnlineImage{}
				reader := binary.NewReader(elem.CommonElem.PbElem[1:])
				length, _ := reader.ReadUvarint()
				err := proto.Unmarshal(reader.ReadBytes(int(length)), &img)
				if err != nil {
					continue
				}
				res = append(res, &ImageElement{
					ImageID: img.FilePath,
					Md5:     img.PicMd5,
					Size:    img.FileLen,
					SubType: img.PbRes.SubType,
					Flash:   true,
					Summary: "[闪照]",
					Width:   img.PicWidth,
					Height:  img.PicHeight,
					URL:     fmt.Sprintf("http://gchat.qpic.cn/gchatpic_new/0/0-0-%X/0", img.PicMd5),
				})
				skipNext = true
			case 33:
				var newSysFaceMsg message.QSmallFaceExtra
				_ = proto.Unmarshal(elem.CommonElem.PbElem, &newSysFaceMsg)
				res = append(res, NewFace(newSysFaceMsg.FaceId))
			case 37:
				var faceExtra message.QFaceExtra
				_ = proto.Unmarshal(elem.CommonElem.PbElem, &faceExtra)
				result, _ := strconv.ParseInt(faceExtra.ResultId.Unwrap(), 10, 32)
				res = append(res, &FaceElement{
					FaceID:      uint32(faceExtra.Qsid.Unwrap()),
					ResultID:    uint32(result),
					isLargeFace: true,
				}) // sticker 永远为单独消息
				skipNext = true
			}
		}

		if elem.TransElem != nil && elem.TransElem.ElemType == 24 {
			payload := binary.NewReader(elem.TransElem.ElemValue)
			payload.SkipBytes(1)
			data := payload.ReadBytesWithLength("u16", false)
			extra := message.GroupFileExtra{}
			if err := proto.Unmarshal(data, &extra); err != nil {
				continue
			}
			res = append(res, &FileElement{
				FileSize: extra.Inner.Info.FileSize,
				FileMd5:  []byte(extra.Inner.Info.FileMd5),
				FileID:   extra.Inner.Info.FileId,
				FileName: extra.Inner.Info.FileName,
			})
		}

		if elem.RichMsg != nil && elem.RichMsg.ServiceId.Unwrap() == 35 {
			if elem.RichMsg.Template1 != nil {
				xmlData := binary.ZlibUncompress(elem.RichMsg.Template1[1:])
				multimsg := MultiMessage{}
				if err := xml.Unmarshal(xmlData, &multimsg); err == nil {
					res = append(res, NewForwardWithResID(multimsg.ResID))
				} else {
					res = append(res, &XMLElement{
						ServiceID: 35,
						Content:   utils.B2S(binary.ZlibUncompress(elem.RichMsg.Template1[1:])),
					})
				}
			}
		}

		if elem.LightAppElem != nil && len(elem.LightAppElem.Data) > 1 {
			var content []byte
			if elem.LightAppElem.Data[0] == 0 {
				content = elem.LightAppElem.Data[1:]
			}
			if elem.LightAppElem.Data[0] == 1 {
				content = binary.ZlibUncompress(elem.LightAppElem.Data[1:])
			}
			if len(content) > 0 && len(content) < 1024*1024*1024 { // 解析出错 or 非法内容
				res = append(res, NewLightApp(utils.B2S(content)))
			}
		}

		if elem.MarketFace != nil {
			res = append(res, &MarketFaceElement{
				Summary:    elem.MarketFace.FaceName.Unwrap(),
				ItemType:   elem.MarketFace.ItemType.Unwrap(),
				FaceInfo:   elem.MarketFace.FaceInfo.Unwrap(),
				FaceID:     elem.MarketFace.FaceId,
				TabID:      elem.MarketFace.TabId.Unwrap(),
				SubType:    elem.MarketFace.SubType.Unwrap(),
				EncryptKey: elem.MarketFace.Key,
				MediaType:  elem.MarketFace.MediaType.Unwrap(),
				MagicValue: utils.B2S(elem.MarketFace.MobileParam),
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
					UUID: ptt.FileUuid,
					Md5:  ptt.FileMd5,
					Node: &oidb2.IndexNode{
						FileUuid: ptt.GroupFileKey,
					},
				})
			case !isGroup:
				res = append(res, &VoiceElement{
					Name: ptt.FileName,
					UUID: ptt.FileUuid,
					Md5:  ptt.FileMd5,
					Node: &oidb2.IndexNode{
						FileUuid: ptt.FileUuid,
					},
				})
			}
		}
		if body.MsgContent != nil {
			extra := message.FileExtra{}
			if err := proto.Unmarshal(body.MsgContent, &extra); err != nil {
				return res
			}
			if extra.File.FileSize.IsSome() && extra.File.FileName.IsSome() && extra.File.FileMd5 != nil && extra.File.FileUuid.IsSome() && extra.File.FileHash.IsSome() {
				res = append(res, &FileElement{
					FileSize: uint64(extra.File.FileSize.Unwrap()),
					FileName: extra.File.FileName.Unwrap(),
					FileMd5:  extra.File.FileMd5,
					FileUUID: extra.File.FileUuid.Unwrap(),
					FileHash: extra.File.FileHash.Unwrap(),
				})
			}
		}
	}
	return res
}

func (msg *GroupMessage) ToString() string {
	//var strBuilder strings.Builder
	//for _, elem := range msg.Elements {
	//	switch e := elem.(type) {
	//	case *TextElement:
	//		strBuilder.WriteString(e.Content)
	//	case *ImageElement:
	//		strBuilder.WriteString("[Image: ")
	//		strBuilder.WriteString(e.ImageID)
	//		strBuilder.WriteString("]")
	//	case *AtElement:
	//		strBuilder.WriteString(e.Display)
	//	case *ReplyElement:
	//		strBuilder.WriteString("[Reply: ")
	//		strBuilder.WriteString(strconv.FormatInt(int64(e.ReplySeq), 10))
	//		strBuilder.WriteString("]")
	//	case *FaceElement:
	//		strBuilder.WriteString("[Face: ")
	//		strBuilder.WriteString(strconv.FormatInt(int64(e.FaceID), 10))
	//		if e.isLargeFace {
	//			strBuilder.WriteString(", isLargeFace: true]")
	//		}
	//		strBuilder.WriteString("]")
	//	case *VoiceElement:
	//		strBuilder.WriteString("[Record: ")
	//		strBuilder.WriteString(e.Name)
	//		strBuilder.WriteString("]")
	//	}
	//}
	//return strBuilder.String()
	return ToReadableString(msg.Elements)
}

func (msg *GroupMessage) GetElements() []IMessageElement {
	return msg.Elements
}

func (msg *GroupMessage) Chat() int64 {
	return int64(msg.ID)
}

func (msg *GroupMessage) Texts() []string {
	texts := make([]string, 0, len(msg.Elements))
	for _, elem := range msg.Elements {
		texts = append(texts, ToReadableStringEle(elem))
	}
	return texts
}

func (msg *PrivateMessage) ToString() string {
	return ToReadableString(msg.Elements)
}

func (msg *PrivateMessage) GetElements() []IMessageElement {
	return msg.Elements
}

func (msg *PrivateMessage) Chat() int64 {
	return int64(msg.ID)
}

func (msg *PrivateMessage) Texts() []string {
	texts := make([]string, 0, len(msg.Elements))
	for _, elem := range msg.Elements {
		texts = append(texts, ToReadableStringEle(elem))
	}
	return texts
}

func (msg *TempMessage) ToString() string {
	return ToReadableString(msg.Elements)
}

func (msg *TempMessage) GetElements() []IMessageElement {
	return msg.Elements
}

func (msg *TempMessage) Chat() int64 {
	return int64(msg.ID)
}

func (msg *TempMessage) Texts() []string {
	texts := make([]string, 0, len(msg.Elements))
	for _, elem := range msg.Elements {
		texts = append(texts, ToReadableStringEle(elem))
	}
	return texts
}

func ToReadableString(m []IMessageElement) string {
	sb := new(strings.Builder)
	for _, elem := range m {
		sb.WriteString(ToReadableStringEle(elem))
	}
	return sb.String()
}

func ToReadableStringEle(elem IMessageElement) string {
	switch e := elem.(type) {
	case *TextElement:
		return e.Content
	case *ImageElement:
		return "[图片]"
	case *AtElement:
		return e.Display
	case *ReplyElement:
		return "[回复]" // [Optional] + ToReadableString(e.Elements), 这里不破坏原义不添加
	case *FaceElement:
		return "[表情]"
	case *VoiceElement:
		return "[语音]"
	case *ShortVideoElement:
		return "[视频]"
	case *LightAppElement:
		return "[卡片消息]"
	case *ForwardMessage:
		return "[转发消息]"
	case *MarketFaceElement:
		return "[魔法表情]"
	default:
		return "[暂不支持该消息类型]"
	}
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

func ElementsHasType(elems []IMessageElement, t ElementType) bool {
	for _, elem := range elems {
		if elem.Type() == t {
			return true
		}
	}
	return false
}

func PackElementsToBody(msgElems []IMessageElement) (msgBody *message.MessageBody) {
	msgBody = &message.MessageBody{
		RichText: &message.RichText{Elems: PackElements(msgElems)},
	}
	for _, elem := range msgElems {
		bd, ok := elem.(MsgContentBuilder)
		if !ok {
			continue
		}
		msgBody.MsgContent = bd.BuildContent()
	}
	return
}

func PackElements(msgElems []IMessageElement) []*message.Elem {
	if len(msgElems) == 0 {
		return nil
	}
	elems := make([]*message.Elem, 0, len(msgElems))
	for _, elem := range msgElems {
		bd, ok := elem.(ElementBuilder)
		if !ok {
			continue
		}
		elems = append(elems, bd.BuildElement()...)
	}
	return elems
}
