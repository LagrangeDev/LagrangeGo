package oidb

import "github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"

func BuildGroupFileCountReq(groupUin uint32) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X6D8{
		Count: &oidb.OidbSvcTrpcTcp0X6D8Count{
			GroupUin: groupUin,
			AppId:    7,
			BusId:    6,
		},
	}
	return BuildOidbPacket(0x6D8, 2, body, false, true)
}

func ParseGroupFileCountResp(data []byte) (fileCount uint32, limitCount uint32, error error) {
	var resp oidb.OidbSvcTrpcTcp0X6D8_1Response
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return 0, 0, err
	}
	return resp.Count.FileCount, resp.Count.LimitCount, nil
}
