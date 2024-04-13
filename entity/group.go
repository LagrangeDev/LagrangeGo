package entity

import (
	"fmt"
)

type Group struct {
	GroupUin    uint32
	GroupName   string
	MemberCount uint32
	MaxMember   uint32
	Avatar      string
}

func NewGroup(groupUin uint32, groupName string, memberCount, maxMember uint32) *Group {
	return &Group{
		GroupUin:    groupUin,
		GroupName:   groupName,
		MemberCount: memberCount,
		MaxMember:   maxMember,
		Avatar:      fmt.Sprintf("https://p.qlogo.cn/gh/%v/%v/0/", groupUin, groupUin),
	}
}
