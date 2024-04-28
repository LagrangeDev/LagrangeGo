package oidb

import (
	"errors"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
)

func BuildFetchGroupRequestsReq() (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X10C0_1{
		Count:  20,
		Field2: 0,
	}
	return BuildOidbPacket(0x10C0, 1, body, false, false)
}

func ParseFetchGroupRequestsReq(data []byte) (*oidb.OidbSvcTrpcTcp0X10C0_1Response, error) {
	var resp oidb.OidbSvcTrpcTcp0X10C0_1Response
	baseResp, err := ParseOidbPacket(data, &resp)
	if err != nil {
		return nil, err
	}
	if baseResp.ErrorCode != 0 {
		return nil, errors.New(baseResp.ErrorMsg)
	}
	return &resp, nil
}
