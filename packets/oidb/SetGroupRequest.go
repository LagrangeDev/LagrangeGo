package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
	"github.com/RomiChan/protobuf/proto"
)

func BuildSetGroupRequestReq(accept bool, sequence uint64, typ uint32, groupUin uint32, message string) (*OidbPacket, error) {
	var acceptInt uint32
	if accept {
		acceptInt = 1
	} else {
		acceptInt = 2
	}
	body := oidb.OidbSvcTrpcTcp0X10C8_1{
		Accept: acceptInt,
		Body: &oidb.OidbSvcTrpcTcp0X10C8_1Body{
			Sequence:  sequence,
			EventType: typ,
			GroupUin:  groupUin,
			Message:   proto.Some(message),
		},
	}
	return BuildOidbPacket(0x10C0, 1, body, false, false)
}

func ParseSetGroupRequestResp(data []byte) (bool, error) {
	baseResp, err := ParseOidbPacket(data, nil)
	if err != nil {
		return false, err
	}
	if baseResp.ErrorCode != 0 {
		return false, nil
	}
	return true, nil
}
