package event

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/message"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

type (
	GroupEvent struct {
		GroupUin uint32
	}

	GroupMemberPermissionChanged struct {
		GroupEvent
		TargetUin uint32
		TargetUID string
		IsAdmin   bool
	}

	GroupNameUpdated struct {
		GroupUin    uint32
		NewName     string
		OperatorUin uint32
		OperatorUID string
	}

	GroupMute struct {
		GroupEvent
		OperatorUID string
		OperatorUin uint32
		TargetUID   string // when TargetUID is empty, mute all members
		TargetUin   uint32
		Duration    uint32 // Duration == math.MaxUint32 when means mute all
	}

	GroupRecall struct {
		GroupEvent
		AuthorUID   string
		AuthorUin   uint32
		OperatorUID string
		OperatorUin uint32
		Sequence    uint64
		Time        uint32
		Random      uint32
	}

	GroupMemberJoinRequest struct {
		GroupEvent
		TargetUID  string
		TargetUin  uint32
		TargetNick string
		InvitorUID string
		InvitorUin uint32
		Answer     string // 问题：(.*)答案：(.*)
		RequestSeq uint64
	}

	GroupMemberIncrease struct {
		GroupEvent
		MemberUID  string
		MemberUin  uint32
		InvitorUID string
		InvitorUin uint32
		JoinType   uint32
	}

	GroupMemberDecrease struct {
		GroupEvent
		MemberUID   string
		MemberUin   uint32
		OperatorUID string
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

	GroupReactionEvent struct {
		GroupUin    uint32
		TargetSeq   uint32
		OperatorUID string
		OperatorUin uint32
		IsAdd       bool
		Code        string
		Count       uint32
	}

	// MemberSpecialTitleUpdated 群成员头衔更新事件 from miraigo
	MemberSpecialTitleUpdated struct {
		GroupUin uint32
		Uin      uint32
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

// CanPreprocess 实现预处理接口，对事件的uid进行转换等操作
type CanPreprocess interface {
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
	g.TargetUin = f(g.TargetUID, g.GroupUin)
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
		},
		TargetUID: uid,
		IsAdmin:   isAdmin,
	}
}

func (g *GroupNameUpdated) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.OperatorUin = f(g.OperatorUID, g.GroupUin)
}

func ParseGroupNameUpdatedEvent(event *message.NotifyMessageBody, groupName string) *GroupNameUpdated {
	return &GroupNameUpdated{
		GroupUin:    event.GroupUin,
		NewName:     groupName,
		OperatorUID: event.OperatorUid,
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
		},
		TargetUID: event.TargetUid,
		Answer:    event.RequestField.Unwrap(),
	}
}

// ParseRequestInvitationNotice 成员被邀请加群
func ParseRequestInvitationNotice(event *message.GroupInvitation) *GroupMemberJoinRequest {
	inn := event.Info.Inner
	return &GroupMemberJoinRequest{
		GroupEvent: GroupEvent{
			GroupUin: inn.GroupUin,
		},
		TargetUID:  inn.TargetUid,
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
	g.MemberUin = f(g.MemberUID, g.GroupUin)
}

func ParseMemberIncreaseEvent(event *message.GroupChange) *GroupMemberIncrease {
	return &GroupMemberIncrease{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
		},
		MemberUID:  event.MemberUid,
		InvitorUID: utils.B2S(event.Operator),
		JoinType:   event.IncreaseType,
	}
}

func (g *GroupMemberDecrease) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.OperatorUin = f(g.OperatorUID, g.GroupUin)
	g.MemberUin = f(g.MemberUID, g.GroupUin)
}

func ParseMemberDecreaseEvent(event *message.GroupChange) *GroupMemberDecrease {
	return &GroupMemberDecrease{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
		},
		MemberUID:   event.MemberUid,
		OperatorUID: utils.B2S(event.Operator),
		ExitType:    event.DecreaseType,
	}
}

func (g *GroupRecall) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.OperatorUin = f(g.OperatorUID, g.GroupUin)
	g.AuthorUin = f(g.AuthorUID, g.GroupUin)
}

func ParseGroupRecallEvent(event *message.NotifyMessageBody) *GroupRecall {
	info := event.Recall.RecallMessages[0]
	result := GroupRecall{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
		},
		AuthorUID:   info.AuthorUid,
		OperatorUID: event.Recall.OperatorUid.Unwrap(),
		Sequence:    info.Sequence,
		Time:        info.Time,
		Random:      info.Random,
	}
	return &result
}

func (g *GroupMute) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.OperatorUin = f(g.OperatorUID, g.GroupUin)
	g.TargetUin = f(g.TargetUID, g.GroupUin)
}

func ParseGroupMuteEvent(event *message.GroupMute) *GroupMute {
	return &GroupMute{
		GroupEvent: GroupEvent{
			GroupUin: event.GroupUin,
		},
		OperatorUID: event.OperatorUid.Unwrap(),
		TargetUID:   event.Data.State.TargetUid.Unwrap(),
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

func ParseGroupPokeEvent(event *message.NotifyMessageBody, groupUin uint32) *GroupPokeEvent {
	e := ParsePokeEvent(event.GrayTipInfo)
	return &GroupPokeEvent{
		GroupUin: groupUin,
		Sender:   e.Sender,
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
		GroupUin: groupUin,
		Uin:      event.TargetUin,
		NewTitle: medalData.Text,
	}
}

func (g *GroupReactionEvent) ResolveUin(f func(uid string, groupUin ...uint32) uint32) {
	g.OperatorUin = f(g.OperatorUID, g.GroupUin)
}

func ParseGroupReactionEvent(event *message.GroupReaction) *GroupReactionEvent {
	return &GroupReactionEvent{
		GroupUin:    event.GroupUid,
		TargetSeq:   event.Data.Data.Data.Target.Sequence,
		OperatorUID: event.Data.Data.Data.Data.OperatorUid,
		IsAdd:       event.Data.Data.Data.Data.Type == 1,
		Code:        event.Data.Data.Data.Data.Code,
		Count:       event.Data.Data.Data.Data.Count,
	}
}

func (g *GroupPokeEvent) From() uint32 {
	return g.GroupUin
}

func (g *GroupPokeEvent) Content() string {
	if g.Suffix != "" {
		return fmt.Sprintf("%d%s%d的%s", g.Sender, g.Action, g.Receiver, g.Suffix)
	}
	return fmt.Sprintf("%d%s%d", g.Sender, g.Action, g.Receiver)
}
