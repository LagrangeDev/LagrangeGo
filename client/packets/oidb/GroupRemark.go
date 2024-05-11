package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupRemarkReq(groupUin uint32, mark string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0XF16_1{
		Body: &oidb.OidbSvcTrpcTcp0XF16_1Body{
			GroupUin:     groupUin,
			TargetRemark: mark,
		},
	}
	return BuildOidbPacket(0xF16, 1, body, false, false)
}

func ParseGroupRemarkResp(data []byte) error {
	return CheckError(data)
}
