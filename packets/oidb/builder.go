package oidb

import (
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/proto"
)

var oidbLogger = utils.GetLogger("oidb")

type OidbPacket struct {
	Cmd       string
	Data      []byte
	ExtraData []byte
}

func BuildOidbPacket(cmd, subCmd uint32, body any, isLafter, isUid bool) (*OidbPacket, error) {
	bodyData, err := proto.Marshal(body)
	if err != nil {
		return nil, err
	}
	oidbPkt := &oidb.OidbSvcTrpcTcpBase{
		Command:    cmd,
		SubCommand: subCmd,
		Body:       bodyData,
		ErrorMsg:   "",
		Lafter:     nil,
		Properties: make([]*oidb.OidbProperty, 0),
		Reserved:   int32(utils.Bool2Int(isUid)),
	}
	if isLafter {
		oidbPkt.Lafter = &oidb.OidbLafter{}
	}

	data, err := proto.Marshal(oidbPkt)
	if err != nil {
		return nil, err
	}

	return &OidbPacket{
		Cmd:  fmt.Sprintf("OidbSvcTrpcTcp.0x%02x_%d", cmd, subCmd),
		Data: data,
	}, nil
}

func ParseOidbPacket(b []byte, pkt any) (*oidb.OidbSvcTrpcTcpBase, error) {
	var oidbBaseResp oidb.OidbSvcTrpcTcpBase
	err := proto.Unmarshal(b, &oidbBaseResp)
	if err != nil {
		return nil, err
	}
	return &oidbBaseResp, proto.Unmarshal(oidbBaseResp.Body, pkt)
}
