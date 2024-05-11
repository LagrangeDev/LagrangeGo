package entity

type GroupMemberPermission uint32

const (
	Member GroupMemberPermission = iota
	Owner
	Admin
)

type GroupMember struct {
	Uin         uint32
	Uid         string
	Permission  GroupMemberPermission
	GroupLevel  uint32
	MemberCard  string
	MemberName  string
	JoinTime    uint32
	LastMsgTime uint32
	Avatar      string
}
