package client

import (
	"errors"
	"fmt"
	"io"

	"github.com/LagrangeDev/LagrangeGo/client/highway"
	highway2 "github.com/LagrangeDev/LagrangeGo/packets/highway"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

func (c *QQClient) EnsureHighwayServers() error {
	if c.highwaySession.SsoAddr == nil || c.highwaySession.SigSession == nil || c.highwaySession.SessionKey == nil {
		packet, err := highway2.BuildHighWayUrlReq(c.sig.Tgt)
		if err != nil {
			return err
		}
		payload, err := c.SendUniPacketAndAwait("HttpConn.0x6ff_501", packet)
		if err != nil {
			return fmt.Errorf("get highway server: %v", err)
		}
		resp, err := highway2.ParseHighWayUrlReq(payload.Data)
		if err != nil {
			return fmt.Errorf("parse highway server: %v", err)
		}
		c.highwaySession.SigSession = resp.HttpConn.SigSession
		c.highwaySession.SessionKey = resp.HttpConn.SessionKey
		for _, info := range resp.HttpConn.ServerInfos {
			if info.ServiceType != 1 {
				continue
			}
			for _, addr := range info.ServerAddrs {
				networkLogger.Debugln("add highway server", binary.UInt32ToIPV4Address(addr.IP), "port", addr.Port)
				c.highwaySession.AppendAddr(addr.IP, addr.Port)
			}
		}
	}
	if c.highwaySession.SsoAddr == nil || c.highwaySession.SigSession == nil || c.highwaySession.SessionKey == nil {
		return errors.New("empty highway servers")
	}
	return nil
}

func (c *QQClient) UploadSrcByStream(commandId int, r io.Reader, fileSize uint64, md5 []byte, extendInfo []byte) error {
	err := c.EnsureHighwayServers()
	if err != nil {
		return err
	}
	networkLogger.Debugln("highway upload cmd:", commandId, "size:", fileSize)
	data, err := c.highwaySession.Upload(highway.Transaction{
		CommandID: uint32(commandId),
		Body:      r,
		Sum:       md5,
		Size:      fileSize,
		Ticket:    c.highwaySession.SigSession,
		LoginSig:  c.sig.Tgt,
		Ext:       extendInfo,
	})
	networkLogger.Debugln("highway upload resp ext:", utils.B2S(data))
	return err
}
