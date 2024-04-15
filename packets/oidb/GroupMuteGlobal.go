package oidb

import (
	"errors"
	"math"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
)

func BuildGroupMuteGlobalReq(groupUin uint32, isMute bool) (*OidbPacket, error) {
	var s uint32 = 0
	if isMute {
		s = math.MaxUint32
	}
	body := &oidb.OidbSvcTrpcTcp0X89A_0{
		GroupUin: groupUin,
		State:    &oidb.OidbSvcTrpcTcp0X89A_0State{S: s},
	}
	return BuildOidbPacket(0x89A, 0, body, false, false)
}

func ParseGroupMuteGlobalResp(data []byte) (bool, error) {
	baseResp, err := ParseOidbPacket(data, nil)
	if err != nil {
		return false, err
	}
	if baseResp.ErrorCode != 0 {
		return false, errors.New(baseResp.ErrorMsg)
	}
	return true, nil
}
