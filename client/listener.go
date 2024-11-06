package client

import (
	"errors"
	"runtime/debug"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/system"

	"github.com/LagrangeDev/LagrangeGo/internal/proto"

	eventConverter "github.com/LagrangeDev/LagrangeGo/client/event"
	"github.com/LagrangeDev/LagrangeGo/client/internal/network"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	msgConverter "github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

// decoders https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/client.go#L150
var decoders = map[string]func(*QQClient, *network.Packet) (any, error){
	"trpc.msg.olpush.OlPushService.MsgPush":            decodeOlPushServicePacket,
	"trpc.qq_new_tech.status_svc.StatusService.KickNT": decodeKickNTPacket,
}

func decodeOlPushServicePacket(c *QQClient, pkt *network.Packet) (any, error) {
	msg := message.PushMsg{}
	err := proto.Unmarshal(pkt.Payload, &msg)
	if err != nil {
		return nil, err
	}
	pkg := msg.Message
	typ := pkg.ContentHead.Type
	defer func() {
		if r := recover(); r != nil {
			c.error("recovered from panic: %v\n%s", r, debug.Stack())
			c.error("protobuf data: %x", pkt.Payload)
		}
	}()
	if pkg.Body == nil {
		return nil, errors.New("message body is empty")
	}
	switch typ {
	case 166, 208, 529: // 166 for private msg, 208 for private record, 529 for private file
		prvMsg := msgConverter.ParsePrivateMessage(pkg)
		c.PreprocessPrivateMessageEvent(prvMsg)
		if prvMsg.Sender.Uin != c.Uin {
			c.PrivateMessageEvent.dispatch(c, prvMsg)
		} else {
			c.SelfPrivateMessageEvent.dispatch(c, prvMsg)
		}
		return nil, nil
	case 82: // group msg
		grpMsg := msgConverter.ParseGroupMessage(pkg)
		c.PreprocessGroupMessageEvent(grpMsg)
		if grpMsg.Sender.Uin != c.Uin {
			c.GroupMessageEvent.dispatch(c, grpMsg)
		} else {
			c.SelfGroupMessageEvent.dispatch(c, grpMsg)
		}
		return nil, nil
	case 141: // temp msg
		tempMsg := msgConverter.ParseTempMessage(pkg)
		if tempMsg.Sender.Uin != c.Uin {
			c.TempMessageEvent.dispatch(c, tempMsg)
		} else {
			c.SelfTempMessageEvent.dispatch(c, tempMsg)
		}
		return nil, nil
	case 33: // member increase
		pb := message.GroupChange{}
		err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		ev := eventConverter.ParseMemberIncreaseEvent(&pb)
		_ = c.PreprocessOther(ev)
		if ev.MemberUin == c.Uin { // bot 进群
			_ = c.RefreshAllGroupsInfo()
			c.GroupJoinEvent.dispatch(c, ev)
		} else {
			_ = c.RefreshGroupMemberCache(ev.GroupUin, ev.MemberUin)
			c.GroupMemberJoinEvent.dispatch(c, ev)
		}
		return nil, nil
	case 34: // member decrease
		pb := message.GroupChange{}
		err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		// 3 是bot自身被踢出，Operator字段会是一个protobuf
		if pb.DecreaseType == 3 && pb.Operator != nil {
			Operator := message.OperatorInfo{}
			err = proto.Unmarshal(pb.Operator, &Operator)
			if err != nil {
				return nil, err
			}
			pb.Operator = utils.S2B(Operator.OperatorField1.OperatorUid)
		}
		ev := eventConverter.ParseMemberDecreaseEvent(&pb)
		_ = c.PreprocessOther(ev)
		if ev.MemberUin == c.Uin {
			c.GroupLeaveEvent.dispatch(c, ev)
		} else {
			c.GroupMemberLeaveEvent.dispatch(c, ev)
		}
		return nil, nil
	case 44: // group admin changed
		pb := message.GroupAdmin{}
		err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		ev := eventConverter.ParseGroupMemberPermissionChanged(&pb)
		_ = c.PreprocessOther(ev)
		_ = c.RefreshGroupMemberCache(ev.GroupUin, ev.TargetUin)
		c.GroupMemberPermissionChangedEvent.dispatch(c, ev)
	case 84: // group request join notice
		pb := message.GroupJoin{}
		err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		ev := eventConverter.ParseRequestJoinNotice(&pb)
		_ = c.PreprocessOther(ev)
		user, _ := c.FetchUserInfo(ev.TargetUid)
		if user != nil {
			ev.TargetUin = user.Uin
			ev.TargetNick = user.Nickname
		}
		commonRequests, reqErr := c.GetGroupSystemMessages(false, 20, ev.GroupUin)
		filteredRequests, freqErr := c.GetGroupSystemMessages(true, 20, ev.GroupUin)
		if reqErr == nil && freqErr == nil {
			for _, request := range append(commonRequests.JoinRequests, filteredRequests.JoinRequests...) {
				if request.TargetUid == ev.TargetUid && !request.Checked() {
					ev.RequestSeq = request.Sequence
					ev.Answer = request.Comment
				}
			}
		}
		c.GroupMemberJoinRequestEvent.dispatch(c, ev)
		return nil, nil
	case 525: // group request invitation notice
		pb := message.GroupInvitation{}
		err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		if pb.Cmd != 87 {
			return nil, nil
		}
		ev := eventConverter.ParseRequestInvitationNotice(&pb)
		_ = c.PreprocessOther(ev)
		user, _ := c.FetchUserInfo(ev.TargetUid)
		if user != nil {
			ev.TargetUin = user.Uin
			ev.TargetNick = user.Nickname
		}
		c.GroupMemberJoinRequestEvent.dispatch(c, ev)
		return nil, nil
	case 87: // group invite notice
		pb := message.GroupInvite{}
		err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		ev := eventConverter.ParseInviteNotice(&pb)
		group, err := c.FetchGroupInfo(ev.GroupUin, true)
		if err == nil {
			ev.GroupName = group.GroupName
		}
		_ = c.PreprocessOther(ev)
		user, _ := c.FetchUserInfo(ev.InvitorUid)
		if user != nil {
			ev.InvitorUin = user.Uin
			ev.InvitorNick = user.Nickname
		}
		commonRequests, reqErr := c.GetGroupSystemMessages(false, 20, ev.GroupUin)
		filteredRequests, freqErr := c.GetGroupSystemMessages(true, 20, ev.GroupUin)
		if reqErr == nil && freqErr == nil {
			for _, request := range append(commonRequests.InvitedRequests, filteredRequests.InvitedRequests...) {
				if !request.Checked() {
					ev.RequestSeq = request.Sequence
					break
				}
			}
		}
		c.GroupInvitedEvent.dispatch(c, ev)
		return nil, nil
	case 0x210: // friend event, 528
		subType := pkg.ContentHead.SubType.Unwrap()
		switch subType {
		case 35: // friend request notice
			pb := message.FriendRequest{}
			err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
			if err != nil {
				return nil, err
			}
			if pb.Info == nil {
				break
			}
			ev := eventConverter.ParseFriendRequestNotice(&pb)
			user, _ := c.FetchUserInfo(ev.SourceUid)
			if user != nil {
				ev.SourceUin = user.Uin
				ev.SourceNick = user.Nickname
			}
			c.NewFriendRequestEvent.dispatch(c, ev)
			return nil, nil
		case 138: // friend recall
			pb := message.FriendRecall{}
			err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
			if err != nil {
				return nil, err
			}
			ev := eventConverter.ParseFriendRecallEvent(&pb)
			_ = c.PreprocessOther(ev)
			c.FriendRecallEvent.dispatch(c, ev)
			return nil, nil
		case 39: // friend rename
			pb := message.FriendRenameMsg{}
			err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
			if err != nil {
				return nil, err
			}
			if pb.Body.Field2 == 20 { // friend name update
				ev := eventConverter.ParseFriendRenameEvent(&pb)
				_ = c.PreprocessOther(ev)
				c.RenameEvent.dispatch(c, ev)
			} // 40 grp name
			return nil, nil
		case 29:
			pb := message.SelfRenameMsg{}
			err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
			if err != nil {
				return nil, err
			}
			c.RenameEvent.dispatch(c, eventConverter.ParseSelfRenameEvent(&pb, &c.transport.Sig))
			return nil, nil
		case 290: // greyTip
			pb := message.GeneralGrayTipInfo{}
			err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
			if err != nil {
				return nil, err
			}
			if pb.BusiType == 12 {
				c.FriendNotifyEvent.dispatch(c, eventConverter.ParsePokeEvent(&pb))
			}
		default:
			c.warning("unknown subtype %d of type 0x210, proto data: %x", subType, pkg.Body.MsgContent)
		}
	case 0x2DC: // grp event, 732
		subType := pkg.ContentHead.SubType.Unwrap()
		switch subType {
		case 21: // set essence
			reader := binary.NewReader(pkg.Body.MsgContent)
			_ = reader.ReadU32() // group uin
			reader.SkipBytes(1)  // unknown byte
			pb := message.NotifyMessageBody{}
			err = proto.Unmarshal(reader.ReadBytesWithLength("u16", false), &pb)
			if err != nil {
				return nil, err
			}
			c.GroupDigestEvent.dispatch(c, eventConverter.ParseGroupDigestEvent(&pb))
			return nil, nil
		case 20: // group greyTip
			reader := binary.NewReader(pkg.Body.MsgContent)
			groupUin := reader.ReadU32() // group uin
			reader.SkipBytes(1)          // unknown byte
			pb := message.NotifyMessageBody{}
			err = proto.Unmarshal(reader.ReadBytesWithLength("u16", false), &pb)
			if err != nil {
				return nil, err
			}
			if pb.GrayTipInfo.BusiType == 12 { // poke
				c.GroupNotifyEvent.dispatch(c, eventConverter.ParseGroupPokeEvent(&pb, groupUin))
			}
			return nil, nil
		case 17: // recall
			reader := binary.NewReader(pkg.Body.MsgContent)
			_ = reader.ReadU32() // group uin
			_ = reader.ReadU8()  // reserve
			pb := message.NotifyMessageBody{}
			err = proto.Unmarshal(reader.ReadBytesWithLength("u16", false), &pb)
			if err != nil {
				return nil, err
			}
			ev := eventConverter.ParseGroupRecallEvent(&pb)
			_ = c.PreprocessOther(ev)
			c.GroupRecallEvent.dispatch(c, ev)
			return nil, nil
		case 16: // group name update & member special title update & group reaction
			reader := binary.NewReader(pkg.Body.MsgContent)
			groupUin := reader.ReadU32()
			reader.SkipBytes(1)
			pb := message.NotifyMessageBody{}
			err = proto.Unmarshal(reader.ReadBytesWithLength("u16", false), &pb)
			if err != nil {
				return nil, err
			}
			switch pb.Field13 {
			case 6: // GroupMemberSpecialTitle
				epb := message.GroupSpecialTitle{}
				err := proto.Unmarshal(pb.EventParam, &epb)
				if err != nil {
					return nil, err
				}
				c.MemberSpecialTitleUpdatedEvent.dispatch(c, eventConverter.ParseGroupMemberSpecialTitleUpdatedEvent(&epb, groupUin))
			case 12: // group name update
				r := binary.NewReader(pb.EventParam)
				r.SkipBytes(3)
				ev := eventConverter.ParseGroupNameUpdatedEvent(&pb, string(r.ReadBytesWithLength("u8", false)))
				_ = c.PreprocessOther(ev)
				c.GroupNameUpdatedEvent.dispatch(c, ev)
			case 35: // group reaction
				r := binary.NewReader(pkg.Body.MsgContent)
				r.ReadU32()
				r.ReadBytes(1)
				rpb := message.GroupReaction{}
				err := proto.Unmarshal(r.ReadBytesWithLength("u16", false), &rpb)
				if err != nil {
					return nil, err
				}
				ev := eventConverter.ParseGroupReactionEvent(&rpb)
				_ = c.PreprocessOther(ev)
				c.GroupReactionEvent.dispatch(c, ev)
			}
		case 12: // mute
			pb := message.GroupMute{}
			err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
			if err != nil {
				return nil, err
			}
			ev := eventConverter.ParseGroupMuteEvent(&pb)
			_ = c.PreprocessOther(ev)
			c.GroupMuteEvent.dispatch(c, ev)
			return nil, nil
		default:
			c.warning("Unsupported group event, subType: %v, proto data: %x", subType, pkg.Body.MsgContent)
		}
	default:
		c.warning("Unsupported message type: %v, proto data: %x", typ, pkg.Body.MsgContent)
	}

	return nil, nil
}

func decodeKickNTPacket(c *QQClient, pkt *network.Packet) (any, error) {
	pb := system.ServiceKickNTResponse{}
	_ = proto.Unmarshal(pkt.Payload, &pb)
	c.KickedEvent.dispatch(c, eventConverter.ParseKickedEvent(&pb))
	return nil, nil
}

func (c *QQClient) PreprocessGroupMessageEvent(msg *msgConverter.GroupMessage) {
	for _, elem := range msg.Elements {
		switch e := elem.(type) {
		case *msgConverter.ImageElement:
			if e.Url != "" {
				continue
			}
			url, _ := c.GetGroupImageUrl(msg.GroupUin, e.MsgInfo.MsgInfoBody[0].Index)
			e.Url = url
		case *msgConverter.VoiceElement:
			url, _ := c.GetGroupRecordUrl(msg.GroupUin, e.Node)
			e.Url = url
		case *msgConverter.ShortVideoElement:
			url, err := c.GetVideoUrl(true, e)
			if err != nil {
				continue
			}
			e.Url = url
		case *msgConverter.FileElement:
			url, _ := c.GetGroupFileUrl(msg.GroupUin, e.FileId)
			e.FileUrl = url
		case *msgConverter.ForwardMessage:
			if e.Nodes == nil {
				if forward, err := c.FetchForwardMsg(e.ResID); err != nil {
					continue
				} else {
					e.Nodes = forward.Nodes
				}
			}
		}
	}
}

func (c *QQClient) PreprocessPrivateMessageEvent(msg *msgConverter.PrivateMessage) {
	if friend := c.GetCachedFriendInfo(msg.Sender.Uin); friend != nil {
		msg.Sender.Nickname = friend.Nickname
	}
	for _, elem := range msg.Elements {
		switch e := elem.(type) {
		case *msgConverter.ImageElement:
			if e.Url != "" {
				continue
			}
			url, _ := c.GetPrivateImageUrl(e.MsgInfo.MsgInfoBody[0].Index)
			e.Url = url
		case *msgConverter.VoiceElement:
			url, err := c.GetPrivateRecordUrl(e.Node)
			if err != nil {
				continue
			}
			e.Url = url
		case *msgConverter.ShortVideoElement:
			url, err := c.GetVideoUrl(false, e)
			if err != nil {
				continue
			}
			e.Url = url
		case *msgConverter.FileElement:
			url, err := c.GetPrivateFileUrl(e.FileUUID, e.FileHash)
			if err != nil {
				continue
			}
			e.FileUrl = url
		case *msgConverter.ForwardMessage:
			if e.Nodes == nil {
				if forward, err := c.FetchForwardMsg(e.ResID); err != nil {
					continue
				} else {
					e.Nodes = forward.Nodes
				}
			}
		}
	}
}

func (c *QQClient) PreprocessOther(g eventConverter.CanPreprocess) error {
	g.ResolveUin(func(uid string, groupUin ...uint32) uint32 {
		return c.GetUin(uid, groupUin...)
	})
	return nil
}
