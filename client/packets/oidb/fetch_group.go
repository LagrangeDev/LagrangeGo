package oidb

import (
	"errors"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils"
)

func BuildFetchGroupReq(groupUin uint32, isStrange bool) (*OidbPacket, error) {
	body := oidb.OidbSvcTrpcTcp0X88D{
		AppID: 537099973,
		Config2: &oidb.OidbSvcTrpcTcp0X88D_Config2{
			GroupUin: groupUin,
			GroupInfo: &oidb.D88DGroupInfo{
				GroupOwner:            proto.Some(true),
				GroupCreateTime:       proto.Some(true),
				GroupFlag:             proto.Some(true),
				GroupFlagExt:          proto.Some(true),
				GroupMemberMaxNum:     proto.Some(true),
				GroupMemberNum:        proto.Some(true),
				GroupOption:           proto.Some(true),
				GroupClassExt:         proto.Some(true),
				GroupSpecialClass:     proto.Some(true),
				GroupLevel:            proto.Some(true),
				GroupFace:             proto.Some(true),
				GroupDefaultPage:      proto.Some(true),
				GroupRoamingTime:      proto.Some(true),
				GroupName:             proto.Some(""),
				GroupFingerMemo:       proto.Some(""),
				GroupClassText:        proto.Some(""),
				GroupUin:              proto.Some(true),
				GroupQuestion:         proto.Some(""),
				GroupAnswer:           proto.Some(""),
				GroupGrade:            proto.Some(true),
				CertificationType:     proto.Some(true),
				CertificationText:     proto.Some(""),
				GroupRichFingerMemo:   proto.Some(""),
				HeadPortraitSeq:       proto.Some(true),
				ShutupTimestamp:       proto.Some(true),
				ShutupTimestampMe:     proto.Some(true),
				CreateSourceFlag:      proto.Some(true),
				GroupTypeFlag:         proto.Some(true),
				AppPrivilegeFlag:      proto.Some(true),
				GroupSecLevel:         proto.Some(true),
				PoidInfo:              proto.Some(""),
				SubscriptionUin:       proto.Some(true),
				GroupFlagext3:         proto.Some(true),
				IsConfGroup:           proto.Some(true),
				IsModifyConfGroupFace: proto.Some(true),
				IsModifyConfGroupName: proto.Some(true),
				NoFingerOpenFlag:      proto.Some(true),
				NoCodeFingerOpenFlag:  proto.Some(true),
			},
		},
	}
	return BuildOidbPacket(0x88D, utils.Ternary[uint32](!isStrange, 0, 14), &body, false, false)
}

func ParseFetchGroupResp(data []byte) (*oidb.D88DGroupInfoResp, error) {
	resp, err := ParseTypedError[oidb.OidbSvcTrpcTcp0X88D_Response](data)
	if err != nil {
		return nil, err
	}
	if resp.Info.GroupInfo == nil {
		return nil, errors.New("group info empty")
	}
	return resp.Info.GroupInfo, nil
}
