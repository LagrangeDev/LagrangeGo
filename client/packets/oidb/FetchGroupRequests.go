package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildFetchGroupRequestsReq() (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X10C0_1{
		Count:  20,
		Field2: 0,
	}
	return BuildOidbPacket(0x10C0, 1, body, false, false)
}

func ParseFetchGroupRequestsReq(data []byte) (*oidb.OidbSvcTrpcTcp0X10C0_1Response, error) {
	return ParseTypedError[oidb.OidbSvcTrpcTcp0X10C0_1Response](data)
}
