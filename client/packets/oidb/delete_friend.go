package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
)

func BuildDeleteFriendReq(uid string, block bool) (*Packet, error) {
	body := oidb.OidbSvcTrpcTcp0X126B_0{
		Field1: &oidb.OidbSvcTrpcTcp0X126B_0_Field1{
			TargetUid: uid,
			Field2: &oidb.OidbSvcTrpcTcp0X126B_0_Field1_2{
				Field1: 130,
				Field2: 109,
				Field3: &oidb.OidbSvcTrpcTcp0X126B_0_Field1_2_3{
					Field1: 8,
					Field2: 8,
					Field3: 50,
				},
			},
			Block:  block,
			Field4: true,
		},
	}
	return BuildOidbPacket(0x126B, 0, &body, false, false)
}

func ParseDeleteFriendResp(data []byte) error {
	return CheckError(data)
}
