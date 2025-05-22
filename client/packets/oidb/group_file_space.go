package oidb

import "github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"

func BuildGroupFileSpaceReq(groupUin uint32) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0X6D8{
		Space: &oidb.OidbSvcTrpcTcp0X6D8Space{
			GroupUin: groupUin,
			AppId:    7,
		},
	}
	return BuildOidbPacket(0x6D8, 3, body, false, true)
}

func ParseGroupFileSpaceResp(data []byte) (uint64, uint64, error) {
	var resp oidb.OidbSvcTrpcTcp0X6D8_1Response
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return 0, 0, err
	}
	return resp.Space.TotalSpace, resp.Space.UsedSpace, nil
}
