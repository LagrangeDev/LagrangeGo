package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/RomiChan/protobuf/proto"
)

func BuildSetGroupRequestReq(accept bool, sequence uint64, typ uint32, groupUin uint32, message string) (*OidbPacket, error) {
	body := oidb.OidbSvcTrpcTcp0X10C8_1{
		Accept: uint32(utils.Bool2Int(!accept) + 1),
		Body: &oidb.OidbSvcTrpcTcp0X10C8_1Body{
			Sequence:  sequence,
			EventType: typ,
			GroupUin:  groupUin,
			Message:   proto.Some(message),
		},
	}
	return BuildOidbPacket(0x10C8, 1, body, false, false)
}

func ParseSetGroupRequestResp(data []byte) error {
	return CheckError(data)
}
