package oidb

import "github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"

func BuildFetchClientKeyReq() (*OidbPacket, error) {
	body := oidb.OidbSvcTrpcTcp0X102A_1{}
	return BuildOidbPacket(0x102A, 1, &body, false, false)
}

func ParseFetchClientKeyResp(data []byte) (string, error) {
	resp, err := ParseTypedError[oidb.OidbSvcTrpcTcp0X102A_1Response](data)
	if err != nil {
		return "", err
	}
	return resp.ClientKey, nil
}
