package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils/io"
)

func BuildFetchMemberReq(groupUin uint32, memberUID string) (*Packet, error) {
	body := &oidb.OidbSvcTrpcTcp0XFE7_4{
		GroupUin: groupUin,
		Field2:   3,
		Field3:   0,
		Body: &oidb.OidbSvcTrpcScp0XFE7_3Body{
			MemberName:       true,
			MemberCard:       true,
			Level:            true,
			SpecialTitle:     true,
			JoinTimestamp:    true,
			LastMsgTimestamp: true,
			ShutUpTimestamp:  true,
			Permission:       true,
		},
		Params: &oidb.OidbSvcTrpcScp0XFE7_4Params{Uid: memberUID},
	}
	return BuildOidbPacket(0xFE7, 4, body, false, false)
}

func ParseFetchMemberResp(data []byte) (*entity.GroupMember, error) {
	var resp oidb.OidbSvcTrpcTcp0XFE7_4Response
	_, err := ParseOidbPacket(data, &resp)
	if err != nil {
		return nil, err
	}
	interner := io.NewStringInterner()
	member := resp.Member
	m := &entity.GroupMember{
		User: entity.User{
			Uin:      member.Uin.Uin,
			UID:      interner.Intern(member.Uin.Uid),
			Nickname: interner.Intern(member.MemberName),
			Avatar:   interner.Intern(entity.UserAvatar(member.Uin.Uin)),
		},
		Permission:   entity.GroupMemberPermission(member.Permission),
		MemberCard:   interner.Intern(member.MemberCard.MemberCard.Unwrap()),
		SpecialTitle: interner.Intern(member.SpecialTitle.Unwrap()),
		JoinTime:     member.JoinTimestamp,
		LastMsgTime:  member.LastMsgTimestamp,
		ShutUpTime:   member.ShutUpTimestamp.Unwrap(),
	}
	if member.Level != nil {
		m.GroupLevel = member.Level.Level
	}
	return m, nil
}
