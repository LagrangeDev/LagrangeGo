package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildSetGroupRemarkReq(groupUin uint32, mark string) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0XF16_1{
		Body: &oidb.OidbSvcTrpcTcp0XF16_1Body{
			GroupUin:     groupUin,
			TargetRemark: mark,
		},
	}
	return BuildOidbPacket(0xF16, 1, body, false, false)
}

func ParseSetGroupRemarkResp(data []byte) error {
	return CheckError(data)
}
