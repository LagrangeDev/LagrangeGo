package client

import (
	"errors"
	"runtime/debug"

	"github.com/RomiChan/protobuf/proto"

	eventConverter "github.com/LagrangeDev/LagrangeGo/event"
	msgConverter "github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/packets/wtlogin"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

var listeners = map[string]func(*QQClient, *wtlogin.SSOPacket) (any, error){
	"trpc.msg.olpush.OlPushService.MsgPush":            decodeOlPushServicePacket,
	"trpc.qq_new_tech.status_svc.StatusService.KickNT": decodeKickNTPacket,
}

func decodeOlPushServicePacket(c *QQClient, pkt *wtlogin.SSOPacket) (any, error) {
	msg := message.PushMsg{}
	err := proto.Unmarshal(pkt.Data, &msg)
	if err != nil {
		return nil, err
	}
	pkg := msg.Message
	typ := pkg.ContentHead.Type
	defer func() {
		if r := recover(); r != nil {
			networkLogger.Errorf("Recovered from panic: %v\n%s", r, debug.Stack())
			networkLogger.Errorf("protobuf data: %x", pkt.Data)
		}
	}()
	if pkg.Body == nil {
		return nil, errors.New("message body is empty")
	}

	switch typ {
	case 166: // private msg
		return msgConverter.ParsePrivateMessage(&msg), nil
	case 82: // group msg
		return msgConverter.ParseGroupMessage(&msg), nil
	case 141: // temp msg
		return msgConverter.ParseTempMessage(&msg), nil
	case 33: // member increase
		pb := message.GroupChange{}
		err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		return eventConverter.ParseMemberIncreaseEvent(&pb), nil
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
		return eventConverter.ParseMemberDecreaseEvent(&pb), nil
	case 84: // group request join notice
		pb := message.GroupJoin{}
		err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		return eventConverter.ParseRequestJoinNotice(&pb), nil
	case 525: // group request invitation notice
		pb := message.GroupInvitation{}
		err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		return eventConverter.ParseRequestInvitationNotice(&pb), nil
	case 87: // group invite notice
		pb := message.GroupInvite{}
		err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		return eventConverter.ParseInviteNotice(&pb), nil
	case 0x210: // friend event, 528
		subType := pkg.ContentHead.SubType.Unwrap()
		switch subType {
		case 35: // friend request notice
			pb := message.FriendRequest{}
			err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
			if err != nil {
				return nil, err
			}
			return eventConverter.ParseFriendRequestNotice(&msg, &pb), nil
		case 138: // friend recall
			pb := message.FriendRecall{}
			err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
			if err != nil {
				return nil, err
			}
			return eventConverter.ParseFriendRecallEvent(&pb), nil
		case 39: // friend rename
			networkLogger.Info("friend rename")
			pb := message.FriendRenameMsg{}
			err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
			if err != nil {
				return nil, err
			}
			return eventConverter.ParseFriendRenameEvent(&pb, c.cache), nil
		case 29:
			networkLogger.Info("self rename")
			pb := message.SelfRenameMsg{}
			err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
			if err != nil {
				return nil, err
			}
			return eventConverter.ParseSelfRenameEvent(&pb, c.sig), nil
		default:
			networkLogger.Warningf("unknown subtype %d of type 0x210, proto data: %x", subType, pkg.Body.MsgContent)
		}
	case 0x2DC: // grp event, 732
		subType := pkg.ContentHead.SubType.Unwrap()
		switch subType {
		case 20: // nudget(grp_id only)
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
			return eventConverter.ParseGroupRecallEvent(&pb), nil
		case 12: // mute
			pb := message.GroupMute{}
			err = proto.Unmarshal(pkg.Body.MsgContent, &pb)
			if err != nil {
				return nil, err
			}
			return eventConverter.ParseGroupMuteEvent(&pb), nil
		default:
			networkLogger.Warningf("Unsupported group event, subType: %v", subType)
		}
	default:
		networkLogger.Warningf("Unsupported message type: %v", typ)
	}

	return nil, nil
}

func decodeKickNTPacket(c *QQClient, pkt *wtlogin.SSOPacket) (any, error) {
	return nil, nil
}
