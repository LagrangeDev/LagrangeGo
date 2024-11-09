package entity

type GroupMemberPermission uint32

const (
	Member GroupMemberPermission = iota
	Owner
	Admin
)

type GroupMember struct {
	Uin          uint32
	UID          string
	Permission   GroupMemberPermission
	GroupLevel   uint32
	MemberCard   string
	MemberName   string
	SpecialTitle string
	JoinTime     uint32
	LastMsgTime  uint32
	ShutUpTime   uint32
	Avatar       string
}

func (m *GroupMember) DisplayName() string {
	if m.MemberCard != "" {
		return m.MemberCard
	}
	return m.MemberName
}
