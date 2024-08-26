package oidb

import (
	"encoding/hex"
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupFSDownloadReq(groupUin uint32, fileId string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X6D6{
		Download: &oidb.OidbSvcTrpcTcp0X6D6Download{
			GroupUin: groupUin,
			AppId:    7,
			BusId:    102,
			FileId:   fileId,
		},
	}
	return BuildOidbPacket(0x6D6, 2, body, false, true)
}

func ParseGroupFSDownloadResp(data []byte) (string, error) {
	var resp oidb.OidbSvcTrpcTcp0X6D6Response
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return "", err
	}
	hexUrl := hex.EncodeToString(resp.Download.DownloadUrl)
	url := fmt.Sprintf("https://%s:443/ftn_handler/%s/?fname=", resp.Download.DownloadIp, hexUrl)
	return url, nil
}
