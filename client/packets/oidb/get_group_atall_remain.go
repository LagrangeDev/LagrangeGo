package oidb

import "github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"

// ref https://github.com/Mrs4s/MiraiGo/blob/54bdd873e3fed9fe1c944918924674dacec5ac76/client/group_msg.go#L213

func BuildGetAtAllRemainRequest(uin, gin uint32) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0X8A7_0_ReqBody{
		SubCmd:                    1,
		LimitIntervalTypeForUin:   2,
		LimitIntervalTypeForGroup: 1,
		Uin:                       uint64(uin),
		GroupUin:                  uint64(gin),
	}
	return BuildOidbPacket(0x8A7, 0, body, false, true)
}

type AtAllRemainInfo struct {
	CanAtAll      bool
	CountForGroup uint32 // 当前群默认可用次数
	CountForUin   uint32 // 当前QQ剩余次数
}

func ParseGetAtAllRemainResponse(data []byte) (*AtAllRemainInfo, error) {
	var rsp oidb.OidbSvcTrpcTcp0X8A7_0_RspBody
	_, err := ParseOidbPacket(data, &rsp)
	if err != nil {
		return nil, err
	}
	return &AtAllRemainInfo{
		CanAtAll:      rsp.CanAtAll,
		CountForGroup: rsp.CountForGroup,
		CountForUin:   rsp.CountForUin,
	}, nil
}
