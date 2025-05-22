package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildSetGroupNameReq(groupUin uint32, name string) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0X89A_15{
		GroupUin: groupUin,
		Body:     &oidb.OidbSvcTrpcTcp0X89A_15Body{TargetName: name},
	}
	return BuildOidbPacket(0x89A, 15, body, false, false)
}

func ParseSetGroupNameResp(data []byte) error {
	return CheckError(data)
}
