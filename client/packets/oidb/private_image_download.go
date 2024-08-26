package oidb

import (
	"fmt"

	oidb2 "github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildPrivateImageDownloadReq(selfUid string, node *oidb2.IndexNode) (*OidbPacket, error) {
	body := &oidb2.NTV2RichMediaReq{
		ReqHead: &oidb2.MultiMediaReqHead{
			Common: &oidb2.CommonHead{
				RequestId: 1,
				Command:   200,
			},
			Scene: &oidb2.SceneInfo{
				RequestType:  2,
				BusinessType: 1,
				SceneType:    1,
				C2C: &oidb2.C2CUserInfo{
					AccountType: 2,
					TargetUid:   selfUid,
				},
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
	return BuildOidbPacket(0x11C5, 200, body, false, true)
}

func ParsePrivateImageDownloadResp(data []byte) (string, error) {
	resp, err := ParseTypedError[oidb2.NTV2RichMediaResp](data)
	if err != nil {
		return "", err
	}
	body := resp.Download
	return fmt.Sprintf("https://%s%s%s", body.Info.Domain, body.Info.UrlPath, body.RKeyParam), nil
}
