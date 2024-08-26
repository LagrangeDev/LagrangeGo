package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
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

func ParseGroupSetSpecialTitleResp(data []byte) error {
	return CheckError(data)
}
