package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
)

func BuildGroupLeaveReq(groupUin uint32) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X1097_1{GroupUin: groupUin}
	return BuildOidbPacket(0x1097, 1, body, false, false)
}

func ParseGroupLeaveResp(data []byte) (bool, error) {
	baseResp, err := ParseOidbPacket(data, nil)
	if err != nil {
		return false, err
	}
	if baseResp.ErrorCode != 0 {
		return false, errors.New(baseResp.ErrorMsg)
	}
	return true, nil
}
