package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildKickGroupMemberReq(groupUin uint32, uid string, rejectAddRequest bool) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0X8A0_1{
		GroupUin:         groupUin,
		TargetUid:        uid,
		RejectAddRequest: rejectAddRequest,
		Field5:           "",
	}
	return BuildOidbPacket(0x8A0, 1, body, false, false)
}

func ParseKickGroupMemberResp(data []byte) error {
	return CheckTypedError[oidb.OidbSvcTrpcTcp0X8A0_1Response](data)
}
