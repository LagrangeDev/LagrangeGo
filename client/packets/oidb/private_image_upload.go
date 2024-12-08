package oidb

import (
	"encoding/hex"
	"errors"
	"math/rand"

	"github.com/LagrangeDev/LagrangeGo/message"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

func BuildPrivateImageUploadReq(targetUID string, image *message.ImageElement) (*Packet, error) {
	// OidbSvcTrpcTcp.0x11c5_100
	if image.Stream == nil {
		return nil, errors.New("image data is null")
	}

	format, size, err := utils.ImageResolve(image.Stream)
	if err != nil {
		return nil, err
	}
	imageExt := format.String()

	body := &oidb.NTV2RichMediaReq{
		ReqHead: &oidb.MultiMediaReqHead{
			Common: &oidb.CommonHead{
				RequestId: 1,
				Command:   100,
			},
			Scene: &oidb.SceneInfo{
				RequestType:  2,
				BusinessType: 1,
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
						FileSize: image.Size,
						FileHash: hex.EncodeToString(image.Md5),
						FileSha1: hex.EncodeToString(image.Sha1),
						FileName: hex.EncodeToString(image.Md5) + "." + imageExt,
						Type: &oidb.FileType{
							Type:        1,
							PicFormat:   uint32(format),
							VideoFormat: 0,
							VoiceFormat: 0,
						},
						Width:    uint32(size.Width),
						Height:   uint32(size.Height),
						Time:     0,
						Original: 1,
					},
					SubFileType: 0,
				},
			},
			TryFastUploadCompleted: true,
			SrvSendMsg:             false,
			ClientRandomId:         uint64(rand.Int63()),
			CompatQMsgSceneType:    1,
			ExtBizInfo: &oidb.ExtBizInfo{
				Pic: &oidb.PicExtBizInfo{
					BizType:     uint32(image.SubType),
					TextSummary: image.Summary,
					ExtData: &oidb.PicExtData{
						SubType:     uint32(image.SubType),
						TextSummary: image.Summary,
					},
				},
				Video: &oidb.VideoExtBizInfo{
					BytesPbReserve: []byte{},
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
	return BuildOidbPacket(0x11c5, 100, body, false, true)
}

func ParsePrivateImageUploadResp(data []byte) (*oidb.NTV2RichMediaResp, error) {
	return ParseTypedError[oidb.NTV2RichMediaResp](data)
}
