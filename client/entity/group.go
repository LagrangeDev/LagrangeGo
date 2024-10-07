package entity

import "fmt"

type (
	EventState uint32
	EventType  uint32
)

const (
	NoNeed EventState = iota
	Unprocessed
	Processed
)

const (
	// UserJoinRequest 用户申请加群
	UserJoinRequest EventType = 1
	// GroupInvited 被邀请加群
	GroupInvited EventType = 2
	// AssignedAsAdmin 被设置为管理员
	AssignedAsAdmin EventType = 3
	// Kicked 被踢出群聊
	Kicked EventType = 7
	// RemoveAdmin 被取消管理员
	RemoveAdmin EventType = 15
	// UserInvited 群员邀请其他人
	UserInvited EventType = 22
)

type (
	Group struct {
		GroupUin        uint32
		GroupName       string
		GroupOwner      uint32
		GroupCreateTime uint32
		GroupMemo       string
		GroupLevel      uint32
		MemberCount     uint32
		MaxMember       uint32
	}

	UserJoinGroupRequest struct {
		GroupUin    uint32
		InvitorUin  uint32
		InvitorUid  string
		TargetUin   uint32
		TargetUid   string
		OperatorUin uint32
		OperatorUid string
		Sequence    uint64
		State       EventState
		EventType   EventType
		Comment     string
		IsFiltered  bool
	}

	GroupInvitedRequest struct {
		GroupUin   uint32
		InvitorUin uint32
		InvitorUid string
		Sequence   uint64
		State      EventState
		EventType  EventType
		IsFiltered bool
	}

	GroupSystemMessages struct {
		InvitedRequests []*GroupInvitedRequest
		JoinRequests    []*UserJoinGroupRequest
	}
)

func (r *UserJoinGroupRequest) Checked() bool {
	return r.State != Unprocessed
}

func (r *GroupInvitedRequest) Checked() bool {
	return r.State != Unprocessed
}

func (g *Group) Avatar() string {
	return fmt.Sprintf("https://p.qlogo.cn/gh/%d/%d/0/", g.GroupUin, g.GroupUin)
}
