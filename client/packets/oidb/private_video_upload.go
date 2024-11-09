package oidb

import (
	"encoding/hex"
	"errors"

	"github.com/LagrangeDev/LagrangeGo/utils/crypto"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"

	"github.com/LagrangeDev/LagrangeGo/message"
)

func BuildPrivateVideoUploadReq(targetUID string, video *message.ShortVideoElement) (*Packet, error) {
	if video.Stream == nil {
		return nil, errors.New("video data is nil")
	}
	body := &oidb.NTV2RichMediaReq{
		ReqHead: &oidb.MultiMediaReqHead{
			Common: &oidb.CommonHead{
				RequestId: 3,
				Command:   100,
			},
			Scene: &oidb.SceneInfo{
				RequestType:  2,
				BusinessType: 2,
				SceneType:    1,
				C2C: &oidb.C2CUserInfo{
					AccountType: 2,
					TargetUid:   targetUID,
				},
			},
			Client: &oidb.ClientMeta{
				AgentType: 2,
			},
		},
		Upload: &oidb.UploadReq{
			UploadInfo: []*oidb.UploadInfo{
				{
					FileInfo: &oidb.FileInfo{
						FileSize: video.Size,
						FileHash: hex.EncodeToString(video.Md5),
						FileSha1: hex.EncodeToString(video.Sha1),
						FileName: "video.mp4",
						Type: &oidb.FileType{
							Type:        2,
							PicFormat:   0,
							VideoFormat: 0,
							VoiceFormat: 0,
						},
						Height:   0,
						Width:    0,
						Time:     video.Duration,
						Original: 0,
					},
					SubFileType: 0,
				}, {
					FileInfo: &oidb.FileInfo{
						FileSize: video.Thumb.Size,
						FileHash: hex.EncodeToString(video.Thumb.Md5),
						FileSha1: hex.EncodeToString(video.Thumb.Sha1),
						FileName: "video.jpg",
						Type: &oidb.FileType{
							Type:        1,
							PicFormat:   0,
							VideoFormat: 0,
							VoiceFormat: 0,
						},
						Width:    video.Thumb.Width,
						Height:   video.Thumb.Height,
						Time:     0,
						Original: 0,
					},
					SubFileType: 100,
				},
			},
			TryFastUploadCompleted: true,
			SrvSendMsg:             false,
			ClientRandomId:         uint64(crypto.RandU32()),
			CompatQMsgSceneType:    2,
			ExtBizInfo: &oidb.ExtBizInfo{
				Pic: &oidb.PicExtBizInfo{
					BizType:     0,
					TextSummary: video.Summary,
				},
				Video: &oidb.VideoExtBizInfo{
					BytesPbReserve: []byte{0x80, 0x01, 0x00},
				},
				Ptt: &oidb.PttExtBizInfo{
					BytesReserve:      []byte{},
					BytesPbReserve:    []byte{},
					BytesGeneralFlags: []byte{},
				},
			},
			ClientSeq:       0,
			NoNeedCompatMsg: false,
		},
	}
	return BuildOidbPacket(0x11E9, 100, body, false, true)
}

func ParsePrivateVideoUploadResp(data []byte) (*oidb.NTV2RichMediaResp, error) {
	return ParseTypedError[oidb.NTV2RichMediaResp](data)
}
