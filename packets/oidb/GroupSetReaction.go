package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
)

func BuildGroupSetReactionReq(groupUin, sequence uint32, code string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X9082_1{
		GroupUin: groupUin,
		Sequence: sequence,
		Code:     proto.Some(code),
		Field5:   true,
		Field6:   false,
		Field7:   false,
	}
	return BuildOidbPacket(0x9082, 1, body, false, true)
}

func ParseGroupSetReactionResp(data []byte) (bool, error) {
	baseResp, err := ParseOidbPacket(data, nil)
	if err != nil {
		return false, err
	}
	if baseResp.ErrorCode != 0 {
		return false, errors.New(baseResp.ErrorMsg)
	}
	return true, nil
}
