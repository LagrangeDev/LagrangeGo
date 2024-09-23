package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

func BuildGroupSetReactionReq(groupUin, sequence uint32, code string, isAdd bool) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X9082{
		GroupUin: groupUin,
		Sequence: sequence,
		Code:     proto.Some(code),
		Field5:   true,
		Field6:   false,
		Field7:   false,
	}
	return BuildOidbPacket(0x9082, utils.Ternary[uint32](isAdd, 1, 2), body, false, true)
}

func ParseGroupSetReactionResp(data []byte) error {
	return CheckError(data)
}
