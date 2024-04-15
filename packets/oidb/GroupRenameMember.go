package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
)

func BuildGroupRenameMemberReq(groupUin uint32, uid, name string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X8FC{
		GroupUin: groupUin,
		Body: &oidb.OidbSvcTrpcTcp0X8FCBody{
			TargetUid:  uid,
			TargetName: name,
		},
	}
	return BuildOidbPacket(0x8FC, 3, body, false, false)
}

func ParseGroupRenameMemberResp(data []byte) (bool, error) {
	baseResp, err := ParseOidbPacket(data, nil)
	if err != nil {
		return false, err
	}
	if baseResp.ErrorCode != 0 {
		return false, errors.New(baseResp.ErrorMsg)
	}
	return true, nil
}
