package event

import (
	"github.com/LagrangeDev/LagrangeGo/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

type (
	GroupEvent struct {
		GroupUin uint32
	}

	GroupMute struct {
		GroupEvent
		OperatorUid string
		TargetUid   string // when TargetUid is empty, mute all members
		Duration    uint32 // Duration == math.MaxUint32 when means mute all
	}

	GroupRecall struct {
		GroupEvent
		AuthorUid   string
		OperatorUid string
		Sequence    uint64
		Time        uint32
		Random      uint32
	}

	GroupMemberJoinRequest struct {
		GroupEvent
		TargetUid  string
		InvitorUid string
		Answer     string // 问题：(.*)答案：(.*)
	}

	GroupMemberIncrease struct {
		GroupEvent
		MemberUid  string
		InvitorUid string
		JoinType   uint32
	}

	GroupMemberDecrease struct {
		GroupEvent
		MemberUid   string
		OperatorUid string
		ExitType    uint32
	}
)

type GroupInvite struct {
	GroupUin   uint32
	InvitorUid string
}

func (gmd *GroupMemberDecrease) IsKicked() bool {
	return gmd.ExitType == 131 || gmd.ExitType == 3
}

// ParseRequestJoinNotice 主动加群
func ParseRequestJoinNotice(event *message.GroupJoin) *GroupMemberJoinRequest {
	return &GroupMemberJoinRequest{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
		},
		TargetUid: event.TargetUid,
		Answer:    event.RequestField.Unwrap(),
	}
}

// ParseRequestInvitationNotice 邀请加群
func ParseRequestInvitationNotice(event *message.GroupInvitation) *GroupMemberJoinRequest {
	if event.Cmd == 87 {
		inn := event.Info.Inner
		return &GroupMemberJoinRequest{
			GroupEvent: GroupEvent{
				GroupUin: inn.GroupUin,
			},
			TargetUid:  inn.TargetUid,
			InvitorUid: inn.InvitorUid,
		}
	}
	return nil
}

// ParseInviteNotice 被邀请加群
func ParseInviteNotice(event *message.GroupInvite) *GroupInvite {
	return &GroupInvite{
		GroupUin:   event.GroupUin,
		InvitorUid: event.InvitorUid,
	}
}

func ParseMemberIncreaseEvent(event *message.GroupChange) *GroupMemberIncrease {
	return &GroupMemberIncrease{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
		},
		MemberUid:  event.MemberUid,
		InvitorUid: utils.B2S(event.Operator),
		JoinType:   event.IncreaseType,
	}
}

func ParseMemberDecreaseEvent(event *message.GroupChange) *GroupMemberDecrease {
	return &GroupMemberDecrease{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
		},
		MemberUid:   event.MemberUid,
		OperatorUid: utils.B2S(event.Operator),
		ExitType:    event.DecreaseType,
	}
}

func ParseGroupRecallEvent(event *message.NotifyMessageBody) *GroupRecall {
	info := event.Recall.RecallMessages[0]
	result := GroupRecall{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
		},
		AuthorUid:   info.AuthorUid,
		OperatorUid: event.Recall.OperatorUid.Unwrap(),
		Sequence:    info.Sequence,
		Time:        info.Time,
		Random:      info.Random,
	}
	return &result
}

func ParseGroupMuteEvent(event *message.GroupMute) *GroupMute {
	return &GroupMute{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
		},
		OperatorUid: event.OperatorUid.Unwrap(),
		TargetUid:   event.Data.State.TargetUid.Unwrap(),
		Duration:    event.Data.State.Duration,
	}
}
