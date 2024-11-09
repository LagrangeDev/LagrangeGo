package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

func BuildSetEssenceMessageReq(groupUin, seq, random uint32, isSet bool) (*Packet, error) {
	body := oidb.OidbSvcTrpcTcp0XEAC{
		GroupUin: groupUin,
		Sequence: seq,
		Random:   random,
	}
	return BuildOidbPacket(0xEAC, utils.Ternary[uint32](isSet, 1, 2), &body, false, false)
}

func ParseSetEssenceMessageResp(data []byte) error {
	return CheckError(data)
}
