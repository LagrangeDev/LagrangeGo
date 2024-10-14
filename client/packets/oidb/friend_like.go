package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
)

func BuildFriendLikeReq(uid string, count uint32) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X7E5_104{
		TargetUid: proto.Some(uid),
		Source   : 71,
		Count    : count,
	}
	return BuildOidbPacket(0x7E5, 104, body, false, false)
}

func ParseFriendLikeResp(data []byte) error {
	return CheckError(data)
}
