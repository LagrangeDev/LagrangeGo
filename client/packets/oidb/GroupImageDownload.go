package oidb

import (
	"fmt"

	oidb2 "github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupImageDownloadReq(groupUin uint32, node *oidb2.IndexNode) (*OidbPacket, error) {
	body := &oidb2.NTV2RichMediaReq{
		ReqHead: &oidb2.MultiMediaReqHead{
			Common: &oidb2.CommonHead{
				RequestId: 1,
				Command:   200,
			},
			Scene: &oidb2.SceneInfo{
				RequestType:  2,
				BusinessType: 1,
				SceneType:    2,
				Group:        &oidb2.NTGroupInfo{GroupUin: groupUin},
			},
			Client: &oidb2.ClientMeta{AgentType: 2},
		},
		Download: &oidb2.DownloadReq{
			Node: node,
			Download: &oidb2.DownloadExt{
				Video: &oidb2.VideoDownloadExt{
					BusiType:  0,
					SceneType: 0,
				},
			},
		},
	}
	return BuildOidbPacket(0x11C4, 200, body, false, true)
}

func ParseGroupImageDownloadResp(data []byte) (string, error) {
	resp, err := ParseTypedError[oidb2.NTV2RichMediaResp](data)
	if err != nil {
		return "", err
	}
	body := resp.Download
	return fmt.Sprintf("https://%s%s%s", body.Info.Domain, body.Info.UrlPath, body.RKeyParam), nil
}
