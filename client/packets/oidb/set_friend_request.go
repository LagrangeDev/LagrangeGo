package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils/io"
)

func BuildSetFriendRequest(accept bool, targetUID string) (*Packet, error) {
	body := oidb.OidbSvcTrpcTcp0XB5D_44{
		Accept:    io.Ternary[uint32](accept, 3, 5),
		TargetUid: targetUID,
	}
	return BuildOidbPacket(0xb5d, 44, &body, false, false)
}

func ParseSetFriendRequestResp(data []byte) error {

	return CheckError(data)
}
