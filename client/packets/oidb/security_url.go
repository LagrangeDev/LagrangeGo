package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

// see https://github.com/Mrs4s/MiraiGo/blob/master/client/security.go

type URLSecurityLevel int

const (
	URLSecurityLevelSafe URLSecurityLevel = iota + 1
	URLSecurityLevelUnknown
	URLSecurityLevelDanger
)

func (m URLSecurityLevel) String() string {
	switch m {
	case URLSecurityLevelSafe:
		return "safe"
	case URLSecurityLevelDanger:
		return "danger"
	case URLSecurityLevelUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

func BuildURLCheckRequest(botuin uint32, url string) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0XBCB_0_ReqBody{
		CheckUrlReq: &oidb.CheckUrlReq{
			Url:         []string{url},
			QqPfTo:      "mqq.group",
			Type:        2,
			SendUin:     uint64(botuin),
			ReqType:     "webview",
			OriginalUrl: url,
			IsArk:       false,
			IsFinish:    false,
			SrcUrls:     []string{url},
			SrcPlatform: 1,
			Qua:         "AQQ_2013 4.6/2013 8.4.184945&NA_0/000000&ADR&null18&linux&2017&C2293D02BEE31158&7.1.2&V3",
		},
	}
	return BuildOidbPacket(0xBCB, 0, body, false, false)
}

func ParseURLCheckResponse(data []byte) (URLSecurityLevel, error) {
	var rsp oidb.OidbSvcTrpcTcp0XBCB_0_RspBody
	_, err := ParseOidbPacket(data, &rsp)
	if err != nil {
		return URLSecurityLevelUnknown, err
	}
	if rsp.CheckUrlRsp == nil || len(rsp.CheckUrlRsp.Results) == 0 {
		return URLSecurityLevelUnknown, errors.New("response is empty")
	}
	if rsp.CheckUrlRsp.Results[0].JumpUrl != "" {
		return URLSecurityLevelDanger, nil
	}
	if rsp.CheckUrlRsp.Results[0].Umrtype == 2 {
		return URLSecurityLevelSafe, nil
	}
	return URLSecurityLevelUnknown, nil
}
