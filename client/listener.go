package client

import (
	"errors"
	"runtime/debug"

	"github.com/RomiChan/protobuf/proto"

	eventConverter "github.com/LagrangeDev/LagrangeGo/event"
	msgConverter "github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/status"
	"github.com/LagrangeDev/LagrangeGo/packets/wtlogin"
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

	defer func() {
		if r := recover(); r != nil {
			networkLogger.Errorf("Recovered from panic: %v\n%s", r, debug.Stack())
			networkLogger.Errorf("protobuf data: %x", pkt.Data)
		}
	}()
	if msg.Message.Body == nil {
		return nil, errors.New("message body is empty")
	}

	switch msg.Message.ContentHead.Type {
	case 166: // private msg
		return msgConverter.ParsePrivateMessage(&msg), nil
	case 82: // group msg
		return msgConverter.ParseGroupMessage(&msg), nil
	case 141: // temp msg
		return msgConverter.ParseTempMessage(&msg), nil
	case 33: // member joined
		pb := status.MemberChanged{}
		err := proto.Unmarshal(msg.Message.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		return eventConverter.ParseMemberJoined(&msg, &pb), nil
	case 34: // member exit
		pb := status.MemberChanged{}
		err := proto.Unmarshal(msg.Message.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		return eventConverter.ParseMemberQuit(&msg, &pb), nil
	case 84:
		pb := status.MemberJoinRequest{}
		err := proto.Unmarshal(msg.Message.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		return eventConverter.ParseMemberJoinRequest(&pb), nil
	case 525:
		pb := status.MemberInviteRequest{}
		err := proto.Unmarshal(msg.Message.Body.MsgContent, &pb)
		if err != nil {
			return nil, err
		}
		return eventConverter.ParseMemberJoinRequestFromInvite(&pb), nil
	case 0x2DC: //  grp event, 732
		subType := msg.Message.ContentHead.SubType.Unwrap()
		switch subType {
		case 12: // mute
			pb := map[int]any{}
			err := proto.Unmarshal(msg.Message.Body.MsgContent, &pb)
			if err != nil {
				return nil, err
			}
			return eventConverter.ParseGroupMuteEvent(&pb), nil
		default:
			networkLogger.Warningf("Unsupported group event, subType: %v", subType)
		}
	default:
		networkLogger.Warningf("Unsupported message type: %v", msg.Message.ContentHead.Type)
	}

	return nil, nil
}

func decodeKickNTPacket(c *QQClient, pkt *wtlogin.SSOPacket) (any, error) {
	return nil, nil
}
