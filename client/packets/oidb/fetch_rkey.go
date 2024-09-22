package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildFetchRKeyReq() (*OidbPacket, error) {
	body := &oidb.NTV2RichMediaReq{
		ReqHead: &oidb.MultiMediaReqHead{
			Common: &oidb.CommonHead{
				RequestId: 1,
				Command:   202,
			},
			Scene: &oidb.SceneInfo{
				RequestType:  2,
				BusinessType: 1,
				SceneType:    0,
			},
			Client: &oidb.ClientMeta{
				AgentType: 2,
			},
		},
		DownloadRKey: &oidb.DownloadRKeyReq{
			Types: []int32{10, 20, 2},
		},
	}
	return BuildOidbPacket(0x9067, 202, body, false, false)
}

func ParseFetchRKeyResp(data []byte) (entity.RKeyMap, error) {
	resp := &oidb.NTV2RichMediaResp{}
	_, err := ParseOidbPacket(data, resp)
	if err != nil {
		return nil, err
	}
	var rKeyInfo = entity.RKeyMap{}
	for _, rkey := range resp.DownloadRKey.RKeys {
		typ := entity.RKeyType(rkey.Type.Unwrap())
		rKeyInfo[typ] = &entity.RKeyInfo{
			RKey:       rkey.Rkey,
			RKeyType:   typ,
			CreateTime: uint64(rkey.RkeyCreateTime.Unwrap()),
			ExpireTime: uint64(rkey.RkeyCreateTime.Unwrap()) + rkey.RkeyTtlSec,
		}
	}
	return rKeyInfo, nil
}
