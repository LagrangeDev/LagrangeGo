package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

func BuildFetchGroupSystemMessagesReq(isFiltered bool, count uint32) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0X10C0{
		Count:  count,
		Field2: 0,
	}
	return BuildOidbPacket(0x10C0, utils.Ternary[uint32](isFiltered, 2, 1), body, false, false)
}

func ParseFetchGroupSystemMessagesReq(isFiltered bool, data []byte, groupUin ...uint32) (*entity.GroupSystemMessages, error) {
	resp, err := ParseTypedError[oidb.OidbSvcTrpcTcp0X10C0Response](data)
	if err != nil {
		return nil, err
	}
	requests := entity.GroupSystemMessages{}
	for _, r := range resp.Requests {
		if len(groupUin) > 0 && groupUin[0] != r.Group.GroupUin {
			continue
		}
		//nolint
		switch entity.EventType(r.EventType) {
		case entity.UserJoinRequest, entity.UserInvited:
			requests.JoinRequests = append(requests.JoinRequests, &entity.UserJoinGroupRequest{
				GroupUin: r.Group.GroupUin,
				InvitorUID: utils.LazyTernary(r.Invitor != nil, func() string {
					return r.Invitor.Uid
				}, func() string {
					return ""
				}),
				TargetUID: r.Target.Uid,
				OperatorUID: utils.LazyTernary(r.Invitor != nil, func() string {
					return r.Invitor.Uid
				}, func() string {
					return ""
				}),
				Sequence:   r.Sequence,
				Checked:    entity.EventState(r.State) != entity.Unprocessed,
				State:      entity.EventState(r.State),
				EventType:  entity.EventType(r.EventType),
				Comment:    r.Comment,
				IsFiltered: isFiltered,
			})
		case entity.GroupInvited:
			requests.InvitedRequests = append(requests.InvitedRequests, &entity.GroupInvitedRequest{
				GroupUin: r.Group.GroupUin,
				InvitorUID: utils.LazyTernary(r.Invitor != nil, func() string {
					return r.Invitor.Uid
				}, func() string {
					return ""
				}),
				Sequence:   r.Sequence,
				Checked:    entity.EventState(r.State) != entity.Unprocessed,
				State:      entity.EventState(r.State),
				EventType:  entity.EventType(r.EventType),
				IsFiltered: isFiltered,
			})
		default:
		}
	}
	return &requests, nil
}
