package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
)

func BuildGroupSetAdminReq(groupUin uint32, uid string, isAdmin bool) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X1096_1{
		GroupUin: groupUin,
		Uid:      uid,
		IsAdmin:  isAdmin,
	}
	return BuildOidbPacket(0x1096, 1, body, false, false)
}

func ParseGroupSetAdminResp(data []byte) (bool, error) {
	baseResp, err := ParseOidbPacket(data, nil)
	if err != nil {
		return false, err
	}
	if baseResp.ErrorCode != 0 {
		return false, errors.New(baseResp.ErrorMsg)
	}
	return true, nil
}
