package oidb

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupFSDownloadReq(groupUin uint32, fileID string) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0X6D6{
		Download: &oidb.OidbSvcTrpcTcp0X6D6Download{
			GroupUin: groupUin,
			AppId:    7,
			BusId:    102,
			FileId:   fileID,
		},
	}
	return BuildOidbPacket(0x6D6, 2, body, false, true)
}

func ParseGroupFSDownloadResp(data []byte) (string, error) {
	var resp oidb.OidbSvcTrpcTcp0X6D6Response
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return "", err
	}
	if resp.Download.RetCode != 0 {
		return "", errors.New(resp.Download.ClientWording)
	}
	hexURL := hex.EncodeToString(resp.Download.DownloadUrl)
	url := fmt.Sprintf("https://%s:443/ftn_handler/%s/?fname=", resp.Download.DownloadIp, hexURL)
	return url, nil
}
