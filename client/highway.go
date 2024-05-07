package client

import (
	"bytes"
	binary2 "encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	highway2 "github.com/LagrangeDev/LagrangeGo/packets/highway"
	"github.com/LagrangeDev/LagrangeGo/packets/pb/service/highway"
	"github.com/LagrangeDev/LagrangeGo/utils"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
	"github.com/RomiChan/protobuf/proto"
)

type UpBlock struct {
	CommandId  int
	Uin        uint
	Sequence   uint
	FileSize   uint64
	Offset     uint64
	Ticket     []byte
	FileMd5    []byte
	Block      []byte
	ExtendInfo []byte
	Timestamp  uint64
}

func (c *QQClient) GetServiceServer() ([]byte, map[uint32][]string) {
	if c.highwayUri == nil || c.sigSession == nil {
		c.highwayUri = make(map[uint32][]string)
		packet, err := highway2.BuildHighWayUrlReq(c.sig.Tgt)
		if err != nil {
			return nil, nil
		}
		payload, err := c.SendUniPacketAndAwait("HttpConn.0x6ff_501", packet)
		if err != nil {
			networkLogger.Errorf("Failed to get highway server: %v", err)
			return nil, nil
		}
		resp, err := highway2.ParseHighWayUrlReq(payload.Data)
		if err != nil {
			networkLogger.Errorf("Failed to parse highway server: %v", err)
			return nil, nil
		}
		for _, info := range resp.HttpConn.ServerInfos {
			servicetype := info.ServiceType
			for _, addr := range info.ServerAddrs {
				ip := make([]byte, 4)
				binary2.LittleEndian.PutUint32(ip, addr.IP)
				service := c.highwayUri[servicetype]
				service = append(service, fmt.Sprintf("http://%d.%d.%d.%d:%d/cgi-bin/httpconn?htcmd=0x6FF0087&uin=%d", ip[0], ip[1], ip[2], ip[3], addr.Port, c.sig.Uin))
				c.highwayUri[servicetype] = service
			}
		}
		c.sigSession = resp.HttpConn.SigSession
	}
	return c.sigSession, c.highwayUri
}

func (c *QQClient) UploadSrcByStreamAsync(commonId int, stream io.ReadSeeker, ticket []byte, md5 []byte, extendInfo []byte) bool {
	// Get server URL
	_, server := c.GetServiceServer()
	if server == nil {
		networkLogger.Errorln("Failed to get upload server")
		return false
	}
	success := true
	var upBlocks []UpBlock
	data, err := io.ReadAll(stream)
	if err != nil {
		networkLogger.Errorln("Failed to read stream")
		return false
	}

	fileSize := uint64(len(data))
	offset := uint64(0)
	_, err = stream.Seek(0, io.SeekStart)
	if err != nil {
		networkLogger.Errorln("Failed to seek stream")
		return false
	}

	for offset < fileSize {
		var buffersize uint64
		if uint64(1024*1024) > fileSize-offset {
			buffersize = fileSize - offset
		} else {
			buffersize = uint64(1024 * 1024)
		}
		buffer := make([]byte, buffersize)
		payload, err := io.ReadFull(stream, buffer)
		if err != nil {
			networkLogger.Errorln("Failed to read stream")
			return false
		}
		reqBody := UpBlock{
			CommandId:  commonId,
			Uin:        uint(c.sig.Uin),
			Sequence:   uint(c.highwaySequence.Add(1)),
			FileSize:   fileSize,
			Offset:     offset,
			Ticket:     ticket,
			FileMd5:    md5,
			Block:      buffer,
			ExtendInfo: extendInfo,
		}
		upBlocks = append(upBlocks, reqBody)
		offset += uint64(payload)
		// 4 is HighwayConcurrent
		if len(upBlocks) >= 4 || offset == fileSize {
			for _, block := range upBlocks {
				success = success && c.SendUpBlockAsync(block, server[1][0])
				if !success {
					networkLogger.Errorln("Failed to send block")
					return false
				}
			}
			upBlocks = nil
		}
	}
	return success
}

func (c *QQClient) SendUpBlockAsync(block UpBlock, server string) bool {
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
	md5 := utils.MD5Digest(block.Block)
	segHead := &highway.SegHead{
		ServiceId:     proto.Some(uint32(0)),
		Filesize:      block.FileSize,
		DataOffset:    proto.Some(block.Offset),
		DataLength:    uint32(len(block.Block)),
		RetCode:       proto.Some(uint32(0)),
		ServiceTicket: block.Ticket,
		Md5:           md5,
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
	isEnd := block.Offset+uint64(len(block.Block)) == block.FileSize
	packet := &binary.Builder{}
	packet.WriteBytes(block.Block, false)
	payload, err := SendPacketAsync(highwayHead, packet, server, isEnd)
	if err != nil {
		networkLogger.Errorln("Failed to send packet ", err)
		return false
	}
	resphead, respbody, err := ParsePacket(payload)
	if err != nil {
		networkLogger.Errorln("Failed to parse packet ", err)
		return false
	}
	networkLogger.Debugf("Highway Block Result: %d | %d | %x | %v",
		resphead.ErrorCode, resphead.MsgSegHead.RetCode.Unwrap(), resphead.BytesRspExtendInfo, respbody)
	return resphead.ErrorCode == 0
}

func ParsePacket(data []byte) (head *highway.RespDataHighwayHead, body *binary.Reader, err error) {
	reader := binary.NewReader(data)
	if reader.ReadBytesNoCopy(1)[0] == 0x28 {
		headlength := reader.ReadU32()
		bodylength := reader.ReadU32()
		head = &highway.RespDataHighwayHead{}
		headraw := reader.ReadBytesNoCopy(int(int64(headlength)))
		err = proto.Unmarshal(headraw, head)
		if err != nil {
			return nil, nil, err
		}
		body = binary.NewReader(reader.ReadBytesNoCopy(int(bodylength)))
		if reader.ReadBytesNoCopy(1)[0] == 0x29 {
			return head, body, nil
		}
	}
	return nil, nil, err
}

func SendPacketAsync(packet *highway.ReqDataHighwayHead, buffer *binary.Builder, serverURL string, end bool) ([]byte, error) {
	marshal, err := proto.Marshal(packet)
	if err != nil {
		return nil, err
	}

	println(hex.EncodeToString(marshal))

	writer := binary.NewBuilder(nil).
		WriteBytes([]byte{0x28}, false).
		WriteU32(uint32(len(marshal))).
		WriteU32(uint32(buffer.Len())).
		WriteBytes(marshal, false).
		WriteBytes(buffer.ToBytes(), false).
		WriteBytes([]byte{0x29}, false)

	return SendDataAsync(writer.ToBytes(), serverURL, end)
}

func SendDataAsync(packet []byte, serverURL string, end bool) ([]byte, error) {
	// Parse server URL
	server, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	// Create request
	content := bytes.NewBuffer(packet)
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
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Read response data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
