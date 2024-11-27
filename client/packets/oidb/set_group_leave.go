package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildSetGroupLeaveReq(groupUin uint32) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0X1097_1{GroupUin: groupUin}
	return BuildOidbPacket(0x1097, 1, body, false, false)
}

func ParseSetGroupLeaveResp(data []byte) error {
	return CheckError(data)
}
