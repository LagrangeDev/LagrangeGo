package client

import (
	"github.com/LagrangeDev/LagrangeGo/client/wtlogin"
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

	switch msg.Message.ContentHead.Type {
	case 166: // private msg
		return msgConverter.ParseMessageElements(&msg), nil
	case 82: // group msg
		return msgConverter.ParseMessageElements(&msg), nil
	case 141: // temp msg
		return msgConverter.ParseMessageElements(&msg), nil
	}

	return nil, nil
}

func decodeKickNTPacket(c *QQClient, pkt *wtlogin.SSOPacket) (any, error) {
	return nil, nil
}
