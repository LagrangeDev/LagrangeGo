package event

import (
	"github.com/LagrangeDev/LagrangeGo/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/status"
)

type (
	GroupEvent struct {
		GroupID uint32
	}

	GroupMuteMember struct {
		GroupEvent
		OperatorUid string
		TargetUid   string // when TargetUid is empty, mute all members
		Duration    uint32
	}

	GroupMemberJoinRequest struct {
		GroupEvent
		Uid        string
		InvitorUid string
		Answer     string // 问题：(.*)答案：(.*)
	}

	GroupMemberJoined struct {
		GroupEvent
		Uin      uint32
		Uid      string
		JoinType int32
	}

	GroupMemberQuit struct {
		GroupEvent
		Uin         uint32
		Uid         string
		ExitType    int32
		OperatorUid string
	}
)

func (gmq *GroupMemberQuit) IsKicked() bool {
	return gmq.ExitType == 131
}

func ParseMemberJoinRequest(event *status.MemberJoinRequest) *GroupMemberJoinRequest {
	return &GroupMemberJoinRequest{
		GroupEvent: GroupEvent{
			GroupID: event.GroupID,
		},
		Uid:    event.Uid,
		Answer: event.RequestField.Unwrap(),
	}
}
func ParseMemberJoinRequestFromInvite(event *status.MemberInviteRequest) *GroupMemberJoinRequest {
	if event.Cmd == 87 {
		inn := event.Info.Inner
		return &GroupMemberJoinRequest{
			GroupEvent: GroupEvent{
				GroupID: inn.GroupID,
			},
			Uid:        inn.Uid,
			InvitorUid: inn.InvitorUid,
		}
	}
	return nil
}

func ParseMemberJoined(msg *message.PushMsg, event *status.MemberChanged) *GroupMemberJoined {
	return &GroupMemberJoined{
		GroupEvent: GroupEvent{
			GroupID: msg.Message.ResponseHead.FromUin,
		},
		Uin:      event.Uin,
		Uid:      event.Uid,
		JoinType: event.JoinType.Unwrap(),
	}
}

func ParseMemberQuit(msg *message.PushMsg, event *status.MemberChanged) *GroupMemberQuit {
	return &GroupMemberQuit{
		GroupEvent: GroupEvent{
			GroupID: msg.Message.ResponseHead.FromUin,
		},
		Uin:         event.Uin,
		Uid:         event.Uid,
		OperatorUid: event.OperatorUid.Unwrap(),
		ExitType:    event.ExitType.Unwrap(),
	}
}

func ParseGroupMuteEvent(event *map[int]any) *GroupMuteMember {
	// TODO Parse GroupMuteEvent
	return nil
}
