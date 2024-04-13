package entity

import (
	"fmt"
	"time"
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
	JoinTime    time.Time
	LastMsgTime time.Time
	Avatar      string
}

func NewGroupMember(uin uint32, uid string, permission GroupMemberPermission, GroupLevel uint32, MemberCard,
	MemberName string, JoinTime, LastMsgTime time.Time) *GroupMember {
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
