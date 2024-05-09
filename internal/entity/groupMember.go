package entity

import (
	"fmt"
)

type GroupMemberPermission uint

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

func NewGroupMember(uin uint32, uid string, permission GroupMemberPermission, GroupLevel uint32, MemberCard,
	MemberName string, JoinTime, LastMsgTime uint32) *GroupMember {
	return &GroupMember{
		Uin:         uin,
		Uid:         uid,
		Permission:  permission,
		GroupLevel:  GroupLevel,
		MemberCard:  MemberCard,
		MemberName:  MemberName,
		JoinTime:    JoinTime,
		LastMsgTime: LastMsgTime,
		Avatar:      fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%v&s=640", uin),
	}
}
