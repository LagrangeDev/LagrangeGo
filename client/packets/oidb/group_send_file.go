package oidb

import (
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	"github.com/RomiChan/protobuf/proto"
)

func BuildGroupSendFileReq(groupUin uint32, fileKey string) (*OidbPacket, error) {
	body := &oidb.OidbSvcTrpcTcp0X6D9_4{
		Body: &oidb.OidbSvcTrpcTcp0X6D9_4Body{
			GroupUin: groupUin,
			Type:     4,
			Info: &oidb.OidbSvcTrpcTcp0X6D9_4Info{
				BusiType: 102,
				FileId:   fileKey,
				Field3:   crypto.RandU32(),
				Field4:   proto.String("{}"),
				Field5:   true,
			},
		},
	}
	return BuildOidbPacket(0x6D9, 4, body, false, true)
}

func ParseGroupSendFileResp(data []byte) (*oidb.OidbSvcTrpcTcpBase, error) {
	return ParseTypedError[oidb.OidbSvcTrpcTcpBase](data)
}
