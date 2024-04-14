package client

import (
	"runtime/debug"

	"github.com/LagrangeDev/LagrangeGo/packets/wtlogin"

	msgConverter "github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/message"
	"github.com/RomiChan/protobuf/proto"
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

	switch msg.Message.ContentHead.Type {
	case 166: // private msg
		return msgConverter.ParsePrivateMessage(&msg), nil
	case 82: // group msg
		return msgConverter.ParseGroupMessage(&msg), nil
	case 141: // temp msg
		return msgConverter.ParseTempMessage(&msg), nil
	default:
		networkLogger.Warningf("Unsupported message type: %v", msg.Message.ContentHead.Type)
	}

	return nil, nil
}

func decodeKickNTPacket(c *QQClient, pkt *wtlogin.SSOPacket) (any, error) {
	return nil, nil
}
