package oidb

import (
	"github.com/RomiChan/protobuf/proto"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupPokeReq(groupUin, uin uint32) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0XED3_1{
		Uin:      uin,
		GroupUin: groupUin,
		Ext:      proto.Some[uint32](0),
	}
	return BuildOidbPacket(0xED3, 1, body, false, false)
}

func BuildFriendPokeReq(uin uint32) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0XED3_1{
		Uin:       uin,
		FriendUin: uin,
		Ext:       proto.Some[uint32](0),
	}
	return BuildOidbPacket(0xED3, 1, body, false, false)
}

func ParsePokeResp(data []byte) error {
	return CheckError(data)
}
