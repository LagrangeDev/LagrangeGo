package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	hw "github.com/LagrangeDev/LagrangeGo/client/internal/highway"
	highway2 "github.com/LagrangeDev/LagrangeGo/client/packets/highway"
	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/highway"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	"github.com/RomiChan/protobuf/proto"
)

func (c *QQClient) ensureHighwayServers() error {
	if c.highwaySession.SsoAddr == nil || c.highwaySession.SigSession == nil || c.highwaySession.SessionKey == nil {
		packet, err := highway2.BuildHighWayUrlReq(c.transport.Sig.Tgt)
		if err != nil {
			return err
		}
		payload, err := c.sendUniPacketAndWait("HttpConn.0x6ff_501", packet)
		if err != nil {
			return fmt.Errorf("get highway server: %v", err)
		}
		resp, err := highway2.ParseHighWayUrlReq(payload)
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

func (c *QQClient) highwayUpload(commonId int, r io.Reader, fileSize uint64, md5 []byte, extendInfo []byte) error {
	err := c.ensureHighwayServers()
	if err != nil {
		return err
	}
	trans := &hw.Transaction{
		CommandID: uint32(commonId),
		Body:      r,
		Sum:       md5,
		Size:      fileSize,
		Ticket:    c.highwaySession.SigSession,
		LoginSig:  c.transport.Sig.Tgt,
		Ext:       extendInfo,
	}
	_, err = c.highwaySession.Upload(trans)
	if err == nil {
		return nil
	}
	// fallback to http upload
	servers := c.highwaySession.SsoAddr
	saddr := servers[rand.Intn(len(servers))]
	server := fmt.Sprintf(
		"http://%s:%d/cgi-bin/httpconn?htcmd=0x6FF0087&uin=%d",
		binary.UInt32ToIPV4Address(saddr.IP), saddr.Port, c.transport.Sig.Uin,
	)
	buffer := make([]byte, hw.BlockSize)
	for offset := uint64(0); offset < fileSize; offset += hw.BlockSize {
		if hw.BlockSize > fileSize-offset {
			buffer = buffer[:fileSize-offset]
		}
		_, err := io.ReadFull(r, buffer)
		if err != nil {
			return err
		}
		err = c.highwayUploadBlock(trans, server, offset, crypto.MD5Digest(buffer), buffer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *QQClient) highwayUploadBlock(trans *hw.Transaction, server string, offset uint64, blkmd5 []byte, blk []byte) error {
	blksz := uint64(len(blk))
	isEnd := offset+blksz == trans.Size
	payload, err := sendHighwayPacket(
		trans.Build(&c.highwaySession, offset, uint32(blksz), blkmd5), blk, server, isEnd,
	)
	if err != nil {
		return fmt.Errorf("send highway packet: %v", err)
	}
	defer payload.Close()
	resphead, respbody, err := parseHighwayPacket(payload)
	if err != nil {
		return fmt.Errorf("parse highway packet: %v", err)
	}
	networkLogger.Debugf("Highway Block Result: %d | %d | %x | %v",
		resphead.ErrorCode, resphead.MsgSegHead.RetCode.Unwrap(), resphead.BytesRspExtendInfo, respbody)
	if resphead.ErrorCode != 0 {
		return errors.New("highway error code: " + strconv.Itoa(int(resphead.ErrorCode)))
	}
	return nil
}

func parseHighwayPacket(data io.Reader) (head *highway.RespDataHighwayHead, body *binary.Reader, err error) {
	reader := binary.ParseReader(data)
	if reader.ReadBytesNoCopy(1)[0] != 0x28 {
		return nil, nil, errors.New("invalid highway packet")
	}
	headlength := reader.ReadU32()
	_ = reader.ReadU32() // body len
	head = &highway.RespDataHighwayHead{}
	headraw := reader.ReadBytesNoCopy(int(int64(headlength)))
	err = proto.Unmarshal(headraw, head)
	if err != nil {
		return nil, nil, err
	}
	if reader.ReadBytesNoCopy(1)[0] != 0x29 {
		return nil, nil, errors.New("invalid highway head")
	}
	return head, reader, nil
}

func sendHighwayPacket(packet *highway.ReqDataHighwayHead, block []byte, serverURL string, end bool) (io.ReadCloser, error) {
	marshal, err := proto.Marshal(packet)
	if err != nil {
		return nil, err
	}
	buf := hw.Frame(marshal, block)
	data, err := io.ReadAll(&buf)
	if err != nil {
		return nil, err
	}
	return postHighwayContent(bytes.NewReader(data), serverURL, end)
	/*
		return postHighwayContent(
			binary.NewBuilder(nil).
				WriteBytes([]byte{0x28}).
				WriteU32(uint32(len(marshal))).
				WriteU32(uint32(len(block))).
				WriteBytes(marshal).
				WriteBytes(block).
				WriteBytes([]byte{0x29}).ToReader(), serverURL, end)
	*/
}

func postHighwayContent(content io.Reader, serverURL string, end bool) (io.ReadCloser, error) {
	// Parse server URL
	server, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	// Create request
	networkLogger.Debugln("post content to highway url:", server)
	req, err := http.NewRequest("POST", server.String(), content)
	if err != nil {
		return nil, err
	}

	// Set headers
	if end {
		req.Header.Set("Connection", "close")
	} else {
		req.Header.Set("Connection", "keep-alive")
	}

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
