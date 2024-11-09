package oidb

import (
	"math"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildGroupMuteGlobalReq(groupUin uint32, isMute bool) (*Packet, error) {
	var s uint32
	if isMute {
		s = math.MaxUint32
	}
	body := &oidb.OidbSvcTrpcTcp0X89A_0{
		GroupUin: groupUin,
		State:    &oidb.OidbSvcTrpcTcp0X89A_0State{S: s},
	}
	return BuildOidbPacket(0x89A, 0, body, false, false)
}

func ParseGroupMuteGlobalResp(data []byte) error {
	return CheckError(data)
}
