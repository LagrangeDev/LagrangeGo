package event

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

type (
	GroupEvent struct {
		GroupUin uint32
	}

	GroupMute struct {
		GroupEvent
		OperatorUid string
		OperatorUin uint32
		TargetUid   string // when TargetUid is empty, mute all members
		TargetUin   uint32
		Duration    uint32 // Duration == math.MaxUint32 when means mute all
	}

	GroupRecall struct {
		GroupEvent
		AuthorUid   string
		AuthorUin   uint32
		OperatorUid string
		OperatorUin uint32
		Sequence    uint64
		Time        uint32
		Random      uint32
	}

	GroupMemberJoinRequest struct {
		GroupEvent
		TargetUid  string
		TargetUin  uint32
		InvitorUid string
		InvitorUin uint32
		Answer     string // 问题：(.*)答案：(.*)
	}

	GroupMemberIncrease struct {
		GroupEvent
		MemberUid  string
		MemberUin  uint32
		InvitorUid string
		InvitorUin uint32
		JoinType   uint32
	}

	GroupMemberDecrease struct {
		GroupEvent
		MemberUid   string
		MemberUin   uint32
		OperatorUid string
		OperatorUin uint32
		ExitType    uint32
	}
)

type GroupInvite struct {
	GroupUin   uint32
	InvitorUid string
	InvitorUin uint32
}

// CanPreprocess 实现预处理接口，对事件的uid进行转换等操作
type CanPreprocess interface {
	Preprocess(func(uid string) uint32)
}

func (g *GroupMute) MuteAll() bool {
	return g.OperatorUid == ""
}

func (g *GroupMemberDecrease) IsKicked() bool {
	return g.ExitType == 131 || g.ExitType == 3
}

func (g *GroupMemberJoinRequest) Preprocess(f func(uid string) uint32) {
	g.InvitorUin = f(g.InvitorUid)
	g.TargetUin = f(g.TargetUid)
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

func (g *GroupInvite) Preprocess(f func(uid string) uint32) {
	g.InvitorUin = f(g.InvitorUid)
}

// ParseInviteNotice 被邀请加群
func ParseInviteNotice(event *message.GroupInvite) *GroupInvite {
	return &GroupInvite{
		GroupUin:   event.GroupUin,
		InvitorUid: event.InvitorUid,
	}
}

func (g *GroupMemberIncrease) Preprocess(f func(uid string) uint32) {
	g.InvitorUin = f(g.InvitorUid)
	g.MemberUin = f(g.MemberUid)
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

func (g *GroupMemberDecrease) Preprocess(f func(uid string) uint32) {
	g.OperatorUin = f(g.OperatorUid)
	g.MemberUin = f(g.MemberUid)
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

func (g *GroupRecall) Preprocess(f func(uid string) uint32) {
	g.OperatorUin = f(g.OperatorUid)
	g.AuthorUin = f(g.AuthorUid)
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

func (g *GroupMute) Preprocess(f func(uid string) uint32) {
	g.OperatorUin = f(g.OperatorUid)
	g.TargetUin = f(g.TargetUid)
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
