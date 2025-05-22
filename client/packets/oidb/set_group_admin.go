package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildSetGroupAdminReq(groupUin uint32, uid string, isAdmin bool) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0X1096_1{
		GroupUin: groupUin,
		Uid:      uid,
		IsAdmin:  isAdmin,
	}
	return BuildOidbPacket(0x1096, 1, body, false, false)
}

func ParseSetGroupAdminResp(data []byte) error {
	return CheckError(data)
}
