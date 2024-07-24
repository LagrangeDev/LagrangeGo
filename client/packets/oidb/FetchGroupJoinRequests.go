package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildFetchGroupSystemMessagesReq(count uint32) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X10C0_1{
		Count:  count,
		Field2: 0,
	}
	return BuildOidbPacket(0x10C0, 1, body, false, false)
}

func ParseFetchGroupSystemMessagesReq(data []byte, groupUin ...uint32) ([]*entity.GroupJoinRequest, error) {
	resp, err := ParseTypedError[oidb.OidbSvcTrpcTcp0X10C0_1Response](data)
	if err != nil {
		return nil, err
	}
	requests := make([]*entity.GroupJoinRequest, len(resp.Requests))
	for i, r := range resp.Requests {
		if len(groupUin) > 0 && groupUin[0] != r.Group.GroupUin {
			continue
		}
		requests[i] = &entity.GroupJoinRequest{
			GroupUin:  r.Group.GroupUin,
			TargetUid: r.Target.Uid,
			Sequence:  r.Sequence,
			State:     entity.EventState(r.State),
			EventType: r.EventType,
			Comment:   r.Comment,
		}
		if r.Invitor != nil {
			requests[i].InvitorUid = r.Invitor.Uid
		}
		if r.Operator != nil {
			requests[i].OperatorUid = r.Operator.Uid
		}
	}
	return requests, nil
}
