package oidb

import (
	"encoding/hex"
	"errors"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
)

func BuildPrivateRecordUploadReq(targetUid string, record *message.VoiceElement) (*OidbPacket, error) {
	if record.Stream == nil {
		return nil, errors.New("record data is nil")
	}
	md5 := hex.EncodeToString(record.Md5)
	sha1 := hex.EncodeToString(record.Sha1)
	body := &oidb.NTV2RichMediaReq{
		ReqHead: &oidb.MultiMediaReqHead{
			Common: &oidb.CommonHead{
				RequestId: 4,
				Command:   100,
			},
			Scene: &oidb.SceneInfo{
				RequestType:  2,
				BusinessType: 3,
				SceneType:    1,
				C2C: &oidb.C2CUserInfo{
					AccountType: 2,
					TargetUid:   targetUid,
				},
			},
			Client: &oidb.ClientMeta{
				AgentType: 2,
			},
		},
		Upload: &oidb.UploadReq{
			UploadInfo: []*oidb.UploadInfo{{
				FileInfo: &oidb.FileInfo{
					FileSize: record.Size,
					FileHash: md5,
					FileSha1: sha1,
					FileName: md5 + "amr",
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
			}},
			TryFastUploadCompleted: true,
			SrvSendMsg:             false,
			ClientRandomId:         uint64(crypto.RandU32()),
			CompatQMsgSceneType:    1,
			ExtBizInfo: &oidb.ExtBizInfo{
				Pic: &oidb.PicExtBizInfo{
					TextSummary: record.Summary,
				},
				Ptt: &oidb.PttExtBizInfo{
					BytesReserve:      []byte{0x08, 0x00, 0x38, 0x00},
					BytesGeneralFlags: []byte{0x9a, 0x01, 0x0b, 0xaa, 0x03, 0x08, 0x08, 0x04, 0x12, 0x04, 0x00, 0x00, 0x00, 0x00},
				},
			},
			ClientSeq:       0,
			NoNeedCompatMsg: false,
		},
	}
	return BuildOidbPacket(0x126D, 100, body, false, true)
}

func ParsePrivateRecordUploadResp(data []byte) (*oidb.NTV2RichMediaResp, error) {
	var resp oidb.NTV2RichMediaResp
	baseResp, err := ParseOidbPacket(data, &resp)
	if err != nil {
		return nil, err
	}
	if baseResp.ErrorCode != 0 {
		return nil, errors.New(baseResp.ErrorMsg)
	}

	return &resp, nil
}
