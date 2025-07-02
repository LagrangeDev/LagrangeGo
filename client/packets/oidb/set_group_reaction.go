package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils/io"
)

func BuildSetGroupReactionReq(groupUin, sequence uint32, code string, isAdd bool) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0X9082{
		GroupUin: groupUin,
		Sequence: sequence,
		Code:     proto.Some(code),
		Type:     io.Ternary[uint32](len(code) > 3, 2, 1),
		Field6:   false,
		Field7:   false,
	}
	return BuildOidbPacket(0x9082, io.Ternary[uint32](isAdd, 1, 2), body, false, true)
}

func ParseSetGroupReactionResp(data []byte) error {
	return CheckError(data)
}
