package client

import (
	"sync/atomic"

	message2 "github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/action"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	"github.com/RomiChan/protobuf/proto"
)

func (c *QQClient) SendRawMessage(route *message.RoutingHead, body *message.MessageBody) (resp *action.SendMessageResponse, err error) {
	msg := &message.Message{
		RoutingHead: route,
		ContentHead: &message.ContentHead{
			Type:    1,
			SubType: proto.Some(uint32(0)),
			DivSeq:  proto.Some(uint32(0)),
		},
		Body: body,
		Seq:  proto.Some(atomic.LoadUint32(&c.sig.Sequence)),
		Rand: proto.Some(crypto.RandU32()),
	}
	// grp_id not null
	if (route.Grp != nil && route.Grp.GroupCode.IsSome()) || (route.GrpTmp != nil && route.GrpTmp.GroupUin.IsSome()) {
		msg.Ctrl = &message.MessageControl{MsgFlag: int32(utils.TimeStamp())}
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return
	}

	packet, err := c.SendUniPacketAndAwait("MessageSvc.PbSendMsg", data)
	if err != nil {
		return
	}
	resp = &action.SendMessageResponse{}
	err = proto.Unmarshal(packet.Data, resp)
	if err != nil {
		return
	}
	return
}

func (c *QQClient) SendGroupMessage(groupUin uint32, elements []message2.IMessageElement) (resp *action.SendMessageResponse, err error) {
	elements = preprocessGroupMessage(c, groupUin, elements)
	body := message2.BuildMessageElements(elements)
	route := &message.RoutingHead{
		Grp: &message.Grp{GroupCode: proto.Some(groupUin)},
	}
	return c.SendRawMessage(route, body)
}

func (c *QQClient) SendPrivateMessage(uin uint32, elements []message2.IMessageElement) (resp *action.SendMessageResponse, err error) {
	elements = preprocessPrivateMessage(c, uin, elements)
	body := message2.BuildMessageElements(elements)
	route := &message.RoutingHead{
		C2C: &message.C2C{
			Uid: proto.Some(c.GetUid(uin)),
		},
	}
	return c.SendRawMessage(route, body)
}

func (c *QQClient) SendTempMessage(groupID uint32, uin uint32, elements []message2.IMessageElement) (resp *action.SendMessageResponse, err error) {
	body := message2.BuildMessageElements(elements)
	route := &message.RoutingHead{
		GrpTmp: &message.GrpTmp{
			GroupUin: proto.Some(groupID),
			ToUin:    proto.Some(uin),
		},
	}
	return c.SendRawMessage(route, body)
}

func preprocessGroupMessage(client *QQClient, groupUin uint32, elements []message2.IMessageElement) []message2.IMessageElement {
	for _, element := range elements {
		switch elem := element.(type) {
		case *message2.AtElement:
			member, _ := client.GetCachedMemberInfo(elem.Target, groupUin)
			if member != nil {
				elem.UID = member.Uid
				if member.MemberCard != "" {
					elem.Display = "@" + member.MemberCard
				} else {
					elem.Display = "@" + member.MemberName
				}
			}
		case *message2.GroupImageElement:
			_, err := client.ImageUploadGroup(groupUin, elem)
			if err != nil {
				networkLogger.Errorln(err)
			}
			if elem.MsgInfo == nil {
				networkLogger.Errorln("ImageUploadGroup failed")
				continue
			}
		default:
		}
	}
	return elements
}

func preprocessPrivateMessage(client *QQClient, targetUin uint32, elements []message2.IMessageElement) []message2.IMessageElement {
	for _, element := range elements {
		switch elem := element.(type) {
		case *message2.FriendImageElement:
			targetUid := client.GetUid(targetUin)
			_, err := client.ImageUploadPrivate(targetUid, elem)
			if err != nil {
				networkLogger.Errorln(err)
			}
			if elem.MsgInfo == nil {
				networkLogger.Errorln("ImageUploadPrivate failed")
				continue
			}
		default:
		}
	}
	return elements
}
