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

	highway2 "github.com/LagrangeDev/LagrangeGo/pkg/highway"
	"github.com/LagrangeDev/LagrangeGo/pkg/pb/service/highway"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/LagrangeDev/LagrangeGo/utils/crypto"
	"github.com/RomiChan/protobuf/proto"
)

const (
	uploadBlockSize        = 1024 * 1024
	httpServiceType uint32 = 1
)

type UpBlock struct {
	CommandId  int
	Uin        uint
	Sequence   uint
	FileSize   uint64
	Offset     uint64
	Ticket     []byte
	FileMd5    []byte
	Block      io.Reader
	BlockMd5   []byte
	BlockSize  uint32
	ExtendInfo []byte
	Timestamp  uint64
}

func (c *QQClient) EnsureHighwayServers() error {
	if c.highwayUri == nil || c.sigSession == nil {
		c.highwayUri = make(map[uint32][]string)
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
		for _, info := range resp.HttpConn.ServerInfos {
			servicetype := info.ServiceType
			for _, addr := range info.ServerAddrs {
				service := c.highwayUri[servicetype]
				service = append(service, fmt.Sprintf(
					"http://%s:%d/cgi-bin/httpconn?htcmd=0x6FF0087&uin=%d",
					le32toipstr(addr.IP), addr.Port, c.sig.Uin,
				))
				c.highwayUri[servicetype] = service
			}
		}
		c.sigSession = resp.HttpConn.SigSession
	}
	if c.highwayUri == nil || c.sigSession == nil {
		return errors.New("empty highway servers")
	}
	return nil
}

func (c *QQClient) UploadSrcByStream(commonId int, r io.Reader, fileSize uint64, md5 []byte, extendInfo []byte) error {
	err := c.EnsureHighwayServers()
	if err != nil {
		return err
	}
	servers := c.highwayUri[httpServiceType]
	server := servers[rand.Intn(len(servers))]
	buffer := make([]byte, uploadBlockSize)
	for offset := uint64(0); offset < fileSize; offset += uploadBlockSize {
		if uploadBlockSize > fileSize-offset {
			buffer = buffer[:fileSize-offset]
		}
		_, err := io.ReadFull(r, buffer)
		if err != nil {
			return err
		}
		err = c.SendUpBlock(&UpBlock{
			CommandId:  commonId,
			Uin:        uint(c.sig.Uin),
			Sequence:   uint(c.highwaySequence.Add(1)),
			FileSize:   fileSize,
			Offset:     offset,
			Ticket:     c.sigSession,
			FileMd5:    md5,
			Block:      bytes.NewReader(buffer),
			BlockMd5:   crypto.MD5Digest(buffer),
			BlockSize:  uint32(len(buffer)),
			ExtendInfo: extendInfo,
		}, server)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *QQClient) SendUpBlock(block *UpBlock, server string) error {
	head := &highway.DataHighwayHead{
		Version:    1,
		Uin:        proto.Some(strconv.Itoa(int(block.Uin))),
		Command:    proto.Some("PicUp.DataUp"),
		Seq:        proto.Some(uint32(block.Sequence)),
		RetryTimes: proto.Some(uint32(0)),
		AppId:      uint32(c.appInfo.SubAppID),
		DataFlag:   16,
		CommandId:  uint32(block.CommandId),
	}
	segHead := &highway.SegHead{
		ServiceId:     proto.Some(uint32(0)),
		Filesize:      block.FileSize,
		DataOffset:    proto.Some(block.Offset),
		DataLength:    uint32(block.BlockSize),
		RetCode:       proto.Some(uint32(0)),
		ServiceTicket: block.Ticket,
		Md5:           block.BlockMd5,
		FileMd5:       block.FileMd5,
		CacheAddr:     proto.Some(uint32(0)),
		CachePort:     proto.Some(uint32(0)),
	}
	loginHead := &highway.LoginSigHead{
		Uint32LoginSigType: 8,
		BytesLoginSig:      c.sig.Tgt,
		AppId:              uint32(c.appInfo.AppID),
	}
	highwayHead := &highway.ReqDataHighwayHead{
		MsgBaseHead:        head,
		MsgSegHead:         segHead,
		BytesReqExtendInfo: block.ExtendInfo,
		Timestamp:          block.Timestamp,
		MsgLoginSigHead:    loginHead,
	}
	isEnd := block.Offset+uint64(block.BlockSize) == block.FileSize
	payload, err := sendHighwayPacket(highwayHead, block.Block, block.BlockSize, server, isEnd)
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

func sendHighwayPacket(packet *highway.ReqDataHighwayHead, buffer io.Reader, bufferSize uint32, serverURL string, end bool) (io.ReadCloser, error) {
	marshal, err := proto.Marshal(packet)
	if err != nil {
		return nil, err
	}

	writer := binary.NewBuilder(nil).
		WriteBytes([]byte{0x28}).
		WriteU32(uint32(len(marshal))).
		WriteU32(bufferSize).
		WriteBytes(marshal)
	_, _ = io.Copy(writer, buffer)
	_, _ = writer.Write([]byte{0x29})

	return postHighwayContent(writer.ToReader(), serverURL, end)
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
