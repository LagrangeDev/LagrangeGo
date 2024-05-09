package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/pkg/pb/service/oidb"
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

func ParseGroupKickMemberResp(data []byte) error {
	return CheckTypedError[oidb.OidbSvcTrpcTcp0X8A0_1Response](data)
}
