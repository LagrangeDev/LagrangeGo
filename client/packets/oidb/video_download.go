package oidb

import (
	"encoding/hex"
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

func BuildVideoDownloadReq(selfUid, videoUuid, videoName string, isGroup bool, md5, sha1 []byte) (*OidbPacket, error) {
	body := &oidb.NTV2RichMediaReq{
		ReqHead: &oidb.MultiMediaReqHead{
			Common: &oidb.CommonHead{
				RequestId: utils.Ternary[uint32](isGroup, 3, 34), // private 12
				Command:   200,
			},
			Scene: &oidb.SceneInfo{
				RequestType:  2,
				BusinessType: 2,
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
			Node: &oidb.IndexNode{
				Info: &oidb.FileInfo{
					FileSize: 0,
					FileHash: hex.EncodeToString(md5),
					FileSha1: hex.EncodeToString(sha1),
					FileName: videoName,
					Type: &oidb.FileType{
						Type:        2,
						PicFormat:   0,
						VideoFormat: 0,
						VoiceFormat: 0,
					},
					Width:    0,
					Height:   0,
					Time:     0,
					Original: 0,
				},
				FileUuid:   videoUuid,
				StoreId:    0,
				UploadTime: 0,
				Ttl:        0,
				SubType:    0,
			},
			Download: &oidb.DownloadExt{
				Video: &oidb.VideoDownloadExt{
					BusiType:  0,
					SceneType: 0,
				},
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
