package oidb

import (
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupImageDownloadReq(groupUin uint32, node *oidb.IndexNode) (*OidbPacket, error) {
	body := &oidb.NTV2RichMediaReq{
		ReqHead: &oidb.MultiMediaReqHead{
			Common: &oidb.CommonHead{
				RequestId: 1,
				Command:   200,
			},
			Scene: &oidb.SceneInfo{
				RequestType:  2,
				BusinessType: 1,
				SceneType:    2,
				Group:        &oidb.NTGroupInfo{GroupUin: groupUin},
			},
			Client: &oidb.ClientMeta{AgentType: 2},
		},
		Download: &oidb.DownloadReq{
			Node: node,
			Download: &oidb.DownloadExt{
				Video: &oidb.VideoDownloadExt{
					BusiType:  0,
					SceneType: 0,
				},
			},
		},
	}
	return BuildOidbPacket(0x11C4, 200, body, false, true)
}

func ParseGroupImageDownloadResp(data []byte) (string, error) {
	resp, err := ParseTypedError[oidb.NTV2RichMediaResp](data)
	if err != nil {
		return "", err
	}
	body := resp.Download
	return fmt.Sprintf("https://%s%s%s", body.Info.Domain, body.Info.UrlPath, body.RKeyParam), nil
}
