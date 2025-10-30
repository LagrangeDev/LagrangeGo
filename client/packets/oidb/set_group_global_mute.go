package oidb

import (
	"math"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
)

func BuildSetGroupGlobalMuteReq(groupUin uint32, isMute bool) (*Packet, error) {
	var s uint32
	if isMute {
		s = math.MaxUint32
	}
	body := &oidb.OidbSvcTrpcTcp0X89A_0{
		GroupUin: groupUin,
		State:    &oidb.OidbSvcTrpcTcp0X89A_0State{S: proto.Uint32(s)},
	}
	return BuildOidbPacket(0x89A, 0, body, false, false)
}

func ParseSetGroupGlobalMuteResp(data []byte) error {
	return CheckError(data)
}
