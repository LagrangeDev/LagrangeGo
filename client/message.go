package client

import (
	"fmt"

	"github.com/RomiChan/protobuf/proto"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/action"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	message2 "github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	"github.com/LagrangeDev/LagrangeGo/utils/io"
)

func (c *QQClient) SendRawMessage(route *message.RoutingHead, body *message.MessageBody, random uint32) (*action.SendMessageResponse, uint32, error) {
	clientSeq := c.getSequence()
	msg := &message.Message{
		RoutingHead: route,
		ContentHead: &message.ContentHead{
			Type:    1,
			SubType: proto.Some(uint32(0)),
			DivSeq:  proto.Some(uint32(0)),
		},
		Body:           body,
		ClientSequence: proto.Some(clientSeq),
		Random:         proto.Some(random),
	}
	// grp_id not null
	if (route.Grp != nil && route.Grp.GroupCode.IsSome()) || (route.GrpTmp != nil && route.GrpTmp.GroupUin.IsSome()) {
		msg.Ctrl = &message.MessageControl{MsgFlag: int32(io.TimeStamp())}
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, 0, err
	}
	packet, err := c.sendUniPacketAndWait("MessageSvc.PbSendMsg", data)
	if err != nil {
		return nil, 0, err
	}
	resp := &action.SendMessageResponse{}
	err = proto.Unmarshal(packet, resp)
	if err != nil {
		return nil, 0, err
	}
	return resp, clientSeq, err
}

// SendGroupMessage 发送群聊消息，默认会对消息进行预处理
func (c *QQClient) SendGroupMessage(groupUin uint32, elements []message2.IMessageElement, needPreprocess ...bool) (*message2.GroupMessage, error) {
	if needPreprocess == nil || needPreprocess[0] {
		elements = c.preProcessGroupMessage(groupUin, elements)
	}
	body := message2.PackElementsToBody(elements)
	route := &message.RoutingHead{
		Grp: &message.Grp{GroupCode: proto.Some(groupUin)},
	}
	mr := crypto.RandU32()
	ret, _, err := c.SendRawMessage(route, body, mr)
	if err != nil || ret.GroupSequence.IsNone() {
		return nil, err
	}
	group := c.GetCachedGroupInfo(groupUin)
	minfo := c.GetCachedMemberInfo(c.Uin, groupUin)
	resp := &message2.GroupMessage{
		ID:         ret.GroupSequence.Unwrap(),
		InternalID: mr,
		GroupUin:   groupUin,
		GroupName:  group.GroupName,
		Sender: &message2.Sender{
			Uin:      c.Uin,
			UID:      c.GetUID(c.Uin),
			Nickname: c.NickName(),
			CardName: minfo.MemberCard,
			IsFriend: true,
		},
		Time:           ret.Timestamp1,
		Elements:       elements,
		OriginalObject: nil,
	}
	return resp, nil
}

// SendPrivateMessage 发送私聊消息，默认会对消息进行预处理
func (c *QQClient) SendPrivateMessage(uin uint32, elements []message2.IMessageElement, needPreprocess ...bool) (*message2.PrivateMessage, error) {
	if needPreprocess == nil || needPreprocess[0] {
		elements = c.preProcessPrivateMessage(uin, elements)
	}
	body := message2.PackElementsToBody(elements)
	route := &message.RoutingHead{
		C2C: &message.C2C{
			Uid: proto.Some(c.GetUID(uin)),
		},
	}
	mr := crypto.RandU32()
	ret, clientSeq, err := c.SendRawMessage(route, body, mr)
	if err != nil || ret.PrivateSequence == 0 {
		return nil, err
	}
	resp := &message2.PrivateMessage{
		ID:         ret.PrivateSequence,
		InternalID: mr,
		ClientSeq:  clientSeq,
		Self:       c.Uin,
		Target:     uin,
		Time:       ret.Timestamp1,
		Sender: &message2.Sender{
			Uin:      c.Uin,
			UID:      c.GetUID(c.Uin),
			Nickname: c.NickName(),
			IsFriend: true,
		},
		Elements: elements,
	}
	return resp, nil
}

func (c *QQClient) SendTempMessage(groupUin uint32, uin uint32, elements []message2.IMessageElement) (*message2.TempMessage, error) {
	body := message2.PackElementsToBody(elements)
	route := &message.RoutingHead{
		GrpTmp: &message.GrpTmp{
			GroupUin: proto.Some(groupUin),
			ToUin:    proto.Some(uin),
		},
	}
	mr := crypto.RandU32()
	ret, _, err := c.SendRawMessage(route, body, mr)
	if err != nil || ret.PrivateSequence == 0 {
		return nil, err
	}
	group := c.GetCachedGroupInfo(groupUin)
	resp := &message2.TempMessage{
		ID:        ret.PrivateSequence,
		GroupUin:  groupUin,
		GroupName: group.GroupName,
		Self:      c.Uin,
		Sender: &message2.Sender{
			Uin:      c.Uin,
			UID:      c.GetUID(c.Uin),
			Nickname: c.NickName(),
			IsFriend: true,
		},
		Elements: elements,
	}
	return resp, nil
}

// BuildFakeMessage make a fake message
func (c *QQClient) BuildFakeMessage(msgElems []*message2.ForwardNode) []*message.PushMsgBody {
	body := make([]*message.PushMsgBody, len(msgElems))
	for idx, elem := range msgElems {
		avatar := fmt.Sprintf("https://q.qlogo.cn/headimg_dl?dst_uin=%d&spec=640&img_type=jpg", elem.SenderID)
		body[idx] = &message.PushMsgBody{
			ResponseHead: &message.ResponseHead{
				FromUid: proto.String(""),
				FromUin: elem.SenderID,
			},
			ContentHead: &message.ContentHead{
				Type:      io.Ternary[uint32](elem.GroupID != 0, 82, 9),
				MsgId:     proto.Uint32(crypto.RandU32()),
				Sequence:  proto.Uint32(crypto.RandU32()),
				TimeStamp: proto.Uint32(uint32(io.TimeStamp())),
				Field7:    proto.Uint64(1),
				Field8:    proto.Uint32(0),
				Field9:    proto.Uint32(0),
				Foward: &message.ForwardHead{
					Field1:        proto.Uint32(0),
					Field2:        proto.Uint32(0),
					Field3:        proto.Uint32(io.Ternary[uint32](elem.GroupID != 0, 0, 2)),
					UnknownBase64: proto.String(avatar),
					Avatar:        proto.String(avatar),
				},
			},
		}
		if elem.GroupID != 0 {
			body[idx].ResponseHead.Grp = &message.ResponseGrp{
				GroupUin:   elem.GroupID,
				MemberName: elem.SenderName,
				Unknown5:   2,
			}
			c.preProcessGroupMessage(elem.GroupID, elem.Message)
		} else {
			body[idx].ResponseHead.ToUid = proto.String(c.GetUID(c.Uin))
			body[idx].ResponseHead.Forward = &message.ResponseForward{
				FriendName: proto.String(elem.SenderName),
			}
			body[idx].ContentHead.SubType = proto.Uint32(4)
			body[idx].ContentHead.DivSeq = proto.Uint32(4)
			c.preProcessPrivateMessage(c.Uin, elem.Message)
		}
		body[idx].Body = message2.PackElementsToBody(elem.Message)
	}
	return body
}

func (c *QQClient) preProcessGroupMessage(groupUin uint32, elements []message2.IMessageElement) []message2.IMessageElement {
	for _, element := range elements {
		switch elem := element.(type) {
		case *message2.AtElement:
			member := c.GetCachedMemberInfo(elem.TargetUin, groupUin)
			if member != nil {
				elem.TargetUID = member.UID
				elem.Display = "@" + member.DisplayName()
			}
		case *message2.ImageElement:
			if elem.MsgInfo != nil {
				continue
			}
			_, err := c.UploadGroupImage(groupUin, elem)
			if err != nil {
				c.errorln(err)
			}
			if elem.MsgInfo == nil {
				c.errorln("UploadGroupImage failed")
				continue
			}
		case *message2.VoiceElement:
			if elem.MsgInfo != nil {
				continue
			}
			_, err := c.UploadGroupRecord(groupUin, elem)
			if err != nil {
				c.errorln(err)
			}
			if elem.MsgInfo == nil {
				c.errorln("UploadGroupRecord failed")
				continue
			}
		case *message2.ShortVideoElement:
			if elem.MsgInfo != nil {
				continue
			}
			_, err := c.UploadGroupVideo(groupUin, elem)
			if err != nil {
				c.errorln(err)
			}
			if elem.MsgInfo == nil {
				c.errorln("UploadGroupVideo failed")
				continue
			}
		case *message2.ForwardMessage:
			if elem.ResID != "" && len(elem.Nodes) == 0 {
				forward, _ := c.FetchForwardMsg(elem.ResID)
				elem.IsGroup = true
				elem.Nodes = forward.Nodes
			}
			if elem.ResID == "" && len(elem.Nodes) != 0 {
				_, err := c.UploadForwardMsg(elem, groupUin)
				if err != nil {
					c.errorln(err)
				}
			}
		default:
		}
	}
	return elements
}

func (c *QQClient) preProcessPrivateMessage(targetUin uint32, elements []message2.IMessageElement) []message2.IMessageElement {
	for _, element := range elements {
		switch elem := element.(type) {
		case *message2.ImageElement:
			if elem.MsgInfo != nil {
				continue
			}
			targetUID := c.GetUID(targetUin)
			_, err := c.UploadPrivateImage(targetUID, elem)
			if err != nil {
				c.errorln(err)
				continue
			}
			if elem.MsgInfo == nil {
				c.errorln("UploadPrivateImage failed")
				continue
			}
		case *message2.VoiceElement:
			if elem.MsgInfo != nil {
				continue
			}
			targetUID := c.GetUID(targetUin)
			_, err := c.UploadPrivateRecord(targetUID, elem)
			if err != nil {
				c.errorln(err)
				continue
			}
			if elem.MsgInfo == nil {
				c.errorln("UploadPrivateRecord failed")
				continue
			}
		case *message2.ShortVideoElement:
			if elem.MsgInfo != nil {
				continue
			}
			targetUID := c.GetUID(targetUin)
			_, err := c.UploadPrivateVideo(targetUID, elem)
			if err != nil {
				c.errorln(err)
				continue
			}
			if elem.MsgInfo == nil {
				c.errorln("UploadPrivateVideo failed")
				continue
			}
		case *message2.ForwardMessage:
			if elem.ResID != "" && len(elem.Nodes) == 0 {
				forward, _ := c.FetchForwardMsg(elem.ResID)
				elem.SelfID = c.Uin
				elem.Nodes = forward.Nodes
			}
			if elem.ResID == "" && len(elem.Nodes) != 0 {
				_, err := c.UploadForwardMsg(elem, 0)
				if err != nil {
					c.errorln(err)
				}
			}
		default:
		}
	}
	return elements
}
