package oidb

import (
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

func BuildFetchUserInfoReq[T ~string | ~uint32](value T) (*Packet, error) {
	keys := []uint32{101, //头像
		102, // 简介/签名
		103, // 备注
		104,
		105,   // 等级
		107,   // 业务列表
		20002, // 昵称
		20003, // 国家
		20004, 20005, 20006,
		20009, // 性别
		20011, 20016,
		20020, // 城市
		20021, // 学校
		20022, 20023, 20024,
		20026, // 注册时间
		20031, // 生日
		20037, // 年龄
		24002, 24007, 27037, 27049,
		27372, // 状态
		27394, // QID
		27406, // 自定义状态文本
		41756, 41757, 42257, 42315, 42362, 42432, 45160, 45161, 62026,
	}
	keyList := make([]*oidb.OidbSvcTrpcTcp0XFE1_2Key, len(keys))
	for i, key := range keys {
		keyList[i] = &oidb.OidbSvcTrpcTcp0XFE1_2Key{Key: key}
	}
	var body any
	if v, ok := any(value).(string); ok {
		body = &oidb.OidbSvcTrpcTcp0XFE1_2{
			Uid:    proto.Some(v),
			Field2: 0,
			Keys:   keyList,
		}
		return BuildOidbPacket(0xFE1, 2, body.(*oidb.OidbSvcTrpcTcp0XFE1_2), false, false)
	} else if v, ok := any(value).(uint32); ok {
		body = &oidb.OidbSvcTrpcTcp0XFE1_2Uin{
			Uin:    v,
			Field2: 0,
			Keys:   keyList,
		}
		return BuildOidbPacket(0xFE1, 2, body.(*oidb.OidbSvcTrpcTcp0XFE1_2Uin), false, true)
	}
	return nil, nil
}

func ParseFetchUserInfoResp(data []byte) (*entity.User, error) {
	var resp oidb.OidbSvcTrpcTcp0XFE1_2Response
	_, err := ParseOidbPacket(data, &resp)
	if err != nil {
		return nil, err
	}

	stringProperties := func() map[uint32]string {
		p := make(map[uint32]string)
		for _, property := range resp.Body.Properties.StringProperties {
			p[property.Code] = property.Value
		}
		return p
	}()
	numProperties := func() map[uint32]uint32 {
		p := make(map[uint32]uint32)
		for _, propetry := range resp.Body.Properties.NumberProperties {
			p[propetry.Number1] = propetry.Number2
		}
		return p
	}()

	var business oidb.Business
	_ = proto.Unmarshal(utils.S2B(stringProperties[107]), &business)

	var avatar oidb.Avatar
	_ = proto.Unmarshal(utils.S2B(stringProperties[101]), &avatar)

	return &entity.User{
		Uin: resp.Body.Uin,
		//UID:      resp.Body.UID,
		Avatar:       fmt.Sprintf("%s640", avatar.URL.Unwrap()),
		QID:          stringProperties[27394],
		Nickname:     stringProperties[20002],
		PersonalSign: stringProperties[102],
		Sex:          numProperties[20009],
		Age:          numProperties[20037],
		Level:        numProperties[105],
		VipLevel: func() uint32 {
			for _, b := range business.Body.Lists {
				if b.Type == 1 {
					return b.Level
				}
			}
			return 0
		}(),
	}, nil
}
