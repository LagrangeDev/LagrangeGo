package oidb

import (
	"github.com/RomiChan/protobuf/proto"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

func BuildSetGroupRequestReq(isFiltered bool, accept bool, sequence uint64, typ uint32, groupUin uint32, message string) (*Packet, error) {
	body := oidb.OidbSvcTrpcTcp0X10C8{
		Accept: uint32(utils.Bool2Int(!accept) + 1),
		Body: &oidb.OidbSvcTrpcTcp0X10C8Body{
			Sequence:  sequence,
			EventType: typ,
			GroupUin:  groupUin,
			Message:   proto.Some(message),
		},
	}
	return BuildOidbPacket(0x10C8, utils.Ternary[uint32](isFiltered, 2, 1), &body, false, false)
}

func ParseSetGroupRequestResp(data []byte) error {
	return CheckError(data)
}
