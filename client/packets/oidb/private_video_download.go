package oidb

import (
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildPrivateVideoDownloadReq(selfUID string, node *oidb.IndexNode) (*Packet, error) {
	body := &oidb.NTV2RichMediaReq{
		ReqHead: &oidb.MultiMediaReqHead{
			Common: &oidb.CommonHead{
				RequestId: 34, // private 12
				Command:   200,
			},
			Scene: &oidb.SceneInfo{
				RequestType:  2,
				BusinessType: 2,
				SceneType:    1,
				C2C: &oidb.C2CUserInfo{
					AccountType: 2,
					TargetUid:   selfUID,
				},
			},
			Client: &oidb.ClientMeta{
				AgentType: 2,
			},
		},
		Download: &oidb.DownloadReq{
			Node: node,
			Download: &oidb.DownloadExt{
				Video: &oidb.VideoDownloadExt{},
			},
		},
	}
	return BuildOidbPacket(0x11E9, 200, body, false, true)
}

func ParseVideoDownloadResp(data []byte) (string, error) {
	resp, err := ParseTypedError[oidb.NTV2RichMediaResp](data)
	if err != nil {
		return "", err
	}
	body := resp.Download
	return fmt.Sprintf("https://%s%s%s", body.Info.Domain, body.Info.UrlPath, body.RKeyParam), nil
}
