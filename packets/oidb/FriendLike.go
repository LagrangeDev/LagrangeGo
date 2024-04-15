package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
)

func BuildFriendLikeReq(uid string, count uint32) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X7E5_104{
		TargetUid: proto.Some(uid),
		Field2:    71,
		Field3:    count,
	}
	return BuildOidbPacket(0x7E5, 104, body, false, false)
}

func ParseFriendLikeResp(data []byte) (bool, error) {
	baseResp, err := ParseOidbPacket(data, nil)
	if err != nil {
		return false, err
	}
	if baseResp.ErrorCode != 0 {
		return false, errors.New(baseResp.ErrorMsg)
	}
	return true, nil
}
