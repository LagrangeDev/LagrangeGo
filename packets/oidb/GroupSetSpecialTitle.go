package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
)

func BuildGroupSetSpecialTitleReq(groupUin uint32, uid, title string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X8FC{
		GroupUin: groupUin,
		Body: &oidb.OidbSvcTrpcTcp0X8FCBody{
			TargetUid:              uid,
			SpecialTitle:           title,
			SpecialTitleExpireTime: -1,
			UinName:                title,
		},
	}
	return BuildOidbPacket(0x8FC, 2, body, false, false)
}

func ParseGroupSetSpecialTitleResp(data []byte) (bool, error) {
	baseResp, err := ParseOidbPacket(data, nil)
	if err != nil {
		return false, err
	}
	if baseResp.ErrorCode != 0 {
		return false, errors.New(baseResp.ErrorMsg)
	}
	return true, nil
}
