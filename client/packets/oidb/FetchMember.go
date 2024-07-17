package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

func BuildFetchMemberReq(groupUin uint32, memberUid string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0XFE7_4{
		GroupUin: groupUin,
		Field2:   3,
		Field3:   0,
		Body: &oidb.OidbSvcTrpcScp0XFE7_3Body{
			MemberName:       true,
			MemberCard:       true,
			Level:            true,
			JoinTimestamp:    true,
			LastMsgTimestamp: true,
			Permission:       true,
		},
		Params: &oidb.OidbSvcTrpcScp0XFE7_4Params{Uid: memberUid},
	}
	return BuildOidbPacket(0xFE7, 4, body, false, false)
}

func ParseFetchMemberResp(data []byte) (*entity.GroupMember, error) {
	var resp oidb.OidbSvcTrpcTcp0XFE7_4Response
	_, err := ParseOidbPacket(data, &resp)
	if err != nil {
		return nil, err
	}
	interner := utils.NewStringInterner()
	member := resp.Member
	m := &entity.GroupMember{
		Uin:         member.Uin.Uin,
		Uid:         interner.Intern(member.Uin.Uid),
		Permission:  entity.GroupMemberPermission(member.Permission),
		MemberCard:  interner.Intern(member.MemberCard.MemberCard.Unwrap()),
		MemberName:  interner.Intern(member.MemberName),
		JoinTime:    member.JoinTimestamp,
		LastMsgTime: member.LastMsgTimestamp,
		Avatar:      interner.Intern(entity.FriendAvatar(member.Uin.Uin)),
	}
	if member.Level != nil {
		m.GroupLevel = member.Level.Level
	}
	return m, nil
}
