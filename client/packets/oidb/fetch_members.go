package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

func BuildFetchMembersReq(groupUin uint32, token string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0XFE7_3{
		GroupUin: groupUin,
		Field2:   5,
		Field3:   2,
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
		Token: proto.Some(token),
	}
	return BuildOidbPacket(0xFE7, 3, body, false, false)
}

func ParseFetchMembersResp(data []byte) ([]*entity.GroupMember, string, error) {
	var resp oidb.OidbSvcTrpcTcp0XFE7_2Response
	_, err := ParseOidbPacket(data, &resp)
	if err != nil {
		return nil, "", err
	}
	interner := utils.NewStringInterner()
	members := make([]*entity.GroupMember, len(resp.Members))
	for i, member := range resp.Members {
		// 由于protobuf的优化策略，默认值不会被编码进实际的二进制流中
		m := &entity.GroupMember{
			Uin:          member.Uin.Uin,
			UID:          interner.Intern(member.Uin.Uid),
			Permission:   entity.GroupMemberPermission(member.Permission),
			MemberCard:   interner.Intern(member.MemberCard.MemberCard.Unwrap()),
			MemberName:   interner.Intern(member.MemberName),
			SpecialTitle: interner.Intern(member.SpecialTitle.Unwrap()),
			JoinTime:     member.JoinTimestamp,
			LastMsgTime:  member.LastMsgTimestamp,
			ShutUpTime:   member.ShutUpTimestamp.Unwrap(),
			Avatar:       interner.Intern(entity.FriendAvatar(member.Uin.Uin)),
		}
		if member.Level != nil {
			m.GroupLevel = member.Level.Level
		}
		members[i] = m
	}
	return members, resp.Token.Unwrap(), nil
}
