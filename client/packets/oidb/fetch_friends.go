package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils/io"
)

// BuildFetchFriendsReq OidbSvcTrpcTcp.0xfd4_1
func BuildFetchFriendsReq(token uint32) (*Packet, error) {
	body := oidb.OidbSvcTrpcTcp0XFD4_1{
		Field2: 300,
		Field4: 0,
		NextUin: &oidb.OidbSvcTrpcTcp0XFD4_1Uin{
			Uin: token,
		},
		Field6: 1,
		Body: []*oidb.OidbSvcTrpcTcp0XFD4_1Body{{
			Type:   1,
			Number: &oidb.OidbNumber{Numbers: []uint32{103, 102, 20002, 27394}},
		}, {
			Type:   4,
			Number: &oidb.OidbNumber{Numbers: []uint32{100, 101, 102}},
		}},
		Field10002: []uint32{13578, 13579, 13573, 13572, 13568},
		Field10003: 4051,
	}
	/*
	 * OidbNumber里面的东西代表你想要拿到的Property，这些Property将会在返回的数据里面的Preserve的Field，
	 * 102：个性签名
	 * 103：备注
	 * 20002：昵称
	 */
	return BuildOidbPacket(0xFD4, 1, &body, false, false)
}

func ParseFetchFriendsResp(data []byte) ([]*entity.User, uint32, error) {
	var resp oidb.OidbSvcTrpcTcp0XFD4_1Response
	var next uint32
	if _, err := ParseOidbPacket(data, &resp); err != nil {
		return nil, 0, err
	}
	if resp.Next == nil {
		next = 0
	} else {
		next = resp.Next.Uin
	}
	friends := make([]*entity.User, len(resp.Friends))
	interner := io.NewStringInterner()
	for i, raw := range resp.Friends {
		additional := getFirstFriendAdditionalTypeEqualTo1(raw.Additional)
		properties := parseFriendProperty(additional.Layer1.Properties)
		friends[i] = &entity.User{
			Uin:          raw.Uin,
			UID:          interner.Intern(raw.Uid),
			Nickname:     interner.Intern(properties[20002]),
			Remarks:      interner.Intern(properties[103]),
			PersonalSign: interner.Intern(properties[102]),
			Avatar:       interner.Intern(entity.UserAvatar(raw.Uin)),
		}
	}
	return friends, next, nil
}

func getFirstFriendAdditionalTypeEqualTo1(additionals []*oidb.OidbFriendAdditional) *oidb.OidbFriendAdditional {
	for _, additional := range additionals {
		if additional.Type == 1 {
			return additional
		}
	}
	return nil
}

func parseFriendProperty(properties []*oidb.OidbFriendProperty) map[uint32]string {
	dict := make(map[uint32]string, len(properties))
	for _, property := range properties {
		dict[property.Code] = property.Value
	}
	return dict
}
