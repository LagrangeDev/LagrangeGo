package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"

	"github.com/RomiChan/protobuf/proto"
)

// see https://github.com/Mrs4s/MiraiGo/blob/master/client/security.go

type UrlSecurityLevel int

const (
	UrlSecurityLevelSafe UrlSecurityLevel = iota + 1
	UrlSecurityLevelUnknown
	UrlSecurityLevelDanger
)

func (m UrlSecurityLevel) String() string {
	switch m {
	case UrlSecurityLevelSafe:
		return "safe"
	case UrlSecurityLevelDanger:
		return "danger"
	default:
		return "unknown"
	}
}

func BuildUrlCheckRequest(botuin uint32, url string) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0XBCB_0_ReqBody{
		CheckUrlReq: &oidb.CheckUrlReq{
			Url:         []string{url},
			QqPfTo:      proto.String("mqq.group"),
			Type:        proto.Uint32(2),
			SendUin:     proto.Uint64(uint64(botuin)),
			ReqType:     proto.String("webview"),
			OriginalUrl: proto.Some(url),
			IsArk:       proto.Bool(false),
			IsFinish:    proto.Bool(false),
			SrcUrls:     []string{url},
			SrcPlatform: proto.Uint32(1),
			Qua:         proto.String("AQQ_2013 4.6/2013 8.4.184945&NA_0/000000&ADR&null18&linux&2017&C2293D02BEE31158&7.1.2&V3"),
		},
	}
	return BuildOidbPacket(0xBCB, 0, body, false, false)
}

func ParseUrlCheckResponse(data []byte) (UrlSecurityLevel, error) {
	var rsp oidb.OidbSvcTrpcTcp0XBCB_0_RspBody
	_, err := ParseOidbPacket(data, &rsp)
	if err != nil {
		return UrlSecurityLevelUnknown, err
	}
	if rsp.CheckUrlRsp == nil || len(rsp.CheckUrlRsp.Results) == 0 {
		return UrlSecurityLevelUnknown, errors.New("response is empty")
	}
	if rsp.CheckUrlRsp.Results[0].JumpUrl.IsSome() {
		return UrlSecurityLevelDanger, nil
	}
	if rsp.CheckUrlRsp.Results[0].Umrtype.Unwrap() == 2 {
		return UrlSecurityLevelSafe, nil
	}
	return UrlSecurityLevelUnknown, nil
}