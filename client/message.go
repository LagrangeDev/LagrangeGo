package client

import (
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/action"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	message2 "github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	"github.com/RomiChan/protobuf/proto"
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
		msg.Ctrl = &message.MessageControl{MsgFlag: int32(utils.TimeStamp())}
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
		Id:         ret.GroupSequence.Unwrap(),
		InternalId: mr,
		GroupUin:   groupUin,
		GroupName:  group.GroupName,
		Sender: &message2.Sender{
			Uin:           c.Uin,
			Uid:           c.GetUid(c.Uin),
			Nickname:      c.NickName(),
			CardName:      minfo.MemberCard,
			AnonymousInfo: nil,
			IsFriend:      true,
		},
		Time:           ret.Timestamp1,
		Elements:       elements,
		OriginalObject: nil,
	}
	return resp, nil
}

// SendPrivateMessage 发送群聊消息，默认会对消息进行预处理
func (c *QQClient) SendPrivateMessage(uin uint32, elements []message2.IMessageElement, needPreprocess ...bool) (*message2.PrivateMessage, error) {
	if needPreprocess == nil || needPreprocess[0] {
		elements = c.preProcessPrivateMessage(uin, elements)
	}
	body := message2.PackElementsToBody(elements)
	route := &message.RoutingHead{
		C2C: &message.C2C{
			Uid: proto.Some(c.GetUid(uin)),
		},
	}
	mr := crypto.RandU32()
	ret, clientSeq, err := c.SendRawMessage(route, body, mr)
	if err != nil || ret.PrivateSequence == 0 {
		return nil, err
	}
	resp := &message2.PrivateMessage{
		Id:         ret.PrivateSequence,
		InternalId: mr,
		ClientSeq:  clientSeq,
		Self:       c.Uin,
		Target:     uin,
		Time:       ret.Timestamp1,
		Sender: &message2.Sender{
			Uin:           c.Uin,
			Uid:           c.GetUid(c.Uin),
			Nickname:      c.NickName(),
			AnonymousInfo: nil,
			IsFriend:      true,
		},
		Elements: nil,
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
		Id:        ret.PrivateSequence,
		GroupUin:  groupUin,
		GroupName: group.GroupName,
		Self:      c.Uin,
		Sender: &message2.Sender{
			Uin:           c.Uin,
			Uid:           c.GetUid(c.Uin),
			Nickname:      c.NickName(),
			AnonymousInfo: nil,
			IsFriend:      true,
		},
		Elements: elements,
	}
	return resp, nil
}

// BuildFakeMessage make a fake message
func (c *QQClient) BuildFakeMessage(msgElems []*message2.ForwardNode) []*message.PushMsgBody {
	body := make([]*message.PushMsgBody, len(msgElems))
	for idx, elem := range msgElems {
		avatar := fmt.Sprintf("https://q.qlogo.cn/headimg_dl?dst_uin=%d&spec=640&img_type=jpg", elem.SenderId)
		body[idx] = &message.PushMsgBody{
			ResponseHead: &message.ResponseHead{
				FromUid: proto.String(""),
				FromUin: elem.SenderId,
			},
			ContentHead: &message.ContentHead{
				Type:      uint32(utils.Ternary(elem.GroupId != 0, 82, 9)),
				MsgId:     proto.Uint32(crypto.RandU32()),
				Sequence:  proto.Uint32(crypto.RandU32()),
				TimeStamp: proto.Uint32(uint32(utils.TimeStamp())),
				Field7:    proto.Uint64(1),
				Field8:    proto.Uint32(0),
				Field9:    proto.Uint32(0),
				Foward: &message.ForwardHead{
					Field1:        proto.Uint32(0),
					Field2:        proto.Uint32(0),
					Field3:        utils.Ternary(elem.GroupId != 0, proto.Uint32(0), proto.Uint32(2)),
					UnknownBase64: proto.String(avatar),
					Avatar:        proto.String(avatar),
				},
			},
		}
		if elem.GroupId != 0 {
			body[idx].ResponseHead.Grp = &message.ResponseGrp{
				GroupUin:   elem.GroupId,
				MemberName: elem.SenderName,
				Unknown5:   2,
			}
			c.preProcessGroupMessage(elem.GroupId, elem.Message)
		} else {
			body[idx].ResponseHead.ToUid = proto.String(c.GetUid(c.Uin))
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
				elem.TargetUid = member.Uid
				if member.MemberCard != "" {
					elem.Display = "@" + member.MemberCard
				} else {
					elem.Display = "@" + member.MemberName
				}
			}
		case *message2.ImageElement:
			_, err := c.ImageUploadGroup(groupUin, elem)
			if err != nil {
				c.errorln(err)
			}
			if elem.MsgInfo == nil {
				c.errorln("ImageUploadGroup failed")
				continue
			}
		case *message2.VoiceElement:
			_, err := c.RecordUploadGroup(groupUin, elem)
			if err != nil {
				c.errorln(err)
			}
			if elem.MsgInfo == nil {
				c.errorln("RecordUploadGroup failed")
				continue
			}
		case *message2.ShortVideoElement:
			_, err := c.VideoUploadGroup(groupUin, elem)
			if err != nil {
				c.errorln(err)
			}
			if elem.MsgInfo == nil {
				c.errorln("VideoUploadGroup failed")
				continue
			}
		case *message2.ForwardMessage:
			if elem.ResID != "" && len(elem.Nodes) == 0 {
				forward, _ := c.FetchForwardMsg(elem.ResID)
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
			targetUid := c.GetUid(targetUin)
			_, err := c.ImageUploadPrivate(targetUid, elem)
			if err != nil {
				c.errorln(err)
				continue
			}
			if elem.MsgInfo == nil {
				c.errorln("ImageUploadPrivate failed")
				continue
			}
		case *message2.VoiceElement:
			targetUid := c.GetUid(targetUin)
			_, err := c.RecordUploadPrivate(targetUid, elem)
			if err != nil {
				c.errorln(err)
				continue
			}
			if elem.MsgInfo == nil {
				c.errorln("RecordUploadPrivate failed")
				continue
			}
		case *message2.ShortVideoElement:
			targetUid := c.GetUid(targetUin)
			_, err := c.VideoUploadPrivate(targetUid, elem)
			if err != nil {
				c.errorln(err)
				continue
			}
			if elem.MsgInfo == nil {
				c.errorln("VideoUploadPrivate failed")
				continue
			}
		case *message2.ForwardMessage:
			if elem.ResID != "" && len(elem.Nodes) == 0 {
				forward, _ := c.FetchForwardMsg(elem.ResID)
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
