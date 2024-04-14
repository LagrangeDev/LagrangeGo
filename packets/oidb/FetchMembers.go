package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/entity"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
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
			JoinTimestamp:    true,
			LastMsgTimestamp: true,
			Permission:       true,
		},
		Token: proto.Some(token),
	}
	return BuildOidbPacket(0xFE7, 3, body, false, false)
}

func ParseFetchMembersResp(data []byte) ([]*entity.GroupMember, string, error) {
	var resp oidb.OidbSvcTrpcTcp0XFE7_2Response
	err := ParseOidbPacket(data, &resp)
	if err != nil {
		return nil, "", err
	}
	members := make([]*entity.GroupMember, len(resp.Members))
	for i, member := range resp.Members {
		// 由于protobuf的优化策略，默认值不会被编码进实际的二进制流中
		if member.Level == nil {
			members[i] = entity.NewGroupMember(member.Uin.Uin, member.Uin.Uid, entity.GroupMemberPermission(member.Permission),
				0, member.MemberCard.MemberCard.Unwrap(), member.MemberName,
				member.JoinTimestamp, member.LastMsgTimestamp)
			continue
		}
		members[i] = entity.NewGroupMember(member.Uin.Uin, member.Uin.Uid, entity.GroupMemberPermission(member.Permission),
			member.Level.Level, member.MemberCard.MemberCard.Unwrap(), member.MemberName,
			member.JoinTimestamp, member.LastMsgTimestamp)
	}
	return members, resp.Token.Unwrap(), nil
}
