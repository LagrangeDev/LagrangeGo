package event

import (
	"fmt"

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

	// GroupDigestEvent 群精华消息 from miraigo
	GroupDigestEvent struct {
		GroupUin          uint32
		MessageID         uint32
		InternalMessageID uint32
		OperationType     uint32 // 1 -> 设置精华消息, 2 -> 移除精华消息
		OperateTime       uint32
		SenderUin         uint32
		OperatorUin       uint32
		SenderNick        string
		OperatorNick      string
	}

	// GroupPokeEvent 群戳一戳事件 from miraigo
	GroupPokeEvent struct {
		GroupUin uint32
		Sender   uint32
		Receiver uint32
		Suffix   string
		Action   string
	}
)

type GroupInvite struct {
	GroupUin   uint32
	InvitorUid string
	InvitorUin uint32
}

// CanPreprocess 实现预处理接口，对事件的uid进行转换等操作
type CanPreprocess interface {
	ResolveUin(func(uid string) uint32)
}

func (g *GroupMute) MuteAll() bool {
	return g.OperatorUid == ""
}

func (g *GroupMemberDecrease) IsKicked() bool {
	return g.ExitType == 131 || g.ExitType == 3
}

func (g *GroupDigestEvent) IsSet() bool {
	return g.OperationType == 1
}

func (g *GroupMemberJoinRequest) ResolveUin(f func(uid string) uint32) {
	g.InvitorUin = f(g.InvitorUid)
	g.TargetUin = f(g.TargetUid)
}

// ParseRequestJoinNotice 成员主动加群
func ParseRequestJoinNotice(event *message.GroupJoin) *GroupMemberJoinRequest {
	return &GroupMemberJoinRequest{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
		},
		TargetUid: event.TargetUid,
		Answer:    event.RequestField.Unwrap(),
	}
}

// ParseRequestInvitationNotice 成员被邀请加群
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

func (g *GroupInvite) ResolveUin(f func(uid string) uint32) {
	g.InvitorUin = f(g.InvitorUid)
}

// ParseInviteNotice 被邀请加群
func ParseInviteNotice(event *message.GroupInvite) *GroupInvite {
	return &GroupInvite{
		GroupUin:   event.GroupUin,
		InvitorUid: event.InvitorUid,
	}
}

func (g *GroupMemberIncrease) ResolveUin(f func(uid string) uint32) {
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

func (g *GroupMemberDecrease) ResolveUin(f func(uid string) uint32) {
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

func (g *GroupRecall) ResolveUin(f func(uid string) uint32) {
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

func (g *GroupMute) ResolveUin(f func(uid string) uint32) {
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

func ParseGroupDigestEvent(event *message.NotifyMessageBody) *GroupDigestEvent {
	return &GroupDigestEvent{
		GroupUin:          event.EssenceMessage.GroupUin,
		MessageID:         event.EssenceMessage.MsgSequence,
		InternalMessageID: event.EssenceMessage.Random,
		OperationType:     event.EssenceMessage.SetFlag,
		OperateTime:       event.EssenceMessage.TimeStamp,
		SenderUin:         event.EssenceMessage.AuthorUin,
		OperatorUin:       event.EssenceMessage.OperatorUin,
		SenderNick:        event.EssenceMessage.AuthorName,
		OperatorNick:      event.EssenceMessage.OperatorName,
	}
}

func PaeseGroupPokeEvent(event *message.NotifyMessageBody, groupUin uint32) *GroupPokeEvent {
	e := ParsePokeEvent(event.GrayTipInfo)
	return &GroupPokeEvent{
		GroupUin: groupUin,
		Sender:   e.Sender,
		Receiver: e.Receiver,
		Suffix:   e.Suffix,
		Action:   e.Action,
	}
}

func (g *GroupPokeEvent) From() uint32 {
	return g.GroupUin
}

func (g *GroupPokeEvent) Content() string {
	if g.Suffix != "" {
		return fmt.Sprintf("%d%s%d的%s", g.Sender, g.Action, g.Receiver, g.Suffix)
	} else {
		return fmt.Sprintf("%d%s%d", g.Sender, g.Action, g.Receiver)
	}
}
