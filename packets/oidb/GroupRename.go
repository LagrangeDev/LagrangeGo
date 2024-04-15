package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
)

func BuildGroupRenameReq(groupUin uint32, name string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X89A_15{
		GroupUin: groupUin,
		Body:     &oidb.OidbSvcTrpcTcp0X89A_15Body{TargetName: name},
	}
	return BuildOidbPacket(0x89A, 15, body, false, false)
}

func ParseGroupRenameResp(data []byte) (bool, error) {
	baseResp, err := ParseOidbPacket(data, nil)
	if err != nil {
		return false, err
	}
	if baseResp.ErrorCode != 0 {
		return false, errors.New(baseResp.ErrorMsg)
	}
	return true, nil
}
