package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
)

func BuildGroupKickMemberReq(groupUin uint32, uid string, rejectAddRequest bool) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X8A0_1{
		GroupUin:         groupUin,
		TargetUid:        uid,
		RejectAddRequest: rejectAddRequest,
		Field5:           "",
	}
	return BuildOidbPacket(0x8A0, 1, body, false, false)
}

func ParseGroupKickMemberResp(data []byte) (bool, error) {
	var resp oidb.OidbSvcTrpcTcp0X8A0_1Response
	baseResp, err := ParseOidbPacket(data, &resp)
	if err != nil {
		return false, err
	}
	if baseResp.ErrorCode != 0 {
		return false, errors.New(baseResp.ErrorMsg)
	}
	return true, nil
}
