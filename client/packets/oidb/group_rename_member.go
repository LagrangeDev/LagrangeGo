package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupRenameMemberReq(groupUin uint32, uid, name string) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0X8FC{
		GroupUin: groupUin,
		Body: &oidb.OidbSvcTrpcTcp0X8FCBody{
			TargetUid:  uid,
			TargetName: name,
		},
	}
	return BuildOidbPacket(0x8FC, 3, body, false, false)
}

func ParseGroupRenameMemberResp(data []byte) error {
	return CheckError(data)
}
