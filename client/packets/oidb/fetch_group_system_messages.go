package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

func BuildFetchGroupSystemMessagesReq(isFiltered bool, count uint32) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X10C0{
		Count:  count,
		Field2: 0,
	}
	return BuildOidbPacket(0x10C0, utils.Ternary[uint32](isFiltered, 2, 1), body, false, false)
}

func ParseFetchGroupSystemMessagesReq(isFiltered bool, data []byte, groupUin ...uint32) ([]*entity.GroupJoinRequest, error) {
	resp, err := ParseTypedError[oidb.OidbSvcTrpcTcp0X10C0Response](data)
	if err != nil {
		return nil, err
	}
	requests := make([]*entity.GroupJoinRequest, 0)
	for _, r := range resp.Requests {
		if len(groupUin) > 0 && groupUin[0] != r.Group.GroupUin {
			continue
		}
		req := &entity.GroupJoinRequest{
			GroupUin:   r.Group.GroupUin,
			TargetUid:  r.Target.Uid,
			Sequence:   r.Sequence,
			State:      entity.EventState(r.State),
			EventType:  r.EventType,
			Comment:    r.Comment,
			IsFiltered: isFiltered,
		}
		if r.Invitor != nil {
			req.InvitorUid = r.Invitor.Uid
		}
		if r.Operator != nil {
			req.OperatorUid = r.Operator.Uid
		}
		requests = append(requests, req)
	}
	return requests, nil
}
