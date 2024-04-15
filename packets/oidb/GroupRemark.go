package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
)

func BuildGroupRemark(groupUin uint32, mark string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0XF16_1{
		Body: &oidb.OidbSvcTrpcTcp0XF16_1Body{
			GroupUin:     groupUin,
			TargetRemark: mark,
		},
	}
	return BuildOidbPacket(0xF16, 1, body, false, false)
}

func ParseGroupRemarkResp(data []byte) (bool, error) {
	baseResp, err := ParseOidbPacket(data, nil)
	if err != nil {
		return false, err
	}
	if baseResp.ErrorCode != 0 {
		return false, errors.New(baseResp.ErrorMsg)
	}
	return true, nil
}
