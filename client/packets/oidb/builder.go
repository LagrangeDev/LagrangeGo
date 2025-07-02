package oidb

import (
	"errors"
	"fmt"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/oidb"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils/io"
)

// var oidbLogger = io.GetLogger("oidb")

type Packet struct {
	Cmd       string
	Data      []byte
	ExtraData []byte
}

func BuildOidbPacket(cmd, subCmd uint32, body any, isLafter, isUID bool) (*Packet, error) {
	bodyData, err := proto.Marshal(body)
	if err != nil {
		return nil, err
	}
	oidbPkt := &oidb.OidbSvcTrpcTcpBase{
		Command:    cmd,
		SubCommand: subCmd,
		Body:       bodyData,
		Reserved:   int32(io.Bool2Int(isUID)),
	}
	if isLafter {
		oidbPkt.Lafter = &oidb.OidbLafter{}
	}

	data, err := proto.Marshal(oidbPkt)
	if err != nil {
		return nil, err
	}

	return &Packet{
		Cmd:  fmt.Sprintf("OidbSvcTrpcTcp.0x%02x_%d", cmd, subCmd),
		Data: data,
	}, nil
}

func ParseOidbPacket(b []byte, pkt any) (oidbBaseResp oidb.OidbSvcTrpcTcpBase, err error) {
	err = proto.Unmarshal(b, &oidbBaseResp)
	if err != nil {
		return
	}
	if oidbBaseResp.ErrorCode != 0 {
		return oidbBaseResp, errors.New(oidbBaseResp.ErrorMsg)
	}
	if pkt == nil {
		return
	}
	err = proto.Unmarshal(oidbBaseResp.Body, pkt)
	return
}

func CheckError(data []byte) error {
	baseResp, err := ParseOidbPacket(data, nil)
	if err != nil {
		return err
	}
	if baseResp.ErrorCode != 0 {
		return errors.New(baseResp.ErrorMsg)
	}
	return nil
}

func CheckTypedError[T any](data []byte) error {
	var resp T
	baseResp, err := ParseOidbPacket(data, &resp)
	if err != nil {
		return err
	}
	if baseResp.ErrorCode != 0 {
		return errors.New(baseResp.ErrorMsg)
	}
	return nil
}

func ParseTypedError[T any](data []byte) (*T, error) {
	var resp T
	baseResp, err := ParseOidbPacket(data, &resp)
	if err != nil {
		return nil, err
	}
	if baseResp.ErrorCode != 0 {
		return nil, errors.New(baseResp.ErrorMsg)
	}
	return &resp, nil
}
