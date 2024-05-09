package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/pkg/pb/service/oidb"
)

func BuildGroupRenameReq(groupUin uint32, name string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X89A_15{
		GroupUin: groupUin,
		Body:     &oidb.OidbSvcTrpcTcp0X89A_15Body{TargetName: name},
	}
	return BuildOidbPacket(0x89A, 15, body, false, false)
}

func ParseGroupRenameResp(data []byte) error {
	return CheckError(data)
}
