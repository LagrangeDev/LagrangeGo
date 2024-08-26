package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupFileListReq(groupUin uint32, targetDirectory string, startIndex uint32, fileCount uint32) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X6D8{
		List: &oidb.OidbSvcTrpcTcp0X6D8List{
			GroupUin:        groupUin,
			AppId:           7,
			TargetDirectory: targetDirectory,
			FileCount:       fileCount,
			SortBy:          1,
			StartIndex:      startIndex,
			Field17:         2,
			Field18:         0,
		},
	}
	return BuildOidbPacket(0x6D8, 1, body, false, true)
}

func ParseGroupFileListResp(data []byte) (*oidb.OidbSvcTrpcTcp0X6D8_1Response, error) {
	var resp oidb.OidbSvcTrpcTcp0X6D8_1Response
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return &resp, err
	}
	if resp.List.RetCode != 0 {
		return &resp, errors.New(resp.List.ClientWording)
	}
	return &resp, nil
}
