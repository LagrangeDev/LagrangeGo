package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildFetchGroupsReq() (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0XFE5_2{
		Config: &oidb.OidbSvcTrpcTcp0XFE5_2Config{
			Config1: &oidb.OidbSvcTrpcTcp0XFE5_2Config1{
				GroupOwner: true, Field2: true, MemberMax: true, MemberCount: true, GroupName: true, Field8: true,
				Field9: true, Field10: true, Field11: true, Field12: true, Field13: true, Field14: true, Field15: true,
				Field16: true, Field17: true, Field18: true, Question: true, Field20: true, Field22: true, Field23: true,
				Field24: true, Field25: true, Field26: true, Field27: true, Field28: true, Field29: true, Field30: true,
				Field31: true, Field32: true, Field5001: true, Field5002: true, Field5003: true,
			},
			Config2: &oidb.OidbSvcTrpcTcp0XFE5_2Config2{
				Field1: true, Field2: true, Field3: true, Field4: true, Field5: true, Field6: true, Field7: true,
				Field8: true,
			},
			Config3: &oidb.OidbSvcTrpcTcp0XFE5_2Config3{
				Field5: true, Field6: true,
			},
		},
	}
	return BuildOidbPacket(0xFE5, 2, body, false, true)
}

func ParseFetchGroupsResp(data []byte) ([]*entity.Group, error) {
	var resp oidb.OidbSvcTrpcTcp0XFE5_2Response
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return nil, err
	}
	groups := make([]*entity.Group, len(resp.Groups))
	for i, group := range resp.Groups {
		groups[i] = &entity.Group{
			GroupUin:        group.GroupUin,
			GroupName:       group.Info.GroupName,
			MemberCount:     group.Info.MemberCount,
			MaxMember:       group.Info.MemberMax,
			GroupCreateTime: group.Info.CreateTimeStamp,
			GroupMemo:       group.ExtInfo.GroupMemo,
		}
	}
	return groups, nil
}
