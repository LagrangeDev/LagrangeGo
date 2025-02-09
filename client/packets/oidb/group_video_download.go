package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupVideoDownloadReq(groupUin uint32, node *oidb.IndexNode) (*Packet, error) {
	body := &oidb.NTV2RichMediaReq{
		ReqHead: &oidb.MultiMediaReqHead{
			Common: &oidb.CommonHead{
				RequestId: 1,
				Command:   200,
			},
			Scene: &oidb.SceneInfo{
				RequestType:  2,
				BusinessType: 2,
				SceneType:    2,
				Group: &oidb.NTGroupInfo{
					GroupUin: groupUin,
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
