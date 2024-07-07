package oidb

import (
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

// BuildPrivateRecordDownloadReq 私聊语音
func BuildPrivateRecordDownloadReq(selfUid string, node *oidb.IndexNode) (*OidbPacket, error) {
	body := oidb.NTV2RichMediaReq{
		ReqHead: &oidb.MultiMediaReqHead{
			Common: &oidb.CommonHead{
				RequestId: 1,
				Command:   200,
			},
			Scene: &oidb.SceneInfo{
				RequestType:  1,
				BusinessType: 3,
				SceneType:    1,
				C2C: &oidb.C2CUserInfo{
					AccountType: 2,
					TargetUid:   selfUid,
				},
			},
			Client: &oidb.ClientMeta{
				AgentType: 2,
			},
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

	return BuildOidbPacket(0x126D, 200, &body, false, true)
}

func ParsePrivateRecordDownloadResp(data []byte) (string, error) {
	var resp oidb.NTV2RichMediaResp
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return "", err
	}
	body := resp.Download
	return fmt.Sprintf("https://%s%s%s", body.Info.Domain, body.Info.UrlPath, body.RKeyParam), nil
}
