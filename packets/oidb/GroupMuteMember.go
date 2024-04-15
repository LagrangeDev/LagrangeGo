package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
)

func BuildGroupMuteMemberReq(groupUin, duration uint32, uid string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X1253_1{
		GroupUin: groupUin,
		Type:     1,
		Body: &oidb.OidbSvcTrpcTcp0X1253_1Body{
			TargetUid: uid,
			Duration:  duration,
		},
	}
	return BuildOidbPacket(0x1253, 1, body, false, false)
}

// ParseGroupMuteMemberResp 失败了会返回错误原因
func ParseGroupMuteMemberResp(data []byte) (bool, error) {
	var resp oidb.OidbSvcTrpcTcp0X1253_1Response
	baseResp, err := ParseOidbPacket(data, &resp)
	if err != nil {
		return false, err
	}
	if baseResp.ErrorCode != 0 {
		return false, errors.New(baseResp.ErrorMsg)
	}
	return true, nil
}
