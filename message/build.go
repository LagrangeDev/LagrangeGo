package message

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

type ElementBuilder interface {
	BuildElement() []*message.Elem
}

func (e *TextElement) BuildElement() []*message.Elem {
	return []*message.Elem{{Text: &message.Text{Str: proto.Some(e.Content)}}}
}

func (e *AtElement) BuildElement() []*message.Elem {
	reserveData, _ := proto.Marshal(&message.MentionExtra{
		Type:   proto.Some(int32(utils.Ternary(e.TargetUin == 0, 1, 2))), // atAll
		Uin:    proto.Some(uint32(0)),
		Field5: proto.Some(int32(0)),
		Uid:    proto.Some(e.TargetUID),
	})
	return []*message.Elem{{Text: &message.Text{
		Str:       proto.Some(e.Display),
		PbReserve: reserveData,
	}}}
}

func (e *FaceElement) BuildElement() []*message.Elem {
	if e.isLargeFace {
		name, business, resultid := "", int32(1), ""
		if e.FaceID == 358 {
			name, business = "/骰子", 2
			resultid = fmt.Sprint(e.ResultID)
		} else if e.FaceID == 359 {
			name, business = "/包剪锤", 2
			resultid = fmt.Sprint(e.ResultID)
		}
		qFaceData, _ := proto.Marshal(&message.QFaceExtra{
			PackId:      proto.Some("1"),
			StickerId:   proto.Some("8"),
			Qsid:        proto.Some(int32(e.FaceID)),
			SourceType:  proto.Some(int32(1)),
			StickerType: proto.Some(business),
			ResultId:    proto.Some(resultid),
			Text:        proto.Some(name),
			RandomType:  proto.Some(int32(1)),
		})
		return []*message.Elem{{
			CommonElem: &message.CommonElem{
				ServiceType:  37,
				PbElem:       qFaceData,
				BusinessType: 1,
			},
		}}
	}
	return []*message.Elem{{
		Face: &message.Face{Index: proto.Some(int32(e.FaceID))},
	}}
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
			BusinessType: utils.Ternary(e.IsGroup, uint32(20), uint32(10)),
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
	forwardReserve := message.Preserve{
		MessageId:   uint64(e.ReplySeq),
		ReceiverUid: e.SenderUID,
	}
	forwardReserveData, err := proto.Marshal(&forwardReserve)
	if err != nil {
		return nil
	}
	return []*message.Elem{{
		SrcMsg: &message.SrcMsg{
			OrigSeqs:  []uint32{e.ReplySeq},
			SenderUin: uint64(e.SenderUin),
			Time:      proto.Some(int32(e.Time)),
			Elems:     PackElements(e.Elements),
			PbReserve: forwardReserveData,
			ToUin:     proto.Some(uint64(0)),
		},
	}}
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
	common, err := proto.Marshal(e.MsgInfo)
	if err != nil {
		return nil
	}
	return []*message.Elem{{
		CommonElem: &message.CommonElem{
			ServiceType:  48,
			PbElem:       common,
			BusinessType: 21,
		},
	}}
}

func (e *LightAppElement) BuildElement() []*message.Elem {
	return []*message.Elem{{
		LightAppElem: &message.LightAppElem{
			Data: append([]byte{0x01}, binary.ZlibCompress([]byte(e.Content))...),
		},
	}}
}

func (e *XMLElement) BuildElement() []*message.Elem {
	return []*message.Elem{{
		RichMsg: &message.RichMsg{
			ServiceId: proto.Some(int32(e.ServiceID)),
			Template1: append([]byte{0x01}, binary.ZlibCompress([]byte(e.Content))...),
		},
	}}
}

func (e *ForwardMessage) BuildElement() []*message.Elem {
	fileID := utils.NewUUID()
	extra := MultiMsgLightAppExtra{
		FileName: fileID,
		Sum:      len(e.Nodes),
	}
	extraData, err := json.Marshal(&extra)
	if err != nil {
		return nil
	}

	var news []News
	if len(e.Nodes) != 0 {
		news = make([]News, len(e.Nodes))
		for i, node := range e.Nodes {
			news[i] = News{Text: fmt.Sprintf("%s: %s", node.SenderName, ToReadableString(node.Message))}
		}
	} else {
		news = []News{{Text: "转发消息"}}
	}

	var metaSource string
	if len(e.Nodes) > 0 {
		isSenderNameExist := make(map[string]bool)
		isContainSelf := false
		isCount := 0
		for _, v := range e.Nodes {
			if v.SenderID == e.SelfID && e.SelfID > 0 {
				isContainSelf = true
			}
			if _, ok := isSenderNameExist[v.SenderName]; !ok {
				isCount++
				isSenderNameExist[v.SenderName] = true
				if metaSource == "" {
					metaSource = v.SenderName
				} else {
					metaSource += fmt.Sprintf("和%s", v.SenderName)
				}
			}
		}
		if !isContainSelf || (isCount > 2 && isCount < 1) {
			metaSource = "群聊的聊天记录"
		} else {
			metaSource += "的聊天记录"
		}
	} else {
		metaSource = "聊天记录"
	}

	content := MultiMsgLightApp{
		App: "com.tencent.multimsg",
		Config: Config{
			Autosize: 1,
			Forward:  1,
			Round:    1,
			Type:     "normal",
			Width:    300,
		},
		Desc:  "[聊天记录]",
		Extra: utils.B2S(extraData),
		Meta: Meta{
			Detail: Detail{
				News:    news,
				Resid:   e.ResID,
				Source:  metaSource,
				Summary: fmt.Sprintf("查看%d条转发消息", len(e.Nodes)),
				UniSeq:  fileID,
			},
		},
		Prompt: "[聊天记录]",
		Ver:    "0.0.0.5",
		View:   "contact",
	}

	contentData, err := json.Marshal(&content)
	if err != nil {
		return nil
	}
	return NewLightApp(utils.B2S(contentData)).BuildElement()
}

type MsgContentBuilder interface {
	BuildContent() []byte
}

func (e *FileElement) BuildContent() []byte {
	content, _ := proto.Marshal(
		&message.FileExtra{
			File: &message.NotOnlineFile{
				FileSize:   proto.Int64(int64(e.FileSize)),
				FileType:   proto.Int32(0),
				FileUuid:   proto.String(e.FileUUID),
				FileMd5:    e.FileMd5,
				FileName:   proto.String(e.FileName),
				Subcmd:     proto.Int32(1),
				DangerEvel: proto.Int32(0),
				ExpireTime: proto.Int32(int32(time.Now().Add(7 * 24 * time.Hour).Unix())),
				FileHash:   proto.String(e.FileHash),
			},
		})
	return content
}

func (e *MarketFaceElement) BuildContent() []*message.Elem {
	mFace := &message.MarketFace{
		FaceName:    proto.String(e.Summary),
		ItemType:    proto.Uint32(e.ItemType),
		FaceInfo:    proto.Uint32(1),
		FaceId:      e.FaceID,
		TabId:       proto.Uint32(e.TabID),
		SubType:     proto.Uint32(e.SubType),
		Key:         e.EncryptKey,
		MediaType:   proto.Uint32(e.MediaType),
		ImageWidth:  proto.Uint32(300),
		ImageHeight: proto.Uint32(300),
		MobileParam: utils.S2B(e.MagicValue),
	}
	mFace.PbReserve, _ = proto.Marshal(&message.MarketFacePbReserve{Field8: 1})

	return []*message.Elem{{
		MarketFace: mFace,
	}, {
		Text: &message.Text{Str: proto.Some(e.Summary)},
	}}
}
