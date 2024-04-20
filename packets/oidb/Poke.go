package oidb

import (
	"errors"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
	"github.com/RomiChan/protobuf/proto"
)

func BuildGroupPokeReq(groupUin, uin uint32) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0XED3_1{
		Uin:      uin,
		GroupUin: groupUin,
		Ext:      proto.Some[uint32](1),
	}
	return BuildOidbPacket(0xED3, 1, body, false, false)
}

func BuildFriendPokeReq(uin uint32) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0XED3_1{
		Uin:       uin,
		FriendUin: uin,
		Ext:       proto.Some[uint32](1),
	}
	return BuildOidbPacket(0xED3, 1, body, false, false)
}

func ParsePokeResp(data []byte) (bool, error) {
	baseResp, err := ParseOidbPacket(data, nil)
	if err != nil {
		return false, err
	}
	if baseResp.ErrorCode != 0 {
		return false, errors.New(baseResp.ErrorMsg)
	}
	return true, nil
}
