package oidb

import (
	"errors"
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/message"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
)

func BuildGroupRecordUploadReq(groupUin uint32, record *message.VoiceElement) (*OidbPacket, error) {
	if record.Stream == nil {
		return nil, errors.New("audio data is nil")
	}

	body := &oidb.NTV2RichMediaReq{
		ReqHead: &oidb.MultiMediaReqHead{
			Common: &oidb.CommonHead{
				RequestId: 1,
				Command:   100,
			},
			Scene: &oidb.SceneInfo{
				RequestType:  2,
				BusinessType: 3,
				SceneType:    2,
				Group: &oidb.NTGroupInfo{
					GroupUin: groupUin,
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
						FileSize: record.Size,
						FileHash: fmt.Sprintf("%x", record.Md5),
						FileSha1: fmt.Sprintf("%x", record.Sha1),
						Type: &oidb.FileType{
							Type:        3,
							PicFormat:   0,
							VideoFormat: 0,
							VoiceFormat: 1,
						},
						Width:    0,
						Height:   0,
						Time:     record.Duration,
						Original: 0,
					},
					SubFileType: 0,
				},
			},
			TryFastUploadCompleted: true,
			SrvSendMsg:             false,
			ClientRandomId:         uint64(crypto.RandU32()),
			CompatQMsgSceneType:    2,
			ExtBizInfo: &oidb.ExtBizInfo{
				Pic: &oidb.PicExtBizInfo{
					TextSummary: record.Summary,
				},
				Video: &oidb.VideoExtBizInfo{
					BytesPbReserve: make([]byte, 0),
				},
				Ptt: &oidb.PttExtBizInfo{
					BytesReserve:      make([]byte, 0),
					BytesPbReserve:    []byte{0x08, 0x00, 0x38, 0x00},
					BytesGeneralFlags: []byte{0x9a, 0x01, 0x07, 0xaa, 0x03, 0x04, 0x08, 0x08, 0x12, 0x00},
				},
			},
			ClientSeq:       0,
			NoNeedCompatMsg: false,
		},
	}

	return BuildOidbPacket(0x126E, 100, body, false, true)
}

func ParseGroupRecordUploadResp(data []byte) (*oidb.NTV2RichMediaResp, error) {
	return ParseTypedError[oidb.NTV2RichMediaResp](data)
}
