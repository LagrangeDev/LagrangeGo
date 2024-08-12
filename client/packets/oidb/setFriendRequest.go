package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

func BuildSetFriendRequest(accept bool, targetUid string) (*OidbPacket, error) {
	result := utils.Ternary[uint32](accept, 3, 5)
	packet := oidb.OidbSvcTrpcTcp0XB5D_44{
		Accept:    result, //utils.Bool2Uint32(accept, 3, 5),
		TargetUid: targetUid,
	}
	return BuildOidbPacket(0xb5d, 44, packet, false, false)
}

func ParseSetFriendRequestResp(data []byte) error {

	return CheckError(data)
}
