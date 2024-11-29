package event

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

type (
	// GroupEvent 通用群事件
	GroupEvent struct {
		GroupUin uint32
		// 触发事件的主体 详看各事件注释
		UserUin uint32
		UserUID string
	}

	// GroupMemberPermissionChanged 群成员权限变更
	GroupMemberPermissionChanged struct {
		GroupEvent
		IsAdmin bool
	}

	// GroupNameUpdated 群名变更
	GroupNameUpdated struct {
		GroupEvent
		NewName string
	}

	// GroupMute 群内禁言事件 user为被禁言的成员
	GroupMute struct {
		GroupEvent
		OperatorUID string // when TargetUID is empty, mute all members
		OperatorUin uint32
		Duration    uint32 // Duration == math.MaxUint32 when means mute all
	}

	// GroupRecall 群内消息撤回 user为消息发送者
	GroupRecall struct {
		GroupEvent
		OperatorUID string
		OperatorUin uint32
		Sequence    uint64
		Time        uint32
		Random      uint32
	}

	// GroupMemberJoinRequest 加群请求 user为请求加群的成员
	GroupMemberJoinRequest struct {
		GroupEvent
		TargetNick string
		InvitorUID string
		InvitorUin uint32
		Answer     string // 问题：(.*)答案：(.*)
		RequestSeq uint64
	}

	// GroupMemberIncrease 群成员增加事件 user为新加群的成员
	GroupMemberIncrease struct {
		GroupEvent
		InvitorUID string
		InvitorUin uint32
		JoinType   uint32
	}

	// GroupMemberDecrease 群成员减少事件 user为退群的成员
	GroupMemberDecrease struct {
		GroupEvent
		OperatorUID string
		OperatorUin uint32
		ExitType    uint32
	}

	// GroupDigestEvent 群精华消息 user为消息发送者
	// from miraigo
	GroupDigestEvent struct {
		GroupEvent
		MessageID         uint32
		InternalMessageID uint32
		OperationType     uint32 // 1 -> 设置精华消息, 2 -> 移除精华消息
		OperateTime       uint32
		OperatorUin       uint32
		SenderNick        string
		OperatorNick      string
	}

	// GroupPokeEvent 群戳一戳事件 user为发送者
	// from miraigo
	GroupPokeEvent struct {
		GroupEvent
		Receiver uint32
		Suffix   string
		Action   string
	}

	// GroupReactionEvent 群消息表态 user为发送表态的成员
	GroupReactionEvent struct {
		GroupEvent
		TargetSeq uint32
		IsAdd     bool
		Code      string
		Count     uint32
	}

	// MemberSpecialTitleUpdated 群成员头衔更新事件
	// from miraigo
	MemberSpecialTitleUpdated struct {
		GroupEvent
		NewTitle string
	}
)

type GroupInvite struct {
	GroupUin    uint32
	GroupName   string
	InvitorUID  string
	InvitorUin  uint32
	InvitorNick string
	RequestSeq  uint64
}

type JSONParam struct {
	Cmd  int    `json:"cmd"`
	Data string `json:"data"`
	Text string `json:"text"`
	URL  string `json:"url"`
}

// Iuid2uin uid转uin
type Iuid2uin interface {
	ResolveUin(func(uid string, groupUin ...uint32) uint32)
}

func (g *GroupMute) MuteAll() bool {
	return g.OperatorUID == ""
}

func (g *GroupMemberDecrease) IsKicked() bool {
	return g.ExitType == 131 || g.ExitType == 3
}

func (g *GroupDigestEvent) IsSet() bool {
	return g.OperationType == 1
}

func (g *GroupMemberPermissionChanged) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.UserUin = f(g.UserUID, g.GroupUin)
}

func ParseGroupMemberPermissionChanged(event *message.GroupAdmin) *GroupMemberPermissionChanged {
	var isAdmin bool
	var uid string
	if event.Body.ExtraEnable != nil {
		isAdmin = true
		uid = event.Body.ExtraEnable.AdminUid
	} else if event.Body.ExtraDisable != nil {
		isAdmin = false
		uid = event.Body.ExtraDisable.AdminUid
	}
	return &GroupMemberPermissionChanged{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
			UserUID:  uid,
		},
		IsAdmin: isAdmin,
	}
}

func (g *GroupNameUpdated) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.UserUin = f(g.UserUID, g.GroupUin)
}

func ParseGroupNameUpdatedEvent(event *message.NotifyMessageBody, groupName string) *GroupNameUpdated {
	return &GroupNameUpdated{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
			UserUID:  event.OperatorUid,
		},
		NewName: groupName,
	}
}

func (g *GroupMemberJoinRequest) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.InvitorUin = f(g.InvitorUID, g.GroupUin)
}

// ParseRequestJoinNotice 成员主动加群
func ParseRequestJoinNotice(event *message.GroupJoin) *GroupMemberJoinRequest {
	return &GroupMemberJoinRequest{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
			UserUID:  event.TargetUid,
		},
		Answer: event.RequestField.Unwrap(),
	}
}

// ParseRequestInvitationNotice 成员被邀请加群
func ParseRequestInvitationNotice(event *message.GroupInvitation) *GroupMemberJoinRequest {
	inn := event.Info.Inner
	return &GroupMemberJoinRequest{
		GroupEvent: GroupEvent{
			GroupUin: inn.GroupUin,
			UserUID:  inn.TargetUid,
		},
		InvitorUID: inn.InvitorUid,
	}
}

func (g *GroupInvite) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.InvitorUin = f(g.InvitorUID, g.GroupUin)
}

// ParseInviteNotice 被邀请加群
func ParseInviteNotice(event *message.GroupInvite) *GroupInvite {
	return &GroupInvite{
		GroupUin:   event.GroupUin,
		InvitorUID: event.InvitorUid,
	}
}

func (g *GroupMemberIncrease) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.InvitorUin = f(g.InvitorUID, g.GroupUin)
	g.UserUin = f(g.UserUID, g.GroupUin)
}

func ParseMemberIncreaseEvent(event *message.GroupChange) *GroupMemberIncrease {
	return &GroupMemberIncrease{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
			UserUID:  event.MemberUid,
		},
		InvitorUID: utils.B2S(event.Operator),
		JoinType:   event.IncreaseType,
	}
}

func (g *GroupMemberDecrease) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.OperatorUin = f(g.OperatorUID, g.GroupUin)
	g.UserUin = f(g.UserUID, g.GroupUin)
}

func ParseMemberDecreaseEvent(event *message.GroupChange) *GroupMemberDecrease {
	return &GroupMemberDecrease{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
			UserUID:  event.MemberUid,
		},
		OperatorUID: utils.B2S(event.Operator),
		ExitType:    event.DecreaseType,
	}
}

func (g *GroupRecall) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.OperatorUin = f(g.OperatorUID, g.GroupUin)
	g.UserUin = f(g.UserUID, g.GroupUin)
}

func ParseGroupRecallEvent(event *message.NotifyMessageBody) *GroupRecall {
	info := event.Recall.RecallMessages[0]
	result := GroupRecall{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
			UserUID:  info.AuthorUid,
		},
		OperatorUID: event.Recall.OperatorUid.Unwrap(),
		Sequence:    info.Sequence,
		Time:        info.Time,
		Random:      info.Random,
	}
	return &result
}

func (g *GroupMute) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.OperatorUin = f(g.OperatorUID, g.GroupUin)
	g.UserUin = f(g.UserUID, g.GroupUin)
}

func ParseGroupMuteEvent(event *message.GroupMute) *GroupMute {
	return &GroupMute{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
			UserUID:  event.Data.State.TargetUid.Unwrap(),
		},
		OperatorUID: event.OperatorUid.Unwrap(),
		Duration:    event.Data.State.Duration,
	}
}

func ParseGroupDigestEvent(event *message.NotifyMessageBody) *GroupDigestEvent {
	return &GroupDigestEvent{
		GroupEvent: GroupEvent{
			GroupUin: event.EssenceMessage.GroupUin,
			UserUin:  event.EssenceMessage.AuthorUin,
		},
		MessageID:         event.EssenceMessage.MsgSequence,
		InternalMessageID: event.EssenceMessage.Random,
		OperationType:     event.EssenceMessage.SetFlag,
		OperateTime:       event.EssenceMessage.TimeStamp,
		OperatorUin:       event.EssenceMessage.OperatorUin,
		SenderNick:        event.EssenceMessage.AuthorName,
		OperatorNick:      event.EssenceMessage.OperatorName,
	}
}

func ParseGroupPokeEvent(event *message.NotifyMessageBody, groupUin uint32) *GroupPokeEvent {
	e := ParsePokeEvent(event.GrayTipInfo)
	return &GroupPokeEvent{
		GroupEvent: GroupEvent{
			GroupUin: groupUin,
			UserUin:  e.Sender,
		},
		Receiver: e.Receiver,
		Suffix:   e.Suffix,
		Action:   e.Action,
	}
}

func ParseGroupMemberSpecialTitleUpdatedEvent(event *message.GroupSpecialTitle, groupUin uint32) *MemberSpecialTitleUpdated {
	re := regexp.MustCompile(`<({.*?})>`)
	matches := re.FindAllStringSubmatch(event.Content, -1)
	if len(matches) != 2 {
		return nil
	}
	var medalData JSONParam
	err := json.Unmarshal([]byte(matches[1][1]), &medalData)
	if err != nil {
		return nil
	}
	return &MemberSpecialTitleUpdated{
		GroupEvent: GroupEvent{
			GroupUin: groupUin,
			UserUin:  event.TargetUin,
		},
		NewTitle: medalData.Text,
	}
}

func (g *GroupReactionEvent) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.UserUin = f(g.UserUID, g.GroupUin)
}

func ParseGroupReactionEvent(event *message.GroupReaction) *GroupReactionEvent {
	return &GroupReactionEvent{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUid,
			UserUID:  event.Data.Data.Data.Data.OperatorUid,
		},
		TargetSeq: event.Data.Data.Data.Target.Sequence,
		IsAdd:     event.Data.Data.Data.Data.Type == 1,
		Code:      event.Data.Data.Data.Data.Code,
		Count:     event.Data.Data.Data.Data.Count,
	}
}

func (g *GroupPokeEvent) From() uint32 {
	return g.GroupUin
}

func (g *GroupPokeEvent) Content() string {
	if g.Suffix != "" {
		return fmt.Sprintf("%d%s%d的%s", g.UserUin, g.Action, g.Receiver, g.Suffix)
	}
	return fmt.Sprintf("%d%s%d", g.UserUin, g.Action, g.Receiver)
}
