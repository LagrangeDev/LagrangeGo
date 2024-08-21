package oidb

import "github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"

func BuildGroupFileDeleteReq(groupUin uint32, fileID string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X6D6{
		Delete: &oidb.OidbSvcTrpcTcp0X6D6Delete{
			GroupUin: groupUin,
			BusId:    102,
			FileId:   fileID,
		},
	}
	return BuildOidbPacket(0x6D6, 3, body, false, true)
}

func ParseGroupFileDeleteResp(data []byte) error {
	return CheckError(data)
}
