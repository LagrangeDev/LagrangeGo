package entity

type GroupMemberPermission uint32

const (
	Member GroupMemberPermission = iota
	Owner
	Admin
)

type GroupMember struct {
	User
	Permission   GroupMemberPermission
	GroupLevel   uint32
	MemberCard   string
	SpecialTitle string
	JoinTime     uint32
	LastMsgTime  uint32
	ShutUpTime   uint32
}

func (m *GroupMember) DisplayName() string {
	if m.MemberCard != "" {
		return m.MemberCard
	}
	return m.Nickname
}
