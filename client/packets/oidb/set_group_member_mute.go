package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildSetGroupMemberMuteReq(groupUin, duration uint32, uid string) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0X1253_1{
		GroupUin: groupUin,
		Type:     1,
		Body: &oidb.OidbSvcTrpcTcp0X1253_1Body{
			TargetUid: uid,
			Duration:  duration,
		},
	}
	return BuildOidbPacket(0x1253, 1, body, false, false)
}

// ParseSetGroupMemberMuteResp 失败了会返回错误原因
func ParseSetGroupMemberMuteResp(data []byte) error {
	return CheckTypedError[oidb.OidbSvcTrpcTcp0X1253_1Response](data)
}
